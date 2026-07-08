import { request } from './http'
import type {
  CreateMaterialFolderPayload,
  MaterialDetailVO,
  MaterialTreeNodeVO,
  UpdateMaterialNodePayload,
} from '@/types/material'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

async function requestBlob(path: string, token: string) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: authHeaders(token),
  })

  if (!response.ok) {
    let message = '请求失败'
    try {
      const payload = await response.json()
      message = payload.message ?? message
    } catch {
      // ignore non-json bodies
    }
    throw new Error(message)
  }
  return response.blob()
}

export function getMaterialTree(token: string, courseId: number) {
  return request<MaterialTreeNodeVO[]>(`/api/courses/${courseId}/materials/tree`, {
    headers: authHeaders(token),
  })
}

export function getMaterialDetail(token: string, courseId: number, nodeId: number) {
  return request<MaterialDetailVO>(`/api/courses/${courseId}/materials/${nodeId}`, {
    headers: authHeaders(token),
  })
}

export function createMaterialFolder(token: string, courseId: number, payload: CreateMaterialFolderPayload) {
  return request<MaterialDetailVO>(`/api/courses/${courseId}/materials/folders`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function updateMaterialNode(token: string, courseId: number, nodeId: number, payload: UpdateMaterialNodePayload) {
  return request<MaterialDetailVO>(`/api/courses/${courseId}/materials/${nodeId}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  })
}

export function deleteMaterialNode(token: string, courseId: number, nodeId: number) {
  return request<null>(`/api/courses/${courseId}/materials/${nodeId}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
}

export function uploadMaterialFile(token: string, courseId: number, file: File, parentId?: number) {
  const formData = new FormData()
  formData.append('file', file)
  if (typeof parentId === 'number') {
    formData.append('parentId', String(parentId))
  }

  return fetch(`${API_BASE_URL}/api/courses/${courseId}/materials/upload`, {
    method: 'POST',
    headers: authHeaders(token),
    body: formData,
  }).then(async (response) => {
    const payload = await response.json()
    if (!response.ok || payload.code !== 0) {
      throw new Error(payload.message ?? '上传失败')
    }
    return payload.data as MaterialDetailVO
  })
}

export function previewMaterial(token: string, courseId: number, nodeId: number) {
  return requestBlob(`/api/courses/${courseId}/materials/${nodeId}/preview`, token)
}

export function downloadMaterial(token: string, courseId: number, nodeId: number) {
  return requestBlob(`/api/courses/${courseId}/materials/${nodeId}/download`, token)
}
