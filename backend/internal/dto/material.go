package dto

type CreateMaterialFolderRequest struct {
	ParentID   *uint64 `json:"parentId"`
	FolderName string  `json:"folderName"`
}

type UpdateMaterialNodeRequest struct {
	NodeName  string `json:"nodeName"`
	SortIndex *int   `json:"sortIndex,omitempty"`
}
