import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'

// Import global styles
import './styles/main.css'

// Import views
import Login from './views/Login.vue'
import Dashboard from './views/Dashboard.vue'
import Trade from './views/Trade.vue'

// Define routes
const routes = [
    { path: '/', redirect: '/dashboard' },
    { path: '/login', name: 'Login', component: Login },
    { path: '/dashboard', name: 'Dashboard', component: Dashboard },
    { path: '/trade', name: 'Trade', component: Trade }
]

// Create router
const router = createRouter({
    history: createWebHistory(),
    routes
})

// Create and mount app
createApp(App)
    .use(router)
    .mount('#app')