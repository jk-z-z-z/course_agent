export type MaterialNodeType = 'folder' | 'file'

export interface MaterialTreeNodeVO {
  id: number
  parentId?: number
  name: string
  type: MaterialNodeType
  fileExt?: string
  mimeType?: string
  fileSize?: number
  sortIndex: number
  updatedAt: string
  children?: MaterialTreeNodeVO[]
}

export interface MaterialDetailVO {
  id: number
  courseId: number
  spaceId: number
  parentId?: number
  nodeType: MaterialNodeType
  nodeName: string
  fileExt?: string
  storagePath?: string
  mimeType?: string
  fileSize: number
  latestVersionNo: number
  sortIndex: number
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface CreateMaterialFolderPayload {
  parentId?: number
  folderName: string
}

export interface UpdateMaterialNodePayload {
  nodeName: string
  sortIndex?: number
}
