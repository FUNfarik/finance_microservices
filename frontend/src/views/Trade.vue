<template>
  <div class="trade-page">
    <h1>🔄 Trade Stocks</h1>

    <!-- Stock Search Section -->
    <div class="search-section">
      <h2>Stock Lookup</h2>
      <div class="search-box">
        <input
            v-model="stockSymbol"
            type="text"
            placeholder="Enter stock symbol (e.g., AAPL)"
            @input="onSymbolChange"
            class="search-input"
        />
        <button @click="lookupStock" :disabled="!stockSymbol || loading" class="search-btn">
          {{ loading ? 'Loading...' : 'Get Quote' }}
        </button>
      </div>

      <!-- Display stock info when found -->
      <div v-if="stockData" class="stock-card">
        <div class="stock-header">
          <div class="stock-symbol">{{ stockData.symbol }}</div>
        </div>
        <div class="stock-name">{{ stockData.name }}</div>
        <div class="stock-price">${{ stockData.price.toFixed(2) }}</div>
        <div class="stock-change" :class="changeClass">
          {{ stockData.change > 0 ? '+' : '' }}{{ stockData.change.toFixed(2) }}
          ({{ stockData.changePercent > 0 ? '+' : '' }}{{ stockData.changePercent.toFixed(2) }}%)
        </div>
      </div>

      <!-- Error message -->
      <div v-if="error" class="error">
        {{ error }}
      </div>
    </div>

    <!-- Trading Form Section -->
    <div v-if="stockData" class="search-section">
      <h2>Place Order</h2>

      <div class="trade-type">
        <label>
          <input type="radio" v-model="tradeType" value="buy" />
          Buy
        </label>
        <label>
          <input type="radio" v-model="tradeType" value="sell" />
          Sell
        </label>
      </div>

      <div class="trade-controls">
        <div class="quantity-control">
          <button @click="decreaseShares" class="qty-btn">-</button>
          <input
              v-model.number="shares"
              type="number"
              min="1"
              class="qty-input"
          />
          <button @click="increaseShares" class="qty-btn">+</button>
        </div>

        <!-- Order Summary -->
        <div v-if="shares > 0" class="order-summary">
          <h3>Order Summary</h3>
          <p><strong>Action:</strong> {{ tradeType.toUpperCase() }} {{ shares }} shares of {{ stockData.symbol }}</p>
          <p><strong>Price per share:</strong> ${{ stockData.price.toFixed(2) }}</p>
          <p><strong>Total {{ tradeType === 'buy' ? 'Cost' : 'Proceeds' }}:</strong>
            ${{ totalAmount.toFixed(2) }}
          </p>

          <!-- Available cash check for buying -->
          <div v-if="tradeType === 'buy'" class="cash-check">
            <p><strong>Available Cash:</strong> ${{ userCash.toFixed(2) }}</p>
            <p v-if="totalAmount > userCash" class="insufficient-funds">
              ⚠️ Insufficient funds
            </p>
          </div>

          <!-- Holdings check for selling -->
          <div v-if="tradeType === 'sell'" class="holdings-check">
            <p><strong>Current Holdings:</strong> {{ currentShares }} shares</p>
            <p v-if="shares > currentShares" class="insufficient-shares">
              ⚠️ Insufficient shares
            </p>
          </div>
        </div>

        <div class="trade-buttons">
          <button
              @click="executeTrade"
              :disabled="!canTrade || executing"
              :class="tradeType === 'buy' ? 'buy-btn' : 'sell-btn'"
          >
            {{ executing ? 'Processing...' : `${tradeType.toUpperCase()} ${shares} Shares` }}
          </button>
        </div>
      </div>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="transaction-result" @click="clearMessage">
      <div class="result-card success">
        <h3>Trade Successful!</h3>
        <p>{{ successMessage }}</p>
        <button @click="clearMessage" class="close-btn">Close</button>
      </div>
    </div>

    <div class="navigation">
      <router-link to="/dashboard" class="nav-button">← Back to Dashboard</router-link>
    </div>
  </div>
</template>

<script>
// Import your service files using correct relative paths
import { marketService } from '../services/marketService.js'
import portfolioService from '../services/portfolioService.js'
import authService from '../services/authService.js'

