package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/vo"
)

type CourseService struct {
	repo          *repository.CourseRepository
	userRepo      *repository.UserRepository
	materialRepo  *repository.MaterialRepository
	agentRepo     *repository.AgentRepository
	storageRoot   string
	storageQuota  int64
	defaultAgentName string
	defaultAgentPrompt string
}

func NewCourseService(repo *repository.CourseRepository, userRepo *repository.UserRepository, materialRepo *repository.MaterialRepository, agentRepo *repository.AgentRepository, storageRoot string, storageQuota int64, defaultAgentName, defaultAgentPrompt string) *CourseService {
	return &CourseService{
		repo:               repo,
		userRepo:           userRepo,
		materialRepo:       materialRepo,
		agentRepo:          agentRepo,
		storageRoot:        storageRoot,
		storageQuota:       storageQuota,
		defaultAgentName:   defaultAgentName,
		defaultAgentPrompt: defaultAgentPrompt,
	}
}

func (s *CourseService) CreateCourse(ctx context.Context, userID uint64, courseCode, courseName, courseDescription string) (*vo.CourseVO, error) {
	if strings.TrimSpace(courseCode) == "" || strings.TrimSpace(courseName) == "" {
		return nil, apperrors.ErrInvalidParameter
	}

	if _, err := s.repo.GetCourseByCode(ctx, strings.TrimSpace(courseCode)); err == nil {
		return nil, apperrors.ErrCourseExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var created *model.Course
	now := time.Now()
	err := s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		course := &model.Course{
			CourseCode:        strings.TrimSpace(courseCode),
			CourseName:        strings.TrimSpace(courseName),
			CourseDescription: strings.TrimSpace(courseDescription),
			OwnerUserID:       userID,
			Status:            "active",
		}
		if err := tx.Create(course).Error; err != nil {
			return err
		}

		member := &model.CourseMember{
			CourseID:   course.ID,
			UserID:     userID,
			Role:       "owner",
			JoinStatus: "active",
			JoinedAt:   now,
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		rootPath := filepath.Join(s.storageRoot, fmt.Sprintf("%d", course.ID))
		if err := os.MkdirAll(rootPath, 0o755); err != nil {
			return fmt.Errorf("create course storage directory: %w", err)
		}
		space := &model.CourseStorageSpace{
			CourseID:   course.ID,
			RootPath:   rootPath,
			QuotaBytes: s.storageQuota,
			UsedBytes:  0,
		}
		if err := s.materialRepo.CreateStorageSpaceTx(tx, space); err != nil {
			return err
		}
		courseAgent := &model.CourseAgent{
			CourseID:       course.ID,
			AgentName:      s.defaultAgentName,
			PromptTemplate: s.defaultAgentPrompt,
			Status:         "enabled",
			RetrievalScope: "course_all",
			CreatedBy:      userID,
		}
		if err := s.agentRepo.CreateCourseAgentTx(tx, courseAgent); err != nil {
			return err
		}
		created = course
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := toCourseVO(created, "owner")
	return &result, nil
}

func (s *CourseService) ListCourses(ctx context.Context, userID uint64) ([]vo.CourseVO, error) {
	members, err := s.repo.ListCoursesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]vo.CourseVO, 0, len(members))
	for _, member := range members {
		course, err := s.repo.GetCourseByID(ctx, member.CourseID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		if course.Status != "active" {
			continue
		}
		result = append(result, toCourseVO(course, member.Role))
	}
	return result, nil
}

func (s *CourseService) ListDiscoverableCourses(ctx context.Context, userID uint64) ([]vo.CourseVO, error) {
	courses, err := s.repo.ListActiveCourses(ctx)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.ListCoursesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	rolesByCourseID := make(map[uint64]string, len(members))
	for _, member := range members {
		rolesByCourseID[member.CourseID] = member.Role
	}

	result := make([]vo.CourseVO, 0, len(courses))
	for _, course := range courses {
		result = append(result, toCourseVO(&course, rolesByCourseID[course.ID]))
	}
	return result, nil
}

func (s *CourseService) GetCourseDetail(ctx context.Context, userID, courseID uint64) (*vo.CourseVO, error) {
	member, err := s.repo.GetMember(ctx, courseID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}
	if member.JoinStatus != "active" {
		return nil, apperrors.ErrForbidden
	}

	course, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if course.Status != "active" {
		return nil, apperrors.ErrCourseNotFound
	}
	result := toCourseVO(course, member.Role)
	return &result, nil
}

func (s *CourseService) UpdateCourse(ctx context.Context, userID, courseID uint64, courseName, courseDescription string) (*vo.CourseVO, error) {
	course, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if course.Status != "active" {
		return nil, apperrors.ErrCourseNotFound
	}
	member, err := s.repo.GetMember(ctx, courseID, userID)
	if err != nil {
		return nil, apperrors.ErrForbidden
	}
	if member.Role != "teacher" && member.Role != "owner" {
		return nil, apperrors.ErrForbidden
	}

	if strings.TrimSpace(courseName) != "" {
		course.CourseName = strings.TrimSpace(courseName)
	}
	course.CourseDescription = strings.TrimSpace(courseDescription)
	if err := s.repo.UpdateCourse(ctx, course); err != nil {
		return nil, err
	}
	result := toCourseVO(course, member.Role)
	return &result, nil
}

func (s *CourseService) DeleteCourse(ctx context.Context, userID, courseID uint64) error {
	course, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrCourseNotFound
		}
		return err
	}
	if course.Status != "active" {
		return apperrors.ErrCourseNotFound
	}
	if course.OwnerUserID != userID {
		return apperrors.ErrForbidden
	}
	course.Status = "deleted"
	return s.repo.UpdateCourse(ctx, course)
}

