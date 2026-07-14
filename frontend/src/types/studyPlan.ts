export type StudyPlanStatus = 'active' | 'archived'
export type StudyPlanItemStatus = 'pending' | 'done'

export interface StudyPlanItemVO {
  id: number
  planId: number
  dayIndex: number
  planDate: string
  title: string
  tasksText: string
  suggestedMinutes: number
  materialNodeIds: number[]
  status: StudyPlanItemStatus
  createdAt: string
  updatedAt: string
}

export interface StudyPlanVO {
  id: number
  courseId: number
  userId: number
  goal: string
  dailyMinutes: number
  status: StudyPlanStatus
  generatedSummary: string
  createdAt: string
  updatedAt: string
  items: StudyPlanItemVO[]
}

export interface StudyPlanSummaryVO {
  id: number
  courseId: number
  userId: number
  goal: string
  dailyMinutes: number
  status: StudyPlanStatus
  generatedSummary: string
  itemCount: number
  doneItemCount: number
  createdAt: string
  updatedAt: string
}

export interface GenerateStudyPlanPayload {
  goal: string
  dailyMinutes: number
}
