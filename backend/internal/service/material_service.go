package service

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/vo"
)

type MaterialService struct {
	courseRepo   *repository.CourseRepository
	materialRepo *repository.MaterialRepository
}

func NewMaterialService(courseRepo *repository.CourseRepository, materialRepo *repository.MaterialRepository) *MaterialService {
	return &MaterialService{courseRepo: courseRepo, materialRepo: materialRepo}
}

func (s *MaterialService) GetTree(ctx context.Context, userID, courseID uint64) ([]vo.MaterialTreeNodeVO, error) {
	if _, _, err := s.requireCourseMember(ctx, userID, courseID); err != nil {
		return nil, err
	}

	nodes, err := s.materialRepo.ListActiveNodesByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	return buildMaterialTree(nodes), nil
}

func (s *MaterialService) CreateFolder(ctx context.Context, userID, courseID uint64, parentID *uint64, folderName string) (*vo.MaterialDetailVO, error) {
	space, role, err := s.requireCourseManager(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	_ = role

	name := strings.TrimSpace(folderName)
	if name == "" {
		return nil, apperrors.ErrInvalidParameter
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return nil, apperrors.ErrInvalidParameter
	}

	if parentID != nil {
		parent, err := s.materialRepo.GetActiveNodeByID(ctx, courseID, *parentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.ErrMaterialNotFound
			}
			return nil, err
		}
		if parent.NodeType != "folder" {
			return nil, apperrors.ErrInvalidParameter
		}
	}

	if _, err := s.materialRepo.GetNodeByParentAndName(ctx, courseID, parentID, name); err == nil {
		return nil, apperrors.ErrMaterialExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	node := &model.CourseMaterialNode{
		CourseID:        courseID,
		SpaceID:         space.ID,
		ParentID:        parentID,
		NodeType:        "folder",
		NodeName:        name,
		LatestVersionNo: 1,
		SortIndex:       0,
		CreatedBy:       userID,
	}
	if err := s.materialRepo.CreateNode(ctx, node); err != nil {
		return nil, err
	}
	result := toMaterialDetailVO(node)
	return &result, nil
}

func (s *MaterialService) GetDetail(ctx context.Context, userID, courseID, nodeID uint64) (*vo.MaterialDetailVO, error) {
	if _, _, err := s.requireCourseMember(ctx, userID, courseID); err != nil {
		return nil, err
	}
	node, err := s.materialRepo.GetActiveNodeByID(ctx, courseID, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMaterialNotFound
		}
		return nil, err
	}
	result := toMaterialDetailVO(node)
	return &result, nil
}

func (s *MaterialService) UpdateNode(ctx context.Context, userID, courseID, nodeID uint64, nodeName string, sortIndex *int) (*vo.MaterialDetailVO, error) {
	if _, _, err := s.requireCourseManager(ctx, userID, courseID); err != nil {
		return nil, err
	}
	node, err := s.materialRepo.GetActiveNodeByID(ctx, courseID, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMaterialNotFound
		}
		return nil, err
	}

	if strings.TrimSpace(nodeName) != "" && strings.TrimSpace(nodeName) != node.NodeName {
		newName := strings.TrimSpace(nodeName)
		if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
			return nil, apperrors.ErrInvalidParameter
		}
		if _, err := s.materialRepo.GetNodeByParentAndName(ctx, courseID, node.ParentID, newName); err == nil {
			return nil, apperrors.ErrMaterialExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		node.NodeName = newName
	}
	if sortIndex != nil {
		node.SortIndex = *sortIndex
	}
	if err := s.materialRepo.UpdateNode(ctx, node); err != nil {
		return nil, err
	}
	result := toMaterialDetailVO(node)
	return &result, nil
}

func (s *MaterialService) DeleteNode(ctx context.Context, userID, courseID, nodeID uint64) error {
	if _, _, err := s.requireCourseManager(ctx, userID, courseID); err != nil {
		return err
	}
	node, err := s.materialRepo.GetActiveNodeByID(ctx, courseID, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrMaterialNotFound
		}
		return err
	}

	nodes, err := s.materialRepo.ListActiveDescendantsByPath(ctx, courseID)
	if err != nil {
		return err
	}
	toDelete := collectMaterialSubtree(nodes, node.ID)
	for i := range toDelete {
		toDelete[i].IsDeleted = true
	}
	return s.materialRepo.UpdateNodes(ctx, toDelete)
}

