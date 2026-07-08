package model

import "time"

type CourseStorageSpace struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	CourseID   uint64    `gorm:"not null;uniqueIndex"`
	RootPath   string    `gorm:"size:255;not null"`
	QuotaBytes int64     `gorm:"not null;default:0"`
	UsedBytes  int64     `gorm:"not null;default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (CourseStorageSpace) TableName() string {
	return "course_storage_spaces"
}

type CourseMaterialNode struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	CourseID        uint64    `gorm:"not null;index"`
	SpaceID         uint64    `gorm:"not null;index"`
	ParentID        *uint64   `gorm:"index"`
	NodeType        string    `gorm:"size:16;not null"`
	NodeName        string    `gorm:"size:255;not null"`
	FileExt         string    `gorm:"size:32"`
	StoragePath     string    `gorm:"size:512"`
	MimeType        string    `gorm:"size:128"`
	FileSize        int64     `gorm:"not null;default:0"`
	LatestVersionNo int       `gorm:"not null;default:1"`
	SortIndex       int       `gorm:"not null;default:0"`
	IsDeleted       bool      `gorm:"not null;default:false;index"`
	CreatedBy       uint64    `gorm:"not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (CourseMaterialNode) TableName() string {
	return "course_material_nodes"
}

type CourseMaterialVersion struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement"`
	MaterialNodeID uint64    `gorm:"not null;uniqueIndex:idx_node_version"`
	VersionNo      int       `gorm:"not null;uniqueIndex:idx_node_version"`
	StoragePath    string    `gorm:"size:512;not null"`
	FileSize       int64     `gorm:"not null"`
	MimeType       string    `gorm:"size:128"`
	UploadUserID   uint64    `gorm:"not null;index"`
	CreatedAt      time.Time
}

func (CourseMaterialVersion) TableName() string {
	return "course_material_versions"
}
