<template>
  <a-layout-header class="header">
    <div class="logo">Telegram 图床</div>
    <div class="right">
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="horizontal"
        :style="{ lineHeight: '64px' }"
      >
        <a-menu-item key="home" @click="router.push('/')">
          <template #icon>
            <home-outlined />
          </template>
          首页
        </a-menu-item>
        
        <a-menu-item v-if="userStore.isAdmin" key="admin" @click="router.push('/admin')">
          <template #icon>
            <dashboard-outlined />
          </template>
          管理后台
        </a-menu-item>
        
        <a-sub-menu key="user">
          <template #icon>
            <user-outlined />
          </template>
          <template #title>{{ userStore.userInfo?.username || '用户' }}</template>
          <a-menu-item key="logout" @click="handleLogout">
            <template #icon>
              <logout-outlined />
            </template>
            退出登录
          </a-menu-item>
        </a-sub-menu>
      </a-menu>
    </div>
  </a-layout-header>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { 
  HomeOutlined, 
  DashboardOutlined, 
  UserOutlined, 
  LogoutOutlined 
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { useUserStore } from '../stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

// 当前选中的菜单项
const selectedKeys = ref([route.name === 'Admin' ? 'admin' : 'home'])

// 处理退出登录
const handleLogout = () => {
  userStore.logout()
  message.success('已退出登录')
  router.push('/login')
}

// 如果没有用户信息，则获取
onMounted(async () => {
  if (userStore.isAuthenticated && !userStore.userInfo) {
    try {
      await userStore.fetchUserInfo()
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }
  }
})
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
}

.logo {
  color: white;
  font-size: 18px;
  font-weight: bold;
}

.right {
  display: flex;
  align-items: center;
}
</style>