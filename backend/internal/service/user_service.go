package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/vo"
)

type UserService struct {
	repo     *repository.UserRepository
	redis    *redis.Client
	tokenTTL time.Duration
}

func NewUserService(repo *repository.UserRepository, redisClient *redis.Client) *UserService {
	return &UserService{repo: repo, redis: redisClient, tokenTTL: 7 * 24 * time.Hour}
}

func (s *UserService) Register(ctx context.Context, username, password, phone string) (*vo.UserVO, error) {
	if username == "" || password == "" {
		return nil, apperrors.ErrInvalidParameter
	}

	if _, err := s.repo.GetByUsername(ctx, username); err == nil {
		return nil, apperrors.ErrUserExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Phone:        phone,
		Status:       "active",
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserVO(user), nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, time.Time, *vo.UserVO, error) {
	if username == "" || password == "" {
		return "", time.Time{}, nil, apperrors.ErrInvalidParameter
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", time.Time{}, nil, apperrors.ErrUserNotFound
		}
		return "", time.Time{}, nil, err
	}
	if user.Status != "active" {
		return "", time.Time{}, nil, apperrors.ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", time.Time{}, nil, apperrors.ErrSessionExpired
	}

	token, err := generateToken()
	if err != nil {
		return "", time.Time{}, nil, err
	}

	expiredAt := time.Now().Add(s.tokenTTL)
	if err := s.redis.Set(ctx, sessionKey(token), user.ID, s.tokenTTL).Err(); err != nil {
		return "", time.Time{}, nil, fmt.Errorf("store session: %w", err)
	}

	return token, expiredAt, toUserVO(user), nil
}

func (s *UserService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return apperrors.ErrUnauthorized
	}
	if err := s.redis.Del(ctx, sessionKey(token)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *UserService) Me(ctx context.Context, userID uint64) (*vo.UserVO, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	if user.Status != "active" {
		return nil, apperrors.ErrUserDisabled
	}
	return toUserVO(user), nil
}

func (s *UserService) ResolveToken(ctx context.Context, token string) (uint64, error) {
	if token == "" {
		return 0, apperrors.ErrUnauthorized
	}
	value, err := s.redis.Get(ctx, sessionKey(token)).Uint64()
	if err == redis.Nil {
		return 0, apperrors.ErrSessionExpired
	}
	if err != nil {
		return 0, err
	}
	return value, nil
}

func sessionKey(token string) string {
	return "auth:session:" + token
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toUserVO(user *model.User) *vo.UserVO {
	return &vo.UserVO{
		ID:       user.ID,
		Username: user.Username,
		Phone:    user.Phone,
		Status:   user.Status,
	}
}
