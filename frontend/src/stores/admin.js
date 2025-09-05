import { defineStore } from 'pinia'
import { adminAPI } from '../api/services'

export const useAdminStore = defineStore('admin', {
  state: () => ({
    allImages: [],
    total: 0,
    currentPage: 1,
    pageSize: 20,
    stats: null,
    loading: false,
    error: null,
    filters: {
      userId: '',
      uploadIp: ''
    }
  }),
  
  actions: {
    // 获取所有图片
    async fetchAllImages(params = {}) {
      try {
        this.loading = true
        this.error = null
        
        const queryParams = {
          page: params.page || this.currentPage,
          page_size: params.pageSize || this.pageSize,
          ...this.filters
        }
        
        if (params.userId) queryParams.user_id = params.userId
        if (params.uploadIp) queryParams.upload_ip = params.uploadIp
        
        const response = await adminAPI.getAllImages(queryParams)
        this.allImages = response.images
        this.total = response.total
        this.currentPage = response.page
        return response
      } catch (error) {
        this.error = error.message || '获取图片列表失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    
    // 获取统计信息
    async fetchStats() {
      try {
        this.loading = true
        this.error = null
        
        const response = await adminAPI.getStats()
        this.stats = response
        return response
      } catch (error) {
        this.error = error.message || '获取统计信息失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    
    // 设置过滤条件
    setFilters(filters) {
      this.filters = { ...this.filters, ...filters }
    },
    
    // 重置过滤条件
    resetFilters() {
      this.filters = {
        userId: '',
        uploadIp: ''
      }
    }
  }
})