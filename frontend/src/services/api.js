import axios from 'axios'

// Create axios instances for different services using nginx proxy routes
const authApi = axios.create({
    baseURL: '/api/auth',  // Routes through nginx to auth:8001
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

const portfolioApi = axios.create({
    baseURL: '/api/portfolio',  // Routes through nginx to portfolio:8003
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

const marketApi = axios.create({
    baseURL: '/api/market',  // Routes through nginx to make-data-service:8002
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

// Token management
let authToken = null

const setAuthToken = (token) => {
    authToken = token

    const bearer = token ? `Bearer ${token}` : null
    portfolioApi.defaults.headers.common['Authorization'] = bearer
    authApi.defaults.headers.common['Authorization'] = bearer

    if (token) {
        localStorage.setItem('finance_token', token)
    } else {
        localStorage.removeItem('finance_token')
    }
}

// Add auth interceptor to all APIs
const addAuthInterceptor = (apiInstance) => {
    apiInstance.interceptors.request.use(
        (config) => {
            const token = localStorage.getItem('finance_token') ||
                localStorage.getItem('token') ||
                localStorage.getItem('jwt_token') ||
                localStorage.getItem('access_token')

            if (token) {
                config.headers.Authorization = `Bearer ${token}`
            }

            return config
        },
        (error) => {
            return Promise.reject(error)
        }
    )

    // Add response interceptor for better error handling
    apiInstance.interceptors.response.use(
        (response) => response,
        (error) => {
            if (error.response?.status === 401) {
                // Token expired or invalid
                console.error('Authentication failed - redirecting to login')
                localStorage.removeItem('finance_token')
                localStorage.removeItem('token')
                localStorage.removeItem('jwt_token')
                localStorage.removeItem('access_token')

                // Redirect to login if not already there
                if (!window.location.pathname.includes('/login')) {
                    window.location.href = '/login'
                }
            }
            return Promise.reject(error)
        }
    )
}

// Apply interceptors to all API instances
addAuthInterceptor(authApi)
addAuthInterceptor(portfolioApi)
addAuthInterceptor(marketApi)

// Load stored token on app startup
const loadStoredToken = () => {
    const token = localStorage.getItem('finance_token')
    if (token) {
        setAuthToken(token)
    }
    return token
}

export default portfolioApi
export { authApi, portfolioApi, marketApi, setAuthToken, loadStoredToken }