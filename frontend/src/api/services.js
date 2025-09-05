import api from './index'

// 认证相关 API
export const authAPI = {
  // 获取 GitHub 授权 URL
  getGithubAuthUrl: () => api.get('/api/v1/auth/github'),
  
  // 处理 GitHub 回调
  handleGithubCallback: (code) => api.get(`/api/v1/auth/github/callback?code=${code}`),
  
  // 获取当前用户信息
  getCurrentUser: () => api.get('/api/v1/auth/user')
}

// 图片相关 API
export const imageAPI = {
  // 上传图片
  uploadImage: (formData) => api.post('/api/v1/image/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  }),
  
  // 获取用户图片列表
  getUserImages: (page = 1, pageSize = 10) => 
    api.get(`/api/v1/image/list?page=${page}&page_size=${pageSize}`),
  
  // 删除图片
  deleteImage: (id) => api.delete(`/api/v1/image/${id}`)
}

// 管理员 API
export const adminAPI = {
  // 获取所有图片
  getAllImages: (params) => api.get('/api/v1/admin/images', { params }),
  
  // 获取统计信息
  getStats: () => api.get('/api/v1/admin/stats')
}