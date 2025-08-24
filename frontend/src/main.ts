import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import App from './App.vue'

// Import global styles
import './styles/main.css'

// Import views
import Login from './views/Login.vue'
import Dashboard from './views/Dashboard.vue'
import Trade from './views/Trade.vue'

// Define routes with proper typing
const routes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: '/dashboard'
    },
    {
        path: '/login',
        name: 'Login',
        component: Login
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: Dashboard
    },
    {
        path: '/trade',
        name: 'Trade',
        component: Trade
    }
]

// Create router
const router = createRouter({
    history: createWebHistory(),
    routes
})

// Create Pinia store
const pinia = createPinia()

// Create and mount app
createApp(App)
    .use(pinia)
    .use(router)
    .mount('#app')