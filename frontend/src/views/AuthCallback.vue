<template>
  <div class="callback-container">
    <a-spin tip="正在处理登录..." size="large" :spinning="loading">
      <div class="callback-content">
        <a-result v-if="error" status="error" :title="error" :sub-title="errorMessage">
          <template #extra>
            <a-button type="primary" @click="goToLogin">返回登录</a-button>
          </template>
        </a-result>
      </div>
    </a-spin>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { useUserStore } from '../stores/user'
import { authAPI } from '../api/services'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(true)
const error = ref('')
const errorMessage = ref('')

const goToLogin = () => {
  router.push('/login')
}

onMounted(async () => {
  try {
    // 检查URL中是否直接包含token和user_id
    const token = route.query.token
    const userId = route.query.user_id
    const username = route.query.username
    
    // 如果URL中包含token和user_id，直接使用
    if (token && userId) {
      // 保存 token 和 user_id
      userStore.setAuth(token, userId, username)
      
      message.success('登录成功')
      
      // 跳转到首页
      router.push('/')
      return
    }
    
    // 从 URL 获取授权码
    const code = route.query.code
    
    if (!code) {
      error.value = '授权失败'
      errorMessage.value = '未获取到授权码，请重新登录'
      return
    }
    
    // 处理 GitHub 回调
    const response = await authAPI.handleGithubCallback(code)
    
    // 保存 token 和 user_id
    userStore.setAuth(response.token, response.user_id)
    
    // 获取用户信息
    await userStore.fetchUserInfo()
    
    message.success('登录成功')
    
    // 跳转到首页
    router.push('/')
  } catch (error) {
    console.error('处理回调失败:', error)
    error.value = '授权失败'
    errorMessage.value = '处理 GitHub 授权回调失败，请重新登录'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.callback-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f0f2f5;
}

.callback-content {
  background-color: white;
  padding: 24px;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  min-height: 200px;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>