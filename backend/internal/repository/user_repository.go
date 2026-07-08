package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"course_agent_backend/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsernameOrPhone(ctx context.Context, identifier string) (*model.User, error) {
	trimmed := strings.TrimSpace(identifier)
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("username = ?", trimmed).
		Or("phone = ?", trimmed).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
