<template>
  <div class="trade-page">
    <h1>Trade Stocks</h1>

    <!-- Quick Portfolio Summary -->
    <div class="portfolio-summary">
      <div class="summary-cards">
        <div class="summary-card">
          <h3>Available Cash</h3>
          <p class="cash-amount">${{ formatCurrency(portfolio.cash) }}</p>
        </div>
        <div class="summary-card">
          <h3>Portfolio Value</h3>
          <p class="portfolio-value">${{ formatCurrency(portfolio.total_value) }}</p>
        </div>
        <div class="summary-card">
          <h3>Holdings</h3>
          <p class="holdings-count">{{ portfolio.holdings?.length || 0 }}</p>
        </div>
      </div>
    </div>

    <!-- Stock Search Section -->
    <div class="search-section">
      <h2>📈 Stock Lookup</h2>
      <div class="search-box">
        <input
            v-model="stockSymbol"
            type="text"
            placeholder="Enter stock symbol (e.g., AAPL, MSFT, GOOGL)"
            @input="onSymbolChange"
            @keyup.enter="lookupStock"
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
          {{ stockData.changePercent > 0 ? '+' : '' }}{{ stockData.changePercent.toFixed(2) }}%
        </div>

        <!-- Current Holdings Display -->
        <div v-if="currentShares > 0" class="current-holding">
          <span class="holding-text">You own {{ currentShares }} shares</span>
        </div>
      </div>

      <!-- Error message -->
      <div v-if="error" class="error">
        {{ error }}
      </div>
    </div>

    <!-- Trading Form Section -->
    <div v-if="stockData" class="search-section">
      <h2>📋 Place Order</h2>

      <!-- Trade Type Selection -->
      <div class="trade-type">
        <label>
          <input type="radio" v-model="tradeType" value="buy" />
          <span class="radio-label buy-label">Buy</span>
        </label>
        <label>
          <input type="radio" v-model="tradeType" value="sell" />
          <span class="radio-label sell-label">Sell</span>
        </label>
      </div>

      <div class="trade-controls">
        <!-- Quantity Control -->
        <div class="quantity-control">
          <button @click="decreaseShares" class="qty-btn" :disabled="shares <= 1">-</button>
          <input
              v-model.number="shares"
              type="number"
              min="1"
              :max="tradeType === 'sell' ? currentShares : 9999"
              class="qty-input"
          />
          <button @click="increaseShares" class="qty-btn">+</button>
        </div>

        <!-- Quick quantity buttons -->
        <div class="quick-qty">
          <button v-if="tradeType === 'sell' && currentShares > 0"
                  @click="shares = Math.floor(currentShares / 4)"
                  class="quick-btn">25%</button>
          <button v-if="tradeType === 'sell' && currentShares > 0"
                  @click="shares = Math.floor(currentShares / 2)"
                  class="quick-btn">50%</button>
          <button v-if="tradeType === 'sell' && currentShares > 0"
                  @click="shares = currentShares"
                  class="quick-btn">All</button>

          <button v-if="tradeType === 'buy'"
                  @click="setMaxBuyShares"
                  class="quick-btn">Max</button>
        </div>

        <!-- Order Summary -->
        <div v-if="shares > 0" class="order-summary">
          <h3>Order Summary</h3>
          <div class="summary-row">
            <span>Action:</span>
            <span><strong>{{ tradeType.toUpperCase() }} {{ shares }} shares of {{ stockData.symbol }}</strong></span>
          </div>
          <div class="summary-row">
            <span>Price per share:</span>
            <span>${{ stockData.price.toFixed(2) }}</span>
          </div>
          <div class="summary-row total-row">
            <span>Total {{ tradeType === 'buy' ? 'Cost' : 'Proceeds' }}:</span>
            <span class="total-amount">${{ totalAmount.toFixed(2) }}</span>
          </div>

          <!-- Available cash check for buying -->
          <div v-if="tradeType === 'buy'" class="validation-check">
            <div class="check-row">
              <span>Available Cash:</span>
              <span>${{ formatCurrency(portfolio.cash) }}</span>
            </div>
            <div v-if="totalAmount > portfolio.cash" class="insufficient-funds">
              ⚠️ Insufficient funds (need ${{ formatCurrency(totalAmount - portfolio.cash) }} more)
            </div>
          </div>

          <!-- Holdings check for selling -->
          <div v-if="tradeType === 'sell'" class="validation-check">
            <div class="check-row">
              <span>Current Holdings:</span>
              <span>{{ currentShares }} shares</span>
            </div>
            <div v-if="shares > currentShares" class="insufficient-shares">
              ⚠️ Insufficient shares (you only own {{ currentShares }} shares)
            </div>
          </div>
        </div>

        <!-- Trade Button -->
        <div class="trade-buttons">
          <button
              @click="executeTrade"
              :disabled="!canTrade || executing"
              :class="['trade-btn-main', tradeType === 'buy' ? 'buy-btn' : 'sell-btn']"
          >
            <span v-if="executing" class="loading-spinner">⏳</span>
            {{ executing ? 'Processing...' : `${tradeType.toUpperCase()} ${shares} Shares` }}
          </button>
        </div>
      </div>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="transaction-result" @click="clearMessage">
      <div class="result-card success">
        <div class="success-icon">✅</div>
        <h3>Trade Successful!</h3>
        <p>{{ successMessage }}</p>
        <div class="result-actions">
          <button @click="clearMessage" class="close-btn">Continue Trading</button>
          <router-link to="/dashboard" class="dashboard-btn">View Dashboard</router-link>
        </div>
      </div>
    </div>

    <!-- Holdings Summary -->
    <div v-if="portfolio.holdings?.length > 0" class="holdings-section">
      <h2>📊 Your Holdings</h2>
      <div class="holdings-grid">
        <div v-for="holding in portfolio.holdings" :key="holding.symbol" class="holding-card">
          <div class="holding-header">
            <span class="holding-symbol">{{ holding.symbol }}</span>
            <span class="holding-shares">{{ holding.shares }} shares</span>
          </div>
          <div class="holding-details">
            <div class="holding-row">
              <span>Avg Price:</span>
              <span>${{ formatCurrency(holding.avg_price) }}</span>
            </div>
            <div class="holding-row">
              <span>Current:</span>
              <span>${{ formatCurrency(holding.current_price || 0) }}</span>
            </div>
            <div class="holding-row">
              <span>Value:</span>
              <span>${{ formatCurrency(holding.market_value || 0) }}</span>
            </div>
          </div>
          <div class="holding-actions">
            <button @click="quickTrade(holding.symbol, 'buy')" class="quick-action-btn buy">Buy More</button>
            <button @click="quickTrade(holding.symbol, 'sell')" class="quick-action-btn sell">Sell</button>
          </div>
        </div>
      </div>
    </div>

    <div class="navigation">
      <router-link to="/dashboard" class="nav-button">← Back to Dashboard</router-link>
    </div>
  </div>
