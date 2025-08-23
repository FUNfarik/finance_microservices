import { portfolioApi } from './api.js'

class PortfolioService {
    async getPortfolio() {
        try {
            const response = await portfolioApi.get('/portfolio')
            return response.data
        } catch (error) {
            console.error('Failed to fetch portfolio:', error)
            throw error
        }
    }

    async getPortfolioValue() {
        try {
            const response = await portfolioApi.get('/portfolio/value')
            return response.data
        } catch (error) {
            console.error('Failed to get portfolio value:', error)
            throw error
        }
    }

    async buyStock(symbol, shares) {
        try {
            const response = await portfolioApi.post('/portfolio/buy', {
                symbol,
                shares: parseInt(shares),
                transaction_type: 'BUY'
            })
            return response.data
        } catch (error) {
            console.error('Failed to buy stock:', error)
            throw error
        }
    }

    async sellStock(symbol, shares) {
        try {
            const response = await portfolioApi.post('/portfolio/sell', {
                symbol,
                shares: parseInt(shares),
                transaction_type: 'SELL'
            })
            return response.data
        } catch (error) {
            console.error('Failed to sell stock:', error)
            throw error
        }
    }

    async getTransactionsHistory() {
        try {
            const response = await portfolioApi.get('/portfolio/transactions')
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
}

const portfolioService = new PortfolioService()
export default portfolioService