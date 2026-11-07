import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { ElButton, ElInput } from 'element-plus'
import 'element-plus/dist/index.css'
import './style.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElButton)
app.use(ElInput)
app.mount('#app')
