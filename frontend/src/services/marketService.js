import { marketApi } from './api.js'

class MarketService {
    async getStockPrice(symbol) {
        try {
            const response = await marketApi.get(`/stock/${symbol}`)
            return response.data
        } catch (error) {
            console.error(`Failed to get price for ${symbol}:`, error)
            throw error
        }
    }

    async getMultipleStockPrices(symbols) {
        try {
            // Your Rust service supports batch requests
            const response = await marketApi.post('/stocks/batch', {
                symbols: symbols
            })
            return response.data
        } catch (error) {
            console.error('Failed to get multiple stock prices:', error)
            throw error
        }
    }

    async searchStocks(query) {
        try {
            const response = await marketApi.get(`/search?q=${encodeURIComponent(query)}`)
            return response.data
        } catch (error) {
            console.error('Failed to search stocks:', error)
            throw error
        }
    }

    async getPopularStocks() {
        try {
            const response = await marketApi.get('/popular')
            return response.data
        } catch (error) {
            console.error('Failed to get popular stocks:', error)
            throw error
        }
    }

    // Test connection to market data service
    async testConnection() {
        try {
            const response = await marketApi.get('/health')
            return response.data
        } catch (error) {
            console.error('Market service connection failed:', error)
            throw error
        }
    }
}

const marketService = new MarketService()
export { marketService }