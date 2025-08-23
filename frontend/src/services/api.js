import axios from 'axios'

// Create axios instances for different services
const authApi = axios.create({
    baseURL: 'http://localhost:8001',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

const portfolioApi = axios.create({
    baseURL: 'http://localhost:8003',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

const marketApi = axios.create({
    baseURL: 'http://localhost:8002',
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

    if (token) {
        localStorage.setItem('finance_token', token)
    } else {
        localStorage.removeItem('finance_token')
    }
}

const loadStoredToken = () => {
    const token = localStorage.getItem('finance_token')
    if (token) {
        setAuthToken(token)
    }
    return token
}

export { authApi, portfolioApi, marketApi, setAuthToken, loadStoredToken }