</template>

<script>
import { useUserStore } from '../stores/user'

export default {
  name: 'Trade',
  setup() {
    const userStore = useUserStore()
    return { userStore }
  },

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
      successMessage: ''
    }
  },

  computed: {
    portfolio() { return this.userStore.portfolio },

    // Calculate total cost/proceeds
    totalAmount() {
      if (!this.stockData || !this.shares) return 0
      return this.stockData.price * this.shares
    },

    // Check if stock price went up or down
    changeClass() {
      if (!this.stockData) return ''
      return this.stockData.changePercent >= 0 ? 'positive' : 'negative'
    },

    // Current shares of this stock
    currentShares() {
      return this.userStore.getShares(this.stockData?.symbol)
    },

    // Can the user execute this trade?
    canTrade() {
      if (!this.stockData || !this.shares || this.shares <= 0) return false

      if (this.tradeType === 'buy') {
        return this.userStore.canAfford(this.totalAmount)
      } else {
        return this.userStore.canSell(this.stockData.symbol, this.shares)
      }
    }
  },

  async mounted() {
    // Load user data and check for symbol in query params
    await this.userStore.loadPortfolio()

    if (this.$route.query.symbol) {
      this.stockSymbol = this.$route.query.symbol.toUpperCase()
      await this.lookupStock()
    }
  },

  methods: {
    // Clear previous data when symbol changes
    onSymbolChange() {
      this.stockData = null
      this.error = null
      this.successMessage = ''
      this.shares = 1
    },

    // Look up stock information
    async lookupStock() {
      if (!this.stockSymbol) return

      this.loading = true
      this.error = null

      try {
        const response = await fetch(`http://localhost:8002/stock/${this.stockSymbol.toUpperCase()}`)

        if (!response.ok) {
          throw new Error(`Stock not found or API error: ${response.status}`)
        }

        const data = await response.json()

        this.stockData = {
          symbol: data.symbol,
          name: data.name,
          price: data.price,
          changePercent: data.change_percent
        }

        // Update URL with symbol
        this.$router.replace({
          query: { ...this.$route.query, symbol: data.symbol }
        })

      } catch (err) {
        console.error('Error fetching stock:', err)
        this.error = err.message || 'Failed to fetch stock data'
        this.stockData = null
      } finally {
        this.loading = false
      }
    },

    // Execute the trade using Pinia store
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

        await this.userStore.executeTrade(tradeData)

        this.successMessage = `Successfully ${this.tradeType === 'buy' ? 'bought' : 'sold'} ${this.shares} shares of ${this.stockData.symbol} at $${this.stockData.price.toFixed(2)} per share`

        // Reset form
        this.shares = 1

      } catch (err) {
        this.error = err.message || 'Trade execution failed'
      } finally {
        this.executing = false
      }
    },

    // Quick trade actions
    quickTrade(symbol, action) {
      this.stockSymbol = symbol
      this.tradeType = action
      this.lookupStock()
    },

    // Quantity control methods
    increaseShares() {
      if (this.tradeType === 'sell' && this.shares >= this.currentShares) {
        return
      }
      this.shares += 1
    },

    decreaseShares() {
      if (this.shares > 1) {
        this.shares -= 1
      }
    },

    // Set maximum buyable shares based on available cash
    setMaxBuyShares() {
      if (this.stockData && this.portfolio.cash > 0) {
        const maxShares = Math.floor(this.portfolio.cash / this.stockData.price)
        this.shares = Math.max(1, maxShares)
      }
    },

    // Clear success message
    clearMessage() {
      this.successMessage = ''
    },

    formatCurrency(amount) {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(amount || 0)
    }
  }
}
</script>

