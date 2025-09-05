import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'

// 导入样式
import './style.css'
import 'ant-design-vue/dist/reset.css'

const app = createApp(App)

// 使用插件
app.use(createPinia())
app.use(router)

app.mount('#app')