func (s *MaterialService) requireCourseMember(ctx context.Context, userID, courseID uint64) (*model.CourseStorageSpace, string, error) {
	course, err := s.courseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperrors.ErrCourseNotFound
		}
		return nil, "", err
	}
	if course.Status != "active" {
		return nil, "", apperrors.ErrCourseNotFound
	}
	member, err := s.courseRepo.GetMember(ctx, courseID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperrors.ErrForbidden
		}
		return nil, "", err
	}
	if member.JoinStatus != "active" {
		return nil, "", apperrors.ErrForbidden
	}
	space, err := s.materialRepo.GetStorageSpaceByCourseID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperrors.ErrStorageSpaceNotFound
		}
		return nil, "", err
	}
	return space, member.Role, nil
}

func (s *MaterialService) requireCourseManager(ctx context.Context, userID, courseID uint64) (*model.CourseStorageSpace, string, error) {
	space, role, err := s.requireCourseMember(ctx, userID, courseID)
	if err != nil {
		return nil, "", err
	}
	if role != "owner" && role != "teacher" {
		return nil, "", apperrors.ErrForbidden
	}
	return space, role, nil
}

func buildMaterialTree(nodes []model.CourseMaterialNode) []vo.MaterialTreeNodeVO {
	byParent := make(map[uint64][]model.CourseMaterialNode)
	roots := make([]model.CourseMaterialNode, 0)
	for _, node := range nodes {
		if node.ParentID == nil {
			roots = append(roots, node)
			continue
		}
		byParent[*node.ParentID] = append(byParent[*node.ParentID], node)
	}

	var walk func(model.CourseMaterialNode) vo.MaterialTreeNodeVO
	walk = func(node model.CourseMaterialNode) vo.MaterialTreeNodeVO {
		children := byParent[node.ID]
		result := vo.MaterialTreeNodeVO{
			ID:        node.ID,
			ParentID:  node.ParentID,
			Name:      node.NodeName,
			Type:      node.NodeType,
			FileExt:   node.FileExt,
			MimeType:  node.MimeType,
			FileSize:  node.FileSize,
			SortIndex: node.SortIndex,
			UpdatedAt: node.UpdatedAt,
		}
		if len(children) > 0 {
			result.Children = make([]vo.MaterialTreeNodeVO, 0, len(children))
			for _, child := range children {
				result.Children = append(result.Children, walk(child))
			}
		}
		return result
	}

	result := make([]vo.MaterialTreeNodeVO, 0, len(roots))
	for _, root := range roots {
		result = append(result, walk(root))
	}
	return result
}

func collectMaterialSubtree(nodes []model.CourseMaterialNode, rootID uint64) []model.CourseMaterialNode {
	children := make(map[uint64][]model.CourseMaterialNode)
	byID := make(map[uint64]model.CourseMaterialNode)
	for _, node := range nodes {
		byID[node.ID] = node
		if node.ParentID != nil {
			children[*node.ParentID] = append(children[*node.ParentID], node)
		}
	}

	result := make([]model.CourseMaterialNode, 0)
	queue := []uint64{rootID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		node, ok := byID[current]
		if !ok {
			continue
		}
		result = append(result, node)
		for _, child := range children[current] {
			queue = append(queue, child.ID)
		}
	}
	return result
}

func toMaterialDetailVO(node *model.CourseMaterialNode) vo.MaterialDetailVO {
	return vo.MaterialDetailVO{
		ID:              node.ID,
		CourseID:        node.CourseID,
		SpaceID:         node.SpaceID,
		ParentID:        node.ParentID,
		NodeType:        node.NodeType,
		NodeName:        node.NodeName,
		FileExt:         node.FileExt,
		StoragePath:     node.StoragePath,
		MimeType:        node.MimeType,
		FileSize:        node.FileSize,
		LatestVersionNo: node.LatestVersionNo,
		SortIndex:       node.SortIndex,
		CreatedBy:       node.CreatedBy,
		CreatedAt:       node.CreatedAt,
		UpdatedAt:       node.UpdatedAt,
	}
}
