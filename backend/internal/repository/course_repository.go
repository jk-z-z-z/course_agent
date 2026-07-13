package repository

import (
	"context"

	"gorm.io/gorm"

	"course_agent_backend/internal/model"
)

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) CreateCourse(ctx context.Context, course *model.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *CourseRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *CourseRepository) GetCourseByID(ctx context.Context, courseID uint64) (*model.Course, error) {
	var course model.Course
	if err := r.db.WithContext(ctx).First(&course, courseID).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) GetCourseByCode(ctx context.Context, courseCode string) (*model.Course, error) {
	var course model.Course
	if err := r.db.WithContext(ctx).Where("course_code = ?", courseCode).First(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) ListCoursesByUserID(ctx context.Context, userID uint64) ([]model.CourseMember, error) {
	var members []model.CourseMember
	if err := r.db.WithContext(ctx).Where("user_id = ? AND join_status = ?", userID, "active").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *CourseRepository) ListActiveCourses(ctx context.Context) ([]model.Course, error) {
	var courses []model.Course
	if err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("updated_at DESC, id DESC").
		Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *CourseRepository) GetMember(ctx context.Context, courseID, userID uint64) (*model.CourseMember, error) {
	var member model.CourseMember
	if err := r.db.WithContext(ctx).Where("course_id = ? AND user_id = ?", courseID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *CourseRepository) GetMemberByID(ctx context.Context, memberID uint64) (*model.CourseMember, error) {
	var member model.CourseMember
	if err := r.db.WithContext(ctx).First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *CourseRepository) CreateMember(ctx context.Context, member *model.CourseMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *CourseRepository) UpdateCourse(ctx context.Context, course *model.Course) error {
	return r.db.WithContext(ctx).Save(course).Error
}

func (r *CourseRepository) UpdateMember(ctx context.Context, member *model.CourseMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *CourseRepository) ListMembersByCourseID(ctx context.Context, courseID uint64) ([]model.CourseMember, error) {
	var members []model.CourseMember
	if err := r.db.WithContext(ctx).Where("course_id = ? AND join_status = ?", courseID, "active").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
