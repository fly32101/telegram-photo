<template>
  <a-layout class="layout">
    <app-header />
    
    <a-layout-content class="content">
      <a-row :gutter="[16, 16]">
        <a-col :span="24">
          <a-card title="上传图片" :bordered="false">
            <a-upload-dragger
              name="image"
              :multiple="false"
              :before-upload="beforeUpload"
              :show-upload-list="false"
              @change="handleUploadChange"
            >
              <p class="ant-upload-drag-icon">
                <inbox-outlined />
              </p>
              <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
              <p class="ant-upload-hint">
                支持单个图片上传，请确保图片格式正确
              </p>
            </a-upload-dragger>
          </a-card>
        </a-col>
        
        <a-col :span="24">
          <a-card title="我的图片" :bordered="false">
            <a-spin :spinning="imageStore.loading">
              <a-empty v-if="imageStore.images.length === 0" description="暂无图片" />
              
              <div v-else class="image-list">
                <a-row :gutter="[16, 16]">
                  <a-col :xs="24" :sm="12" :md="8" :lg="6" v-for="image in imageStore.images" :key="image.id">
                    <a-card hoverable class="image-card">
                      <template #cover>
                        <img 
                          :src="image.proxy_url" 
                          alt="图片" 
                          class="image-preview" 
                        />
                      </template>
                      <template #actions>
                        <copy-outlined @click="copyImageUrl(image.proxy_url)" />
                        <delete-outlined @click="confirmDelete(image.id)" />
                      </template>
                      <a-card-meta :title="formatDate(image.created_at)">
                        <template #description>
                          <a-typography-paragraph copyable :content="image.proxy_url" />
                        </template>
                      </a-card-meta>
                    </a-card>
                  </a-col>
                </a-row>
                
                <a-pagination
                  v-if="imageStore.total > imageStore.pageSize"
                  class="pagination"
                  :current="imageStore.currentPage"
                  :pageSize="imageStore.pageSize"
                  :total="imageStore.total"
                  @change="handlePageChange"
                />
              </div>
            </a-spin>
          </a-card>
        </a-col>
      </a-row>
    </a-layout-content>
    
    <a-layout-footer style="text-align: center">
      Telegram 图床 ©{{ new Date().getFullYear() }}
    </a-layout-footer>
  </a-layout>
</template>

<script setup>
import { onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { InboxOutlined, CopyOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { useImageStore } from '../stores/image'
import AppHeader from '../components/AppHeader.vue'

const imageStore = useImageStore()

// 处理页面变化
const handlePageChange = async (page) => {
  try {
    await imageStore.fetchUserImages(page, imageStore.pageSize)
  } catch (error) {
    console.error('获取图片列表失败:', error)
    message.error('获取图片列表失败，请稍后重试')
  }
}

onMounted(async () => {
  try {
    // 获取图片列表
    await imageStore.fetchUserImages()
  } catch (error) {
    console.error('获取图片列表失败:', error)
    message.error('获取图片列表失败')
  }
})

// 上传前检查
const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    message.error('只能上传图片文件!')
  }
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isLt5M) {
    message.error('图片大小不能超过 5MB!')
  }
  return isImage && isLt5M
}

// 处理上传变化
const handleUploadChange = async (info) => {
  // 只在状态为uploading且文件对象存在时处理上传
  if (info.file.status !== 'uploading' || !info.file.originFileObj) {
    return
  }
  
  // 防止重复上传：给文件对象添加标记
  if (info.file.originFileObj._isUploading) {
    return
  }
  info.file.originFileObj._isUploading = true
  
  try {
    // 使用uploadImage方法，并设置autoRefresh为true，确保上传后刷新列表
    const response = await imageStore.uploadImage(info.file.originFileObj, true)
    if (response.existing) {
      message.info('图片已存在，已返回已有图片链接')
    } else {
      message.success('上传成功')
    }
    
    // 上传成功后已自动刷新图片列表
  } catch (error) {
    console.error('上传失败:', error)
    message.error('上传失败，请稍后重试')
  } finally {
    // 上传完成后移除标记
    if (info.file.originFileObj) {
      info.file.originFileObj._isUploading = false
    }
  }
}

// 复制图片链接
const copyImageUrl = (url) => {
  navigator.clipboard.writeText(url)
    .then(() => {
      message.success('链接已复制到剪贴板')
    })
    .catch(() => {
      message.error('复制失败，请手动复制')
    })
}

// 确认删除
const confirmDelete = (id) => {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这张图片吗？删除后将无法恢复。',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        await imageStore.deleteImage(id)
        message.success('删除成功')
      } catch (error) {
        console.error('删除失败:', error)
        message.error('删除失败，请稍后重试')
      }
    }
  })
}

// 格式化日期
const formatDate = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleString()
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background-color: #f0f2f5;
}

.image-list {
  margin-top: 16px;
}

.image-card {
  margin-bottom: 16px;
}

.image-preview {
  height: 200px;
  object-fit: cover;
}

.pagination {
  margin-top: 16px;
  text-align: center;
}
</style>