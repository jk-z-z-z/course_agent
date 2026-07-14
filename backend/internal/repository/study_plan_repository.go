package repository

import (
	"context"

	"gorm.io/gorm"

	"course_agent_backend/internal/model"
)

type StudyPlanRepository struct {
	db *gorm.DB
}

func NewStudyPlanRepository(db *gorm.DB) *StudyPlanRepository {
	return &StudyPlanRepository{db: db}
}

func (r *StudyPlanRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *StudyPlanRepository) CreatePlanTx(tx *gorm.DB, plan *model.StudyPlan) error {
	return tx.Create(plan).Error
}

func (r *StudyPlanRepository) CreateItemsTx(tx *gorm.DB, items []model.StudyPlanItem) error {
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}

func (r *StudyPlanRepository) UpdatePlanTx(tx *gorm.DB, plan *model.StudyPlan) error {
	return tx.Save(plan).Error
}

func (r *StudyPlanRepository) UpdateItem(ctx context.Context, item *model.StudyPlanItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *StudyPlanRepository) ListPlansByCourseIDAndUserID(ctx context.Context, courseID, userID uint64) ([]model.StudyPlan, error) {
	var plans []model.StudyPlan
	if err := r.db.WithContext(ctx).
		Where("course_id = ? AND user_id = ?", courseID, userID).
		Order("created_at DESC, id DESC").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *StudyPlanRepository) ListItemsByPlanID(ctx context.Context, planID uint64) ([]model.StudyPlanItem, error) {
	var items []model.StudyPlanItem
	if err := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("day_index ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *StudyPlanRepository) GetPlanByID(ctx context.Context, planID uint64) (*model.StudyPlan, error) {
	var plan model.StudyPlan
	if err := r.db.WithContext(ctx).First(&plan, planID).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *StudyPlanRepository) GetItemByID(ctx context.Context, itemID uint64) (*model.StudyPlanItem, error) {
	var item model.StudyPlanItem
	if err := r.db.WithContext(ctx).First(&item, itemID).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