export default {
  name: 'Trade',
  data() {
    return {
      // Stock lookup
      stockSymbol: '',
      stockData: null,
      loading: false,
      error: null,

      // Trading form
      tradeType: 'buy',
      shares: 1,
      executing: false,
      successMessage: '',

      // User data (will come from API)
      userCash: 10000.00,
      userHoldings: {} // Will store symbol: shares pairs
    }
  },
  computed: {
    // Calculate total cost/proceeds
    totalAmount() {
      if (!this.stockData || !this.shares) return 0
      return this.stockData.price * this.shares
    },

    // Check if stock price went up or down
    changeClass() {
      if (!this.stockData) return ''
      return this.stockData.change >= 0 ? 'positive' : 'negative'
    },

    // Current shares of this stock
    currentShares() {
      return this.userHoldings[this.stockData?.symbol] || 0
    },

    // Can the user execute this trade?
    canTrade() {
      if (!this.stockData || !this.shares || this.shares <= 0) return false

      if (this.tradeType === 'buy') {
        return this.totalAmount <= this.userCash
      } else {
        return this.shares <= this.currentShares
      }
    }
  },
  async mounted() {
    // Load user data when component mounts
    await this.loadUserData()
  },
  methods: {
    // Load user portfolio and cash
    async loadUserData() {
      try {
        const portfolioData = await portfolioService.getPortfolio()
        this.userCash = portfolioData.cash || 10000.00
        this.userHoldings = portfolioData.holdings || {}

        console.log('User data loaded (mock)')
      } catch (err) {
        console.error('Failed to load user data:', err)
        // Keep default values
      }
    },

    // Clear previous data when symbol changes
    onSymbolChange() {
      this.stockData = null
      this.error = null
      this.successMessage = ''
    },

    // Look up stock information using market service
    async lookupStock() {
      if (!this.stockSymbol) return

      this.loading = true
      this.error = null

      try {
        // Use your market service - getStockPrice returns full quote data
        const data = await marketService.getStockPrice(this.stockSymbol.toUpperCase())

        this.stockData = {
          symbol: data.symbol,
          name: data.name || data.company_name || 'Unknown Company',
          price: data.price || data.current_price,
          change: data.change || data.price_change || 0,
          changePercent: data.change_percent || data.percentage_change || 0
        }

      } catch (err) {
        this.error = err.message || 'Failed to fetch stock data'
        this.stockData = null
      } finally {
        this.loading = false
      }
    },

    // Execute the trade
    async executeTrade() {
      if (!this.canTrade) return

      this.executing = true
      this.error = null

      try {
        const tradeData = {
           symbol: this.stockData.symbol,
           shares: this.shares,
           price: this.stockData.price,
           trade_type: this.tradeType
        }
         const result = await portfolioService.executeTrade(tradeData)

        // Simulate trade success
        if (this.tradeType === 'buy') {
          this.userCash -= this.totalAmount
          this.userHoldings[this.stockData.symbol] = (this.userHoldings[this.stockData.symbol] || 0) + this.shares
        } else {
          this.userCash += this.totalAmount
          this.userHoldings[this.stockData.symbol] -= this.shares
        }

        this.successMessage = `Successfully ${this.tradeType === 'buy' ? 'bought' : 'sold'} ${this.shares} shares of ${this.stockData.symbol}`

        // Reset form
        this.shares = 1

      } catch (err) {
        this.error = err.message || 'Trade execution failed'
      } finally {
        this.executing = false
      }
    },

    // Quantity control methods
    increaseShares() {
      this.shares += 1
    },

    decreaseShares() {
      if (this.shares > 1) {
        this.shares -= 1
      }
    },

    // Clear success message
    clearMessage() {
      this.successMessage = ''
    }
  }
}
</script>

<style scoped>
@import '../styles/components/trade.css';

/* Additional styles for elements not in your CSS */
.trade-type {
  display: flex;
  gap: 20px;
  margin: 20px 0;
}

.trade-type label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 1.1rem;
}

.order-summary {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  margin: 20px 0;
  border: 1px solid #e9ecef;
}

.order-summary h3 {
  margin-top: 0;
  color: #333;
}

.cash-check, .holdings-check {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #dee2e6;
}

.insufficient-funds, .insufficient-shares {
  color: #dc3545;
  font-weight: 600;
}

.error {
  background: #f8d7da;
  color: #721c24;
  padding: 15px;
  border-radius: 8px;
  border: 1px solid #f5c6cb;
  margin: 15px 0;
}
</style>