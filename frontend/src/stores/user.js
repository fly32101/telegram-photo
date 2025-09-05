import { defineStore } from 'pinia'
import { authAPI } from '../api/services'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userId: localStorage.getItem('user_id') || '',
    username: localStorage.getItem('username') || '',
    userInfo: null,
    loading: false,
    error: null
  }),
  
  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => state.userInfo?.is_admin || false
  },
  
  actions: {
    // 设置认证信息
    setAuth(token, userId, username = '') {
      this.token = token
      this.userId = userId
      this.username = username
      localStorage.setItem('token', token)
      localStorage.setItem('user_id', userId)
      if (username) {
        localStorage.setItem('username', username)
      }
    },
    
    // 清除认证信息
    clearAuth() {
      this.token = ''
      this.userId = ''
      this.username = ''
      this.userInfo = null
      localStorage.removeItem('token')
      localStorage.removeItem('user_id')
      localStorage.removeItem('username')
    },
    
    // 获取用户信息
    async fetchUserInfo() {
      try {
        this.loading = true
        this.error = null
        const response = await authAPI.getCurrentUser()
        this.userInfo = response.user
        return response.user
      } catch (error) {
        this.error = error.message || '获取用户信息失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    
    // 登出
    logout() {
      this.clearAuth()
    }
  }
})