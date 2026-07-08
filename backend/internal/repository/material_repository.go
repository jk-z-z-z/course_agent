package repository

import (
	"context"

	"gorm.io/gorm"

	"course_agent_backend/internal/model"
)

type MaterialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

func (r *MaterialRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *MaterialRepository) CreateStorageSpace(ctx context.Context, space *model.CourseStorageSpace) error {
	return r.db.WithContext(ctx).Create(space).Error
}

func (r *MaterialRepository) CreateStorageSpaceTx(tx *gorm.DB, space *model.CourseStorageSpace) error {
	return tx.Create(space).Error
}

func (r *MaterialRepository) GetStorageSpaceByCourseID(ctx context.Context, courseID uint64) (*model.CourseStorageSpace, error) {
	var space model.CourseStorageSpace
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).First(&space).Error; err != nil {
		return nil, err
	}
	return &space, nil
}

func (r *MaterialRepository) UpdateStorageSpace(ctx context.Context, space *model.CourseStorageSpace) error {
	return r.db.WithContext(ctx).Save(space).Error
}

func (r *MaterialRepository) CreateNode(ctx context.Context, node *model.CourseMaterialNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *MaterialRepository) CreateNodeTx(tx *gorm.DB, node *model.CourseMaterialNode) error {
	return tx.Create(node).Error
}

func (r *MaterialRepository) UpdateNode(ctx context.Context, node *model.CourseMaterialNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *MaterialRepository) UpdateNodeTx(tx *gorm.DB, node *model.CourseMaterialNode) error {
	return tx.Save(node).Error
}

func (r *MaterialRepository) CreateVersion(ctx context.Context, version *model.CourseMaterialVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *MaterialRepository) CreateVersionTx(tx *gorm.DB, version *model.CourseMaterialVersion) error {
	return tx.Create(version).Error
}

func (r *MaterialRepository) UpdateStorageSpaceTx(tx *gorm.DB, space *model.CourseStorageSpace) error {
	return tx.Save(space).Error
}

func (r *MaterialRepository) ListActiveNodesByCourseID(ctx context.Context, courseID uint64) ([]model.CourseMaterialNode, error) {
	var nodes []model.CourseMaterialNode
	if err := r.db.WithContext(ctx).
		Where("course_id = ? AND is_deleted = ?", courseID, false).
		Order("parent_id ASC, sort_index ASC, id ASC").
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *MaterialRepository) ListActiveFileNodesByCourseID(ctx context.Context, courseID uint64) ([]model.CourseMaterialNode, error) {
	var nodes []model.CourseMaterialNode
	if err := r.db.WithContext(ctx).
		Where("course_id = ? AND is_deleted = ? AND node_type = ?", courseID, false, "file").
		Order("id ASC").
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *MaterialRepository) GetActiveNodeByID(ctx context.Context, courseID, nodeID uint64) (*model.CourseMaterialNode, error) {
	var node model.CourseMaterialNode
	if err := r.db.WithContext(ctx).
		Where("course_id = ? AND id = ? AND is_deleted = ?", courseID, nodeID, false).
		First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *MaterialRepository) GetNodeByParentAndName(ctx context.Context, courseID uint64, parentID *uint64, nodeName string) (*model.CourseMaterialNode, error) {
	var node model.CourseMaterialNode
	query := r.db.WithContext(ctx).Where("course_id = ? AND node_name = ? AND is_deleted = ?", courseID, nodeName, false)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	if err := query.First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *MaterialRepository) ListActiveDescendantsByPath(ctx context.Context, courseID uint64) ([]model.CourseMaterialNode, error) {
	var nodes []model.CourseMaterialNode
	if err := r.db.WithContext(ctx).
		Where("course_id = ? AND is_deleted = ?", courseID, false).
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *MaterialRepository) UpdateNodes(ctx context.Context, nodes []model.CourseMaterialNode) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range nodes {
			if err := tx.Save(&nodes[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
