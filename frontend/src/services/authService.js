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

    async register(userData) {
        try {
            const response = await authApi.post('/register', {
                username: `${userData.firstName} ${userData.lastName}`.trim(), // Combine first and last name
                email: userData.email,
                password: userData.password
            })

            return response.data
        } catch (error) {
            console.error('Registration failed:', error)

            // Handle different error types based on your Go backend
            if (error.response?.status === 400) {
                throw new Error(error.response.data.error || 'Registration failed - please check your information')
            } else if (error.response?.status === 409) {
                throw new Error('Username or email already exists')
            } else {
                throw new Error('Registration failed - please try again')
            }
        }
    }

    async login(email, password) {
        try {
            const response = await authApi.post('/login', {
                email,
                password
            })

            // Store the token if provided
            if (response.data.token) {
                setAuthToken(response.data.token)
                localStorage.setItem('finance_token', response.data.token)

                // Extract user_id from JWT token and store it
                try {
                    const payload = JSON.parse(atob(response.data.token.split('.')[1]))
                    if (payload.user_id) {
                        localStorage.setItem('user_id', String(payload.user_id))
                        console.log('Stored user_id:', payload.user_id)
                    }
                } catch (error) {
                    console.error('Failed to extract user_id from token:', error)
                }

                // Store user info if provided
                if (response.data.user) {
                    localStorage.setItem('finance_user', JSON.stringify(response.data.user))

                    // Also store user_id from user object if available
                    if (response.data.user.id && !localStorage.getItem('user_id')) {
                        localStorage.setItem('user_id', String(response.data.user.id))
                    }
                }
            }

            return response.data
        } catch (error) {
            console.error('Login failed:', error)

            // Handle different error types based on your Go backend
            if (error.response?.status === 401) {
                throw new Error('Invalid email or password')
            } else if (error.response?.status === 400) {
                throw new Error(error.response.data.error || 'Please check your email and password')
            } else {
                throw new Error('Login failed - please try again')
            }
        }
    }

    logout() {
        // Clear all stored auth data
        localStorage.removeItem('finance_token')
        localStorage.removeItem('finance_user')
        localStorage.removeItem('user_id') // Also clear user_id
        setAuthToken(null)
    }

    isAuthenticated() {
        const token = localStorage.getItem('finance_token')
        if (!token) return false

        try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            const currentTime = Date.now() / 1000

            if (payload.exp < currentTime) {
                this.logout() // Clear expired token
                return false
            }

            return true
        } catch (error) {
            this.logout() // Clear invalid token
            return false
        }
    }

    getCurrentUser() {
        const userStr = localStorage.getItem('finance_user')
        return userStr ? JSON.parse(userStr) : null
    }

    getCurrentUserId() {
        // First try to get from localStorage
        const storedUserId = localStorage.getItem('user_id')
        if (storedUserId) {
            return parseInt(storedUserId, 10)
        }

        // If not found, try to extract from token
        const token = localStorage.getItem('finance_token')
        if (token) {
            try {
                const payload = JSON.parse(atob(token.split('.')[1]))
                if (payload.user_id) {
                    localStorage.setItem('user_id', String(payload.user_id))
                    return payload.user_id
                }
            } catch (error) {
                console.error('Failed to extract user_id from token:', error)
            }
        }

        return null
    }

    getToken() {
        return localStorage.getItem('finance_token')
    }
}

const authService = new AuthService()
export default authService