func (s *CourseService) JoinCourse(ctx context.Context, userID, courseID uint64) (*vo.CourseMemberVO, error) {
	course, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if course.Status != "active" {
		return nil, apperrors.ErrCourseNotFound
	}
	if course.OwnerUserID == userID {
		return nil, apperrors.ErrCourseMemberExists
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	if user.Status != "active" {
		return nil, apperrors.ErrUserDisabled
	}

	existing, err := s.repo.GetMember(ctx, courseID, userID)
	if err == nil {
		if existing.JoinStatus == "active" {
			return nil, apperrors.ErrCourseMemberExists
		}
		existing.Role = "student"
		existing.JoinStatus = "active"
		existing.JoinedAt = time.Now()
		if err := s.repo.UpdateMember(ctx, existing); err != nil {
			return nil, err
		}
		memberVO := toCourseMemberVO(existing, user.Username)
		return &memberVO, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	member := &model.CourseMember{
		CourseID:   courseID,
		UserID:     userID,
		Role:       "student",
		JoinStatus: "active",
		JoinedAt:   time.Now(),
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, err
	}
	memberVO := toCourseMemberVO(member, user.Username)
	return &memberVO, nil
}

func (s *CourseService) ListMembers(ctx context.Context, userID, courseID uint64) ([]vo.CourseMemberVO, error) {
	course, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if course.Status != "active" {
		return nil, apperrors.ErrCourseNotFound
	}

	member, err := s.repo.GetMember(ctx, courseID, userID)
	if err != nil {
		return nil, apperrors.ErrForbidden
	}
	if member.JoinStatus != "active" {
		return nil, apperrors.ErrForbidden
	}

	members, err := s.repo.ListMembersByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	result := make([]vo.CourseMemberVO, 0, len(members))
	for _, item := range members {
		user, err := s.userRepo.GetByID(ctx, item.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}

		if member.Role == "student" && item.UserID != userID && item.Role == "student" {
			continue
		}

		result = append(result, vo.CourseMemberVO{
			ID:         item.ID,
			CourseID:   item.CourseID,
			UserID:     item.UserID,
			Username:   user.Username,
			Role:       item.Role,
			JoinStatus: item.JoinStatus,
			JoinedAt:   item.JoinedAt,
		})
	}
	return result, nil
}

func (s *CourseService) AddMember(ctx context.Context, operatorUserID, courseID uint64, username, role string) (*vo.CourseMemberVO, error) {
	if strings.TrimSpace(username) == "" || !isAssignableRole(role) {
		return nil, apperrors.ErrInvalidParameter
	}

	course, operatorMember, err := s.loadCourseAndMember(ctx, operatorUserID, courseID)
	if err != nil {
		return nil, err
	}
	if operatorMember.Role != "owner" && operatorMember.Role != "teacher" {
		return nil, apperrors.ErrForbidden
	}
	if operatorMember.Role == "teacher" && role != "student" {
		return nil, apperrors.ErrForbidden
	}

	user, err := s.userRepo.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	if user.Status != "active" {
		return nil, apperrors.ErrUserDisabled
	}

	if user.ID == course.OwnerUserID {
		return nil, apperrors.ErrCourseMemberExists
	}

	existing, err := s.repo.GetMember(ctx, courseID, user.ID)
	if err == nil {
		if existing.JoinStatus == "active" {
			return nil, apperrors.ErrCourseMemberExists
		}
		existing.Role = role
		existing.JoinStatus = "active"
		existing.JoinedAt = time.Now()
		if err := s.repo.UpdateMember(ctx, existing); err != nil {
			return nil, err
		}
		memberVO := toCourseMemberVO(existing, user.Username)
		return &memberVO, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	member := &model.CourseMember{
		CourseID:   courseID,
		UserID:     user.ID,
		Role:       role,
		JoinStatus: "active",
		JoinedAt:   time.Now(),
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, err
	}
	memberVO := toCourseMemberVO(member, user.Username)
	return &memberVO, nil
}

func (s *CourseService) UpdateMemberRole(ctx context.Context, operatorUserID, courseID, memberID uint64, role string) (*vo.CourseMemberVO, error) {
	if memberID == 0 || !isAssignableRole(role) {
		return nil, apperrors.ErrInvalidParameter
	}

	course, operatorMember, err := s.loadCourseAndMember(ctx, operatorUserID, courseID)
	if err != nil {
		return nil, err
	}
	if operatorMember.Role != "owner" {
		return nil, apperrors.ErrForbidden
	}

	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseMemberNotFound
		}
		return nil, err
	}
	if member.CourseID != courseID || member.JoinStatus != "active" {
		return nil, apperrors.ErrCourseMemberNotFound
	}
	if member.UserID == course.OwnerUserID || member.Role == "owner" {
		return nil, apperrors.ErrForbidden
	}

	member.Role = role
	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, member.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}

	memberVO := toCourseMemberVO(member, user.Username)
	return &memberVO, nil
}