<style scoped>
@import '../styles/components/trade.css';

/* Additional enhanced styles */
.portfolio-summary {
  margin-bottom: 30px;
}

.trade-type {
  display: flex;
  gap: 30px;
  margin: 20px 0;
  justify-content: center;
}

.trade-type label {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.trade-type input[type="radio"] {
  display: none;
}

.radio-label {
  padding: 12px 24px;
  border: 2px solid #dee2e6;
  border-radius: 25px;
  font-weight: 600;
  font-size: 1.1rem;
  transition: all 0.3s;
  min-width: 80px;
  text-align: center;
}

.buy-label {
  color: #28a745;
  border-color: #28a745;
}

.sell-label {
  color: #dc3545;
  border-color: #dc3545;
}

.trade-type input[type="radio"]:checked + .buy-label {
  background: #28a745;
  color: white;
}

.trade-type input[type="radio"]:checked + .sell-label {
  background: #dc3545;
  color: white;
}

.quick-qty {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin: 15px 0;
}

.quick-btn {
  padding: 6px 12px;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.quick-btn:hover {
  background: #e9ecef;
}

.order-summary {
  background: linear-gradient(135deg, #f8f9fa, #ffffff);
  border-radius: 12px;
  padding: 20px;
  margin: 20px 0;
  border: 2px solid #e9ecef;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  padding: 5px 0;
}

.total-row {
  border-top: 2px solid #dee2e6;
  margin-top: 15px;
  padding-top: 15px;
  font-size: 1.1rem;
}

.total-amount {
  font-weight: bold;
  color: #007bff;
  font-size: 1.2rem;
}

.validation-check {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #dee2e6;
}

.check-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.insufficient-funds, .insufficient-shares {
  color: #dc3545;
  font-weight: 600;
  background: #f8d7da;
  padding: 8px 12px;
  border-radius: 6px;
  margin-top: 10px;
}

.trade-btn-main {
  width: 100%;
  padding: 15px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1.1rem;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.buy-btn {
  background: linear-gradient(135deg, #28a745, #1e7e34);
  color: white;
  border: none;
}

.sell-btn {
  background: linear-gradient(135deg, #dc3545, #c82333);
  color: white;
  border: none;
}

.buy-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(40, 167, 69, 0.3);
}

.sell-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(220, 53, 69, 0.3);
}

.buy-btn:disabled,
.sell-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.transaction-result {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.result-card {
  background: white;
  padding: 40px;
  border-radius: 16px;
  text-align: center;
  max-width: 450px;
  margin: 20px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.3);
  animation: slideIn 0.4s ease;
}

.success-icon {
  font-size: 64px;
  margin-bottom: 20px;
}

.result-actions {
  display: flex;
  gap: 15px;
  margin-top: 25px;
}

.close-btn, .dashboard-btn {
  flex: 1;
  padding: 12px 20px;
  border-radius: 6px;
  font-weight: 600;
  text-decoration: none;
  text-align: center;
  transition: all 0.3s;
}

.close-btn {
  background: #28a745;
  color: white;
  border: none;
  cursor: pointer;
}

.dashboard-btn {
  background: #007bff;
  color: white;
}

.holdings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.holding-card {
  background: white;
  border: 2px solid #e9ecef;
  border-radius: 12px;
  padding: 20px;
  transition: all 0.3s;
}

.holding-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 20px rgba(0,0,0,0.1);
}

.holding-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.holding-symbol {
  font-weight: bold;
  font-size: 1.3rem;
  color: #007bff;
}

.holding-shares {
  background: #e7f3ff;
  color: #0066cc;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
}

.holding-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 0.95rem;
}

.holding-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.quick-action-btn {
  flex: 1;
  padding: 8px 12px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
  transition: all 0.3s;
}

.quick-action-btn.buy {
  background: #28a745;
  color: white;
}

.quick-action-btn.sell {
  background: #dc3545;
  color: white;
}

.quick-action-btn:hover {
  transform: translateY(-1px);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-30px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (max-width: 768px) {
  .trade-type {
    flex-direction: column;
    align-items: center;
    gap: 15px;
  }

  .holdings-grid {
    grid-template-columns: 1fr;
  }

  .result-actions {
    flex-direction: column;
  }
}
</style>