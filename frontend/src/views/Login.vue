<template>
  <div class="login-container">
    <a-card title="Telegram 图床" class="login-card">
      <a-space direction="vertical" size="large" style="width: 100%">
        <a-typography-title :level="4" style="text-align: center">
          使用 GitHub 账号登录
        </a-typography-title>
        
        <a-button type="primary" block @click="handleLogin" :loading="loading">
          <template #icon>
            <github-outlined />
          </template>
          GitHub 登录
        </a-button>
      </a-space>
    </a-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { GithubOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { authAPI } from '../api/services'

const loading = ref(false)

const handleLogin = async () => {
  try {
    loading.value = true
    // 使用window.open打开新窗口，避免跨域问题
    window.open('/api/v1/auth/github', '_self')
  } catch (error) {
    console.error('登录失败:', error)
    message.error('获取 GitHub 授权链接失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f0f2f5;
}

.login-card {
  width: 400px;
  max-width: 90%;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>