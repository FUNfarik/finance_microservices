import { portfolioApi } from './api.js'

class PortfolioService {
    getUserId() {
        return parseInt(localStorage.getItem('user_id')) || 1  // Use integer 1 instead of "demo-user"
    }

    async getPortfolio() {
        try {
            const userId = this.getUserId()
            console.log('Fetching portfolio for user:', userId)

            const response = await portfolioApi.get(`/portfolio/${userId}`)
            console.log('Portfolio API response:', response.data)
            return response.data
        } catch (error) {
            console.error('Failed to fetch portfolio:', error)

            // If user doesn't exist (404), return default empty portfolio
            if (error.response?.status === 404) {
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
        try {
            const portfolio = await this.getPortfolio()
            return {
                total_value: portfolio.total_value || 0,
                cash: portfolio.cash || 0
            }
        } catch (error) {
            console.error('Failed to get portfolio value:', error)
            throw error
        }
    }

    async buyStock(symbol, shares) {
        try {
            const userId = this.getUserId()
            console.log('Buying stock:', { userId, symbol, shares })

            const response = await portfolioApi.post('/buy', {
                user_id: userId,
                symbol: symbol.toUpperCase(),
                shares: parseInt(shares)
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
                shares: parseInt(shares)
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
            return response.data
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
        localStorage.setItem('user_id', userId)
    }

    // Clear user data (for logout)
    clearUser() {
        localStorage.removeItem('user_id')
    }
}

const portfolioService = new PortfolioService()
export default portfolioService