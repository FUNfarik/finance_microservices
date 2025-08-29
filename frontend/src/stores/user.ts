import { defineStore } from 'pinia'
import portfolioService from '../services/portfolioService.js'
import authService from '../services/authService.js'
import { marketService } from '../services/marketService.js'

// Types
interface User {
    id: number        // backend returns integer ID
    email: string
    name?: string
}

interface Holding {
    symbol: string
    shares: number
    avg_price: number
    current_price?: number
    market_value?: number
    gain_loss?: number
}

interface Portfolio {
    holdings: Holding[]
    total_value: number
    cash: number
    total_gain_loss: number
    total_gain_loss_percent: number
}

interface LoadingState {
    portfolio: boolean
    user: boolean
    trade: boolean
}

interface ErrorState {
    portfolio: string | null
    user: string | null
    trade: string | null
}

interface TradeData {
    symbol: string
    shares: number
    price: number
    trade_type: 'buy' | 'sell'
}

interface UserState {
    user: User | null
    isAuthenticated: boolean
    portfolio: Portfolio
    loading: LoadingState
    errors: ErrorState
    lastUpdate: Date | null
}

export const useUserStore = defineStore('user', {
    state: (): UserState => ({
        user: null,
        isAuthenticated: false,

        portfolio: {
            holdings: [],
            total_value: 0,
            cash: 10000.00,
            total_gain_loss: 0,
            total_gain_loss_percent: 0
        },

        loading: {
            portfolio: false,
            user: false,
            trade: false
        },

        errors: {
            portfolio: null,
            user: null,
            trade: null
        },

        lastUpdate: null
    }),

    getters: {
        totalNetWorth: (state): number => {
            return state.portfolio.total_value + state.portfolio.cash
        },

        holdingsMap: (state): Record<string, Holding> => {
            const map: Record<string, Holding> = {}
            state.portfolio.holdings.forEach(holding => {
                map[holding.symbol] = holding
            })
            return map
        },

        canAfford: (state) => (amount: number): boolean => {
            return state.portfolio.cash >= amount
        },

        canSell: (state) => (symbol: string, shares: number): boolean => {
            const holding = (state as any).holdingsMap[symbol] as Holding | undefined
            return holding ? holding.shares >= shares : false
        },

        getShares: (state) => (symbol: string): number => {
            const holding = (state as any).holdingsMap[symbol] as Holding | undefined
            return holding ? holding.shares : 0
        }
    },

    actions: {
        async login(email: string, password: string) {
            try {
                this.loading.user = true
                this.errors.user = null

                const response = await authService.login(email, password)

                this.user = response.user
                this.isAuthenticated = true

                if (response.user?.id) {
                    // persist user_id for portfolio service
                    localStorage.setItem('user_id', String(response.user.id))
                    portfolioService.setUserId(response.user.id)
                }

                await this.loadPortfolio()

                return response
            } catch (error: any) {
                this.errors.user = error.message
                throw error
            } finally {
                this.loading.user = false
            }
        },

        logout(): void {
            this.user = null
            this.isAuthenticated = false
            this.portfolio = {
                holdings: [],
                total_value: 0,
                cash: 10000.00,
                total_gain_loss: 0,
                total_gain_loss_percent: 0
            }
            portfolioService.clearUser()
            localStorage.removeItem('finance_token')
        },

        async loadPortfolio(): Promise<void> {
            try {
                this.loading.portfolio = true
                this.errors.portfolio = null

                console.log('Loading portfolio from Go API...')
                const portfolio = await portfolioService.getPortfolio()
                console.log('Portfolio data received:', portfolio)

                if (portfolio) {
                    this.portfolio = {
                        holdings: portfolio.holdings ?? [],
                        total_value: portfolio.total_value ?? 0,
                        cash: portfolio.cash ?? 0,
                        total_gain_loss: portfolio.total_gain_loss ?? 0,
                        total_gain_loss_percent: portfolio.total_gain_loss_percent ?? 0
                    }
                }

                this.lastUpdate = new Date()

                if (this.portfolio.holdings.length > 0) {
                    await this.updateHoldingsPrices()
                }

                console.log('Portfolio loaded successfully:', this.portfolio)

            } catch (error: any) {
                console.error('Failed to load portfolio:', error)

                if (error.code === 'ERR_NETWORK') {
                    this.errors.portfolio = 'Portfolio service not running on port 8003'
                } else if (error.response?.status === 404) {
                    console.log('New user - using default portfolio')
                    this.portfolio = {
                        holdings: [],
                        total_value: 0,
                        cash: 10000.00,
                        total_gain_loss: 0,
                        total_gain_loss_percent: 0
                    }
                    this.errors.portfolio = null
                } else {
                    this.errors.portfolio = error.message || 'Failed to load portfolio'
                    this.portfolio = {
                        holdings: [],
                        total_value: 0,
                        cash: 10000.00,
                        total_gain_loss: 0,
                        total_gain_loss_percent: 0
                    }
                }
            } finally {
                this.loading.portfolio = false
            }
        },

        async updateHoldingsPrices(): Promise<void> {
            if (this.portfolio.holdings.length === 0) return

            try {
                const symbols = this.portfolio.holdings.map(h => h.symbol)
                let prices: Record<string, number> = {}

                for (const symbol of symbols) {
                    try {
                        const response = await fetch(`http://localhost:8002/stock/${symbol}`)
                        if (response.ok) {
                            const stock = await response.json()
                            prices[symbol] = stock.price
                        }
                    } catch (err) {
                        console.warn(`Failed to get price for ${symbol}:`, err)
                    }
                }

                this.portfolio.holdings = this.portfolio.holdings.map(holding => {
                    const currentPrice = prices[holding.symbol]
                    if (currentPrice) {
                        const marketValue = holding.shares * currentPrice
                        const gainLoss = marketValue - (holding.shares * holding.avg_price)

                        return {
                            ...holding,
                            current_price: currentPrice,
                            market_value: marketValue,
                            gain_loss: gainLoss
                        }
                    }
                    return holding
                })

                this.recalculatePortfolioValue()

            } catch (error) {
                console.error('Failed to update holdings prices:', error)
            }
        },

        recalculatePortfolioValue(): void {
            const totalMarketValue = this.portfolio.holdings.reduce((sum, holding) => {
                return sum + (holding.market_value || 0)
            }, 0)

            const totalCost = this.portfolio.holdings.reduce((sum, holding) => {
                return sum + (holding.shares * holding.avg_price)
            }, 0)

            this.portfolio.total_value = totalMarketValue
            this.portfolio.total_gain_loss = totalMarketValue - totalCost
            this.portfolio.total_gain_loss_percent = totalCost > 0
                ? (this.portfolio.total_gain_loss / totalCost) * 100
                : 0
        },

        async executeTrade(tradeData: TradeData) {
            try {
                this.loading.trade = true
                this.errors.trade = null

                const { symbol, shares, price, trade_type } = tradeData
                console.log('Executing trade:', tradeData)

                let result
                if (trade_type === 'buy') {
                    result = await portfolioService.buyStock(symbol, shares)
                } else {
                    result = await portfolioService.sellStock(symbol, shares)
                }

                console.log('Trade result:', result)

                this.updatePortfolioAfterTrade(tradeData)

                setTimeout(() => {
                    this.loadPortfolio().catch(err =>
                        console.log('Background refresh failed:', err)
                    )
                }, 1000)

                return result
            } catch (error: any) {
                console.error('Trade execution failed:', error)
                this.errors.trade = error.message || 'Trade failed'
                throw error
            } finally {
                this.loading.trade = false
            }
        },

        updatePortfolioAfterTrade(tradeData: TradeData): void {
            const { symbol, shares, price, trade_type } = tradeData
            const totalAmount = shares * price

            if (trade_type === 'buy') {
                this.portfolio.cash -= totalAmount

                const existingHoldingIndex = this.portfolio.holdings.findIndex(h => h.symbol === symbol)
                if (existingHoldingIndex >= 0) {
                    const holding = this.portfolio.holdings[existingHoldingIndex]
                    const newShares = holding.shares + shares
                    const newAvgPrice = ((holding.shares * holding.avg_price) + totalAmount) / newShares

                    this.portfolio.holdings[existingHoldingIndex] = {
                        ...holding,
                        shares: newShares,
                        avg_price: newAvgPrice,
                        market_value: newShares * (holding.current_price || price)
                    }
                } else {
                    this.portfolio.holdings.push({
                        symbol,
                        shares,
                        avg_price: price,
                        current_price: price,
                        market_value: totalAmount,
                        gain_loss: 0
                    })
                }
            } else {
                this.portfolio.cash += totalAmount

                const holdingIndex = this.portfolio.holdings.findIndex(h => h.symbol === symbol)
                if (holdingIndex >= 0) {
                    const holding = this.portfolio.holdings[holdingIndex]
                    const newShares = holding.shares - shares

                    if (newShares <= 0) {
                        this.portfolio.holdings.splice(holdingIndex, 1)
                    } else {
                        this.portfolio.holdings[holdingIndex] = {
                            ...holding,
                            shares: newShares,
                            market_value: newShares * (holding.current_price || price)
                        }
                    }
                }
            }

            this.recalculatePortfolioValue()
        },

        clearError(type: keyof ErrorState): void {
            if (this.errors[type]) {
                this.errors[type] = null
            }
        },

        async refreshPortfolio(): Promise<void> {
            await this.loadPortfolio()
        },

        initDemoUser(): void {
            if (!localStorage.getItem('user_id')) {
                localStorage.setItem('user_id', 'demo-user')
            }
            this.isAuthenticated = true
        }
    }
})