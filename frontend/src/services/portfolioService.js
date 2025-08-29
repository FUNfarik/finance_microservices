import { portfolioApi } from './api.js'

class PortfolioService {
    getUserId() {
        // Always parse back to number, fallback to 1
        const raw = localStorage.getItem('user_id')
        return raw ? parseInt(raw, 10) : 1
    }

    async getPortfolio() {
        try {
            const userId = this.getUserId()
            console.log('Fetching portfolio for user:', userId)

            const response = await portfolioApi.get(`/portfolio/${userId}`)
            console.log('Portfolio API response:', response.data)

            // unwrap { status, message, data }
            const body = response.data
            const payload = body?.data ?? body

            // normalize shape
            return {
                holdings: Array.isArray(payload.holdings) ? payload.holdings : [],
                total_value: payload.total_value ?? 0,
                cash: payload.cash ?? 0,
                total_gain_loss: payload.total_gain_loss ?? 0,
                total_gain_loss_percent: payload.total_gain_loss_percent ?? 0
            }
        } catch (error) {
            console.error('Failed to fetch portfolio:', error)

            if (error.response?.status === 404) {
                // New user – return empty portfolio
                return {
                    holdings: [],
                    total_value: 0,
                    cash: 10000.00,
                    total_gain_loss: 0,
                    total_gain_loss_percent: 0
                }
            }
            throw error
        }
    }

    async getPortfolioValue() {
        const portfolio = await this.getPortfolio()
        return {
            total_value: portfolio.total_value ?? 0,
            cash: portfolio.cash ?? 0
        }
    }

    async buyStock(symbol, shares) {
        try {
            const userId = this.getUserId()
            console.log('Buying stock:', { userId, symbol, shares })

            const response = await portfolioApi.post('/buy', {
                user_id: userId,
                symbol: symbol.toUpperCase(),
                shares: parseInt(shares, 10)
            })

            console.log('Buy response:', response.data)
            return response.data
        } catch (error) {
            console.error('Failed to buy stock:', error)
            throw error
        }
    }

    async sellStock(symbol, shares) {
        try {
            const userId = this.getUserId()
            console.log('Selling stock:', { userId, symbol, shares })

            const response = await portfolioApi.post('/sell', {
                user_id: userId,
                symbol: symbol.toUpperCase(),
                shares: parseInt(shares, 10)
            })

            console.log('Sell response:', response.data)
            return response.data
        } catch (error) {
            console.error('Failed to sell stock:', error)
            throw error
        }
    }

    async getTransactionsHistory() {
        try {
            const userId = this.getUserId()
            const response = await portfolioApi.get(`/transactions/${userId}`)
            // unwrap here too if backend returns {status,message,data}
            return response.data?.data ?? response.data
        } catch (error) {
            console.error('Failed to get transactions history:', error)
            throw error
        }
    }

    async testConnection() {
        try {
            const response = await portfolioApi.get('/health')
            return response.data
        } catch (error) {
            console.error('Failed to connect to Portfolio service:', error)
            throw error
        }
    }

    // Set user ID (for when user logs in)
    setUserId(userId) {
        localStorage.setItem('user_id', String(userId))
    }

    // Clear user data (for logout)
    clearUser() {
        localStorage.removeItem('user_id')
    }
}

const portfolioService = new PortfolioService()
export default portfolioService