import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { loadStoredToken } from './services/api.js'
loadStoredToken()
import App from './App.vue'

// Import global styles
import './styles/main.css'

// Import views
import Login from './views/Login.vue'
import Dashboard from './views/Dashboard.vue'
import Trade from './views/Trade.vue'
import TransactionHistory from './views/TransactionHistory.vue'

// Import auth service for route guards
import authService from './services/authService.js'

// Define routes with proper typing
const routes: RouteRecordRaw[] = [
    {
        path: '/',
        redirect: (to) => {
            // Redirect to dashboard if authenticated, otherwise to login
            return authService.isAuthenticated() ? '/dashboard' : '/login'
        }
    },
    {
        path: '/login',
        name: 'Login',
        component: Login,
        beforeEnter: (to, from, next) => {
            // If already authenticated, redirect to dashboard
            if (authService.isAuthenticated()) {
                next('/dashboard')
            } else {
                next()
            }
        }
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: Dashboard,
        meta: { requiresAuth: true }
    },
    {
        path: '/trade',
        name: 'Trade',
        component: Trade,
        meta: { requiresAuth: true }
    },
    {
        path: '/transactions',
        name: 'TransactionHistory',
        component: TransactionHistory,
        meta: { requiresAuth: true }
    }
]

// Create router
const router = createRouter({
    history: createWebHistory(),
    routes
})

// Global navigation guard
router.beforeEach((to, from, next) => {
    const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
    const isAuthenticated = authService.isAuthenticated()

    if (requiresAuth && !isAuthenticated) {
        // Route requires auth but user is not authenticated
        next('/login')
    } else if (to.path === '/login' && isAuthenticated) {
        // User is authenticated but trying to access login page
        next('/dashboard')
    } else {
        // Allow navigation
        next()
    }
})

// Create Pinia store
const pinia = createPinia()

// Create and mount app
createApp(App)
    .use(pinia)
    .use(router)
    .mount('#app')