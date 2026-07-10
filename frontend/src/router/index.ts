import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    children: [
      {
        path: '',
        redirect: '/courses',
      },
      {
        path: 'courses',
        name: 'courses',
        component: () => import('@/views/CourseListView.vue'),
      },
      {
        path: 'courses/:courseId',
        component: () => import('@/views/CourseLayoutView.vue'),
        children: [
          {
            path: '',
            redirect: (to: { params: { courseId?: string } }) => `/courses/${to.params.courseId}/overview`,
          },
          {
            path: 'overview',
            name: 'course-overview',
            component: () => import('@/views/CourseOverviewView.vue'),
          },
          {
            path: 'members',
            name: 'course-members',
            component: () => import('@/views/CourseMembersView.vue'),
          },
          {
            path: 'materials',
            name: 'course-materials',
            component: () => import('@/views/CourseMaterialsView.vue'),
          },
          {
            path: 'agent',
            name: 'course-agent',
            component: () => import('@/views/CourseAgentView.vue'),
          },
        ],
      },
    ],
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { guestOnly: true },
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!auth.isLoggedIn.value && to.name !== 'login') {
    return { name: 'login' }
  }
  if (auth.isLoggedIn.value && to.meta.guestOnly) {
    return { name: 'courses' }
  }
  return true
})
