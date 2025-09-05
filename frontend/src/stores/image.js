import { defineStore } from 'pinia'
import { imageAPI } from '../api/services'

export const useImageStore = defineStore('image', {
  state: () => ({
    images: [],
    total: 0,
    currentPage: 1,
    pageSize: 10,
    loading: false,
    uploading: false,
    error: null
  }),
  
  actions: {
    // 获取用户图片列表
    async fetchUserImages(page = 1, pageSize = 10) {
      try {
        this.loading = true
        this.error = null
        this.currentPage = page
        this.pageSize = pageSize
        
        const response = await imageAPI.getUserImages(page, pageSize)
        
        this.images = response.images
        
        this.total = response.total
        return response
      } catch (error) {
        this.error = error.message || '获取图片列表失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    
    // 上传图片
    async uploadImage(file, autoRefresh = true) {
      try {
        this.uploading = true
        this.error = null
        
        const formData = new FormData()
        formData.append('image', file)
        
        const response = await imageAPI.uploadImage(formData)
        
        // 只有当autoRefresh为true时才刷新图片列表
        if (autoRefresh) {
          await this.fetchUserImages(this.currentPage, this.pageSize)
        }
        
        return response
      } catch (error) {
        this.error = error.message || '上传图片失败'
        throw error
      } finally {
        this.uploading = false
      }
    },
    
    // 删除图片
    async deleteImage(id) {
      try {
        this.loading = true
        this.error = null
        
        await imageAPI.deleteImage(id)
        // 刷新图片列表
        await this.fetchUserImages(this.currentPage, this.pageSize)
        return { success: true }
      } catch (error) {
        this.error = error.message || '删除图片失败'
        throw error
      } finally {
        this.loading = false
      }
    }
  }
})