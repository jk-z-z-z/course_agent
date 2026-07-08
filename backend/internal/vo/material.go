package vo

import "time"

type MaterialTreeNodeVO struct {
	ID         uint64               `json:"id"`
	ParentID   *uint64              `json:"parentId,omitempty"`
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	FileExt    string               `json:"fileExt,omitempty"`
	MimeType   string               `json:"mimeType,omitempty"`
	FileSize   int64                `json:"fileSize,omitempty"`
	SortIndex  int                  `json:"sortIndex"`
	Children   []MaterialTreeNodeVO `json:"children,omitempty"`
	UpdatedAt  time.Time            `json:"updatedAt"`
}

type MaterialDetailVO struct {
	ID              uint64    `json:"id"`
	CourseID        uint64    `json:"courseId"`
	SpaceID         uint64    `json:"spaceId"`
	ParentID        *uint64   `json:"parentId,omitempty"`
	NodeType        string    `json:"nodeType"`
	NodeName        string    `json:"nodeName"`
	FileExt         string    `json:"fileExt,omitempty"`
	StoragePath     string    `json:"storagePath,omitempty"`
	MimeType        string    `json:"mimeType,omitempty"`
	FileSize        int64     `json:"fileSize"`
	LatestVersionNo int       `json:"latestVersionNo"`
	SortIndex       int       `json:"sortIndex"`
	CreatedBy       uint64    `json:"createdBy"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
