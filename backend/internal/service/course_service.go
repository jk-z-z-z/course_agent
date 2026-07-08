package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/vo"
)

type CourseService struct {
	repo     *repository.CourseRepository
	userRepo *repository.UserRepository
}

func NewCourseService(repo *repository.CourseRepository, userRepo *repository.UserRepository) *CourseService {
	return &CourseService{repo: repo, userRepo: userRepo}
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
	if strings.TrimSpace(courseDescription) != "" {
		course.CourseDescription = strings.TrimSpace(courseDescription)
	}
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
	if course.OwnerUserID != userID {
		return apperrors.ErrForbidden
	}
	course.Status = "deleted"
	return s.repo.UpdateCourse(ctx, course)
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
