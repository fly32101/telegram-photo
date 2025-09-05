<template>
  <a-layout class="layout">
    <app-header />
    
    <a-layout-content class="content">
      <a-tabs>
        <a-tab-pane key="images" tab="所有图片">
          <a-card :bordered="false">
            <template #title>
              <div class="filter-container">
                <a-form layout="inline" :model="filters">
                  <a-form-item label="用户ID">
                    <a-input v-model:value="filters.userId" placeholder="按用户ID筛选" />
                  </a-form-item>
                  <a-form-item label="上传IP">
                    <a-input v-model:value="filters.uploadIp" placeholder="按上传IP筛选" />
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleFilter">筛选</a-button>
                    <a-button style="margin-left: 8px" @click="resetFilter">重置</a-button>
                  </a-form-item>
                </a-form>
              </div>
            </template>
            
            <a-spin :spinning="adminStore.loading">
              <a-table
                :dataSource="adminStore.allImages"
                :columns="columns"
                :pagination="{
                  current: adminStore.currentPage,
                  pageSize: adminStore.pageSize,
                  total: adminStore.total,
                  onChange: handlePageChange
                }"
                rowKey="id"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'proxy_url'">
                    <a-image
                      :width="50"
                      :src="record.proxy_url"
                      :preview="{
                        src: record.proxy_url,
                      }"
                    />
                  </template>
                  <template v-if="column.key === 'action'">
                    <a-space>
                      <a-button type="link" size="small" @click="copyImageUrl(record.proxy_url)">
                        <template #icon><copy-outlined /></template>
                        复制链接
                      </a-button>
                    </a-space>
                  </template>
                </template>
              </a-table>
            </a-spin>
          </a-card>
        </a-tab-pane>
        
        <a-tab-pane key="stats" tab="统计信息">
          <a-spin :spinning="adminStore.loading">
            <a-row :gutter="[16, 16]" v-if="adminStore.stats">
              <a-col :span="8">
                <a-card>
                  <a-statistic
                    title="总图片数"
                    :value="adminStore.stats.total_images"
                    :value-style="{ color: '#3f8600' }"
                  >
                    <template #prefix>
                      <picture-outlined />
                    </template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="8">
                <a-card>
                  <a-statistic
                    title="今日上传"
                    :value="adminStore.stats.today_images"
                    :value-style="{ color: '#cf1322' }"
                  >
                    <template #prefix>
                      <rise-outlined />
                    </template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="8">
                <a-card>
                  <a-statistic
                    title="用户数"
                    :value="adminStore.stats.user_count"
                    :value-style="{ color: '#1677ff' }"
                  >
                    <template #prefix>
                      <team-outlined />
                    </template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="24">
                <a-card title="用户排行榜">
                  <a-table
                    :dataSource="adminStore.stats.user_rankings"
                    :columns="rankingColumns"
                    :pagination="false"
                    rowKey="UserID"
                  />
                </a-card>
              </a-col>
            </a-row>
          </a-spin>
        </a-tab-pane>
      </a-tabs>
    </a-layout-content>
    
    <a-layout-footer style="text-align: center">
      Telegram 图床 ©{{ new Date().getFullYear() }}
    </a-layout-footer>
  </a-layout>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { 
  CopyOutlined, 
  PictureOutlined, 
  RiseOutlined, 
  TeamOutlined 
} from '@ant-design/icons-vue'
import { useAdminStore } from '../stores/admin'
import AppHeader from '../components/AppHeader.vue'

const adminStore = useAdminStore()

// 表格列定义
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80
  },
  {
    title: '预览',
    dataIndex: 'proxy_url',
    key: 'proxy_url',
    width: 100
  },
  {
    title: '文件ID',
    dataIndex: 'file_id',
    key: 'file_id',
    ellipsis: true
  },
  {
    title: '用户ID',
    dataIndex: 'user_id',
    key: 'user_id',
    width: 150
  },
  {
    title: '上传IP',
    dataIndex: 'upload_ip',
    key: 'upload_ip',
    width: 150
  },
  {
    title: '上传时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 180,
    render: (text) => new Date(text).toLocaleString()
  },
  {
    title: '操作',
    key: 'action',
    width: 150
  }
]

// 排行榜列定义
const rankingColumns = [
  {
    title: '用户ID',
    dataIndex: 'UserID',
    key: 'UserID'
  },
  {
    title: '上传数量',
    dataIndex: 'Count',
    key: 'Count'
  }
]

// 筛选条件
const filters = reactive({
  userId: '',
  uploadIp: ''
})

onMounted(async () => {
  try {
    // 获取所有图片
    await adminStore.fetchAllImages()
    
    // 获取统计信息
    await adminStore.fetchStats()
  } catch (error) {
    console.error('获取数据失败:', error)
    message.error('获取数据失败，请稍后重试')
  }
})

// 处理筛选
const handleFilter = () => {
  adminStore.setFilters({
    userId: filters.userId,
    uploadIp: filters.uploadIp
  })
  adminStore.fetchAllImages({ page: 1 })
}

// 重置筛选
const resetFilter = () => {
  filters.userId = ''
  filters.uploadIp = ''
  adminStore.resetFilters()
  adminStore.fetchAllImages({ page: 1 })
}

// 处理分页变化
const handlePageChange = (page) => {
  adminStore.fetchAllImages({ page })
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
</script>

<style scoped>
.layout {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background-color: #f0f2f5;
}

.filter-container {
  margin-bottom: 16px;
}
</style>