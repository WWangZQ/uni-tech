import axios from 'axios'
import { useAuthStore } from '../stores/authStore'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api

// Auth API
export const authApi = {
  login: (studentId: string, password: string) =>
    api.post('/auth/login', { student_id: studentId, password }),
  register: (data: { student_id: string; name: string; email: string; password: string }) =>
    api.post('/auth/register', data),
}

// User API
export const userApi = {
  getCurrentUser: () => api.get('/users/me'),
  getCredits: () => api.get('/users/credits'),
  getQuota: () => api.get('/users/quota'),
}

// Space API
export const spaceApi = {
  getSpaces: (params?: { type?: string; building?: string }) =>
    api.get('/spaces', { params }),
  getSpace: (id: number) => api.get(`/spaces/${id}`),
  getSlots: (id: number, date: string) =>
    api.get(`/spaces/${id}/slots`, { params: { date } }),
  createBooking: (data: { resource_id: number; slot_id: number; date: string }) =>
    api.post('/spaces/bookings', data),
  getBookings: () => api.get('/bookings'),
  cancelBooking: (id: number) => api.delete(`/bookings/${id}`),
}

// Activity API
export const activityApi = {
  getActivities: (params?: { status?: string }) =>
    api.get('/activities', { params }),
  getActivity: (id: number) => api.get(`/activities/${id}`),
  doSeckill: (id: number) => api.post(`/activities/${id}/seckill`),
  getMyTicket: (id: number) => api.get(`/activities/${id}/ticket`),
}

// Order API
export const orderApi = {
  getOrders: () => api.get('/orders'),
  getOrder: (id: number) => api.get(`/orders/${id}`),
  payOrder: (id: number) => api.post(`/orders/${id}/pay`),
  cancelOrder: (id: number) => api.post(`/orders/${id}/cancel`),
}