func (s *CourseService) RemoveMember(ctx context.Context, operatorUserID, courseID, memberID uint64) error {
	course, operatorMember, err := s.loadCourseAndMember(ctx, operatorUserID, courseID)
	if err != nil {
		return err
	}
	if operatorMember.Role != "owner" && operatorMember.Role != "teacher" {
		return apperrors.ErrForbidden
	}

	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrCourseMemberNotFound
		}
		return err
	}
	if member.CourseID != courseID || member.JoinStatus != "active" {
		return apperrors.ErrCourseMemberNotFound
	}
	if member.UserID == course.OwnerUserID || member.Role == "owner" {
		return apperrors.ErrForbidden
	}
	if operatorMember.Role == "teacher" && member.Role != "student" {
		return apperrors.ErrForbidden
	}

	member.JoinStatus = "removed"
	return s.repo.UpdateMember(ctx, member)
}

func toCourseVO(course *model.Course, myRole string) vo.CourseVO {
	return vo.CourseVO{
		ID:                course.ID,
		CourseCode:        course.CourseCode,
		CourseName:        course.CourseName,
		CourseDescription: course.CourseDescription,
		OwnerUserID:       course.OwnerUserID,
		Status:            course.Status,
		CreatedAt:         course.CreatedAt,
		UpdatedAt:         course.UpdatedAt,
		MyRole:            myRole,
	}
}

func toCourseMemberVO(member *model.CourseMember, username string) vo.CourseMemberVO {
	return vo.CourseMemberVO{
		ID:         member.ID,
		CourseID:   member.CourseID,
		UserID:     member.UserID,
		Username:   username,
		Role:       member.Role,
		JoinStatus: member.JoinStatus,
		JoinedAt:   member.JoinedAt,
	}
}

func (s *CourseService) loadCourseAndMember(ctx context.Context, userID, courseID uint64) (*model.Course, *model.CourseMember, error) {
	course, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperrors.ErrCourseNotFound
		}
		return nil, nil, err
	}
	if course.Status != "active" {
		return nil, nil, apperrors.ErrCourseNotFound
	}

	member, err := s.repo.GetMember(ctx, courseID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperrors.ErrForbidden
		}
		return nil, nil, err
	}
	if member.JoinStatus != "active" {
		return nil, nil, apperrors.ErrForbidden
	}

	return course, member, nil
}

func isAssignableRole(role string) bool {
	switch strings.TrimSpace(role) {
	case "teacher", "student":
		return true
	default:
		return false
	}
}
