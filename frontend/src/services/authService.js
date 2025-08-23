import { authApi, setAuthToken } from './api.js'

class AuthService {
    async testConnection() {
        try {
            const response = await authApi.get('/health')
            return response.data
        } catch (error) {
            console.error('Auth service connection failed:', error)
            throw error
        }
    }

    async login(email, password) {
        try {
            const response = await authApi.post('/login', {
                email,
                password
            })

            if (response.data.token) {
                setAuthToken(response.data.token)
            }

            return response.data
        } catch (error) {
            console.error('Login failed:', error)
            throw error
        }
    }

    isAuthenticated() {
        const token = localStorage.getItem('finance_token')
        return !!token
    }
}

const authService = new AuthService()
export default authService