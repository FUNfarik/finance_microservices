<template>
  <div class="trade-page">
    <!-- Trade Header -->
    <div class="trade-header">
      <div class="header-content">
        <h1>Trade Stocks</h1>
        <p class="welcome-text">Buy and sell stocks with real-time market data</p>
      </div>
      <div class="header-actions">
        <router-link to="/dashboard" class="nav-button">
          ← Back to Dashboard
        </router-link>
      </div>
    </div>

    <!-- Portfolio Summary -->
    <div class="stats-grid cols-2">
      <div class="stat-card primary">
        <div class="stat-icon">💵</div>
        <div class="stat-content">
          <h3>Available Cash</h3>
          <p class="big-number">${{ formatCurrency(portfolio.cash) }}</p>
          <span class="subtext">Ready to invest</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <h3>Holdings</h3>
          <p class="big-number">{{ portfolio.holdings?.length || 0 }}</p>
          <span class="subtext">Unique positions</span>
        </div>
      </div>
    </div>

    <!-- Stock Lookup & Trading Section -->
    <div class="trading-section">
      <div class="section-header">
        <h2>Stock Lookup & Trading</h2>
      </div>

      <!-- Stock Search -->
      <div class="search-container">
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

        <!-- Error message -->
        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </div>

      <!-- Stock Information & Trading Form (Combined) -->
      <div v-if="stockData" class="stock-trading-card">
        <!-- Stock Info Header -->
        <div class="stock-info-header">
          <div class="stock-details">
            <div class="stock-symbol">{{ stockData.symbol }}</div>
            <div class="stock-name">{{ stockData.name }}</div>
            <div class="stock-price-info">
              <span class="current-price">${{ stockData.price.toFixed(2) }}</span>
              <span class="price-change" :class="changeClass">
                {{ stockData.changePercent > 0 ? '+' : '' }}{{ stockData.changePercent.toFixed(2) }}%
              </span>
            </div>
          </div>

          <!-- Current Holdings (if any) -->
          <div v-if="currentShares > 0" class="current-holdings">
            <div class="holdings-badge">
              <span class="holdings-text">You own</span>
              <span class="holdings-count">{{ currentShares }}</span>
              <span class="holdings-text">shares</span>
            </div>
          </div>
        </div>

        <!-- Trading Form (Integrated) -->
        <div class="trading-form">
          <div class="form-header">
            <h3>Place Order</h3>
          </div>

          <!-- Trade Type Selection -->
          <div class="trade-type-selector">
            <label class="trade-type-option">
              <input type="radio" v-model="tradeType" value="buy" />
              <span class="option-button buy-option">
                <span class="option-text">Buy</span>
              </span>
            </label>
            <label class="trade-type-option">
              <input type="radio" v-model="tradeType" value="sell" />
              <span class="option-button sell-option">
                <span class="option-text">Sell</span>
              </span>
            </label>
          </div>

          <!-- Quantity Controls -->
          <div class="quantity-section">
            <div class="quantity-controls">
              <button @click="decreaseShares" class="qty-btn" :disabled="shares <= 1">−</button>
              <input
                  v-model.number="shares"
                  type="number"
                  min="1"
                  :max="tradeType === 'sell' ? currentShares : 9999"
                  class="qty-input"
              />
              <button @click="increaseShares" class="qty-btn">+</button>
            </div>

            <!-- Quick Quantity Buttons -->
            <div class="quick-quantities">
              <template v-if="tradeType === 'sell' && currentShares > 0">
                <button @click="shares = Math.max(1, Math.floor(currentShares / 4))" class="quick-btn">25%</button>
                <button @click="shares = Math.max(1, Math.floor(currentShares / 2))" class="quick-btn">50%</button>
                <button @click="shares = currentShares" class="quick-btn">All</button>
              </template>
              <template v-if="tradeType === 'buy'">
                <button @click="setMaxBuyShares" class="quick-btn">Max</button>
              </template>
            </div>
          </div>

          <!-- Order Summary -->
          <div v-if="shares > 0" class="order-preview">
            <div class="preview-header">
              <h4>Order Preview</h4>
            </div>

            <div class="preview-details">
              <div class="preview-row">
                <span class="label">Action:</span>
                <span class="value">
                  <strong>{{ tradeType.toUpperCase() }} {{ shares }} shares</strong>
                </span>
              </div>

              <div class="preview-row">
                <span class="label">Price per share:</span>
                <span class="value">${{ stockData.price.toFixed(2) }}</span>
              </div>

              <div class="preview-row total">
                <span class="label">Total {{ tradeType === 'buy' ? 'Cost' : 'Proceeds' }}:</span>
                <span class="value total-amount">${{ totalAmount.toFixed(2) }}</span>
              </div>
            </div>

            <!-- Validation Messages -->
            <div class="validation-section">
              <!-- Cash validation for buying -->
              <div v-if="tradeType === 'buy'" class="validation-info">
                <div class="validation-row">
                  <span>Available Cash: ${{ formatCurrency(portfolio.cash) }}</span>
                </div>
                <div v-if="totalAmount > portfolio.cash" class="validation-error">
                  ❌ Insufficient funds (need ${{ formatCurrency(totalAmount - portfolio.cash) }} more)
                </div>
              </div>

              <!-- Holdings validation for selling -->
              <div v-if="tradeType === 'sell'" class="validation-info">
                <div class="validation-row">
                  <span>Current Holdings: {{ currentShares }} shares</span>
                </div>
                <div v-if="shares > currentShares" class="validation-error">
                  ❌ Insufficient shares (you only own {{ currentShares }} shares)
                </div>
              </div>
            </div>

            <!-- Execute Trade Button -->
            <div class="trade-execution">
              <button
                  @click="executeTrade"
                  :disabled="!canTrade || executing"
                  :class="['execute-btn', tradeType === 'buy' ? 'buy-btn' : 'sell-btn']"
              >
                <span v-if="executing" class="spinner">⏳</span>
                {{ executing ? 'Processing...' : `${tradeType.toUpperCase()} ${shares} Shares` }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="successMessage" class="success-modal" @click="clearMessage">
      <div class="modal-content" @click.stop>
        <div class="success-header">
          <div class="success-icon">✓</div>
          <h3>Trade Executed Successfully!</h3>
        </div>
        <div class="success-body">
          <p>{{ successMessage }}</p>
        </div>
        <div class="success-actions">
          <button @click="clearMessage" class="continue-btn">Continue Trading</button>
          <router-link to="/dashboard" class="dashboard-btn">View Dashboard</router-link>
        </div>
      </div>
    </div>

    <!-- Current Holdings -->
    <div v-if="portfolio.holdings?.length > 0" class="holdings-section">
      <div class="section-header">
        <h2>Your Current Holdings</h2>
        <span class="holdings-count-text">{{ portfolio.holdings.length }} positions</span>
      </div>

      <div class="holdings-grid">
        <div v-for="holding in portfolio.holdings" :key="holding.symbol" class="holding-card">
          <div class="holding-header">
            <span class="holding-symbol">{{ holding.symbol }}</span>
            <span class="holding-shares">{{ holding.shares }} shares</span>
          </div>

          <div class="holding-details">
            <div class="detail-row">
              <span>Avg Price:</span>
              <span>${{ formatCurrency(holding.avg_price) }}</span>
            </div>
            <div class="detail-row">
              <span>Current:</span>
              <span>${{ formatCurrency(holding.current_price || 0) }}</span>
            </div>
            <div class="detail-row">
              <span>Market Value:</span>
              <span>${{ formatCurrency(holding.market_value || 0) }}</span>
            </div>
            <div class="detail-row gain-loss-row">
              <span>Gain/Loss:</span>
              <span :class="['gain-loss', (holding.gain_loss || 0) >= 0 ? 'positive' : 'negative']">
                {{ (holding.gain_loss || 0) >= 0 ? '+' : '' }}${{ formatCurrency(Math.abs(holding.gain_loss || 0)) }}
              </span>
            </div>
          </div>

          <div class="holding-actions">
            <button @click="quickTrade(holding.symbol, 'buy')" class="action-btn buy-more-btn">
              Buy More
            </button>
            <button @click="quickTrade(holding.symbol, 'sell')" class="action-btn sell-btn">
              Sell
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { useUserStore } from '../stores/user'
import { marketApi } from '../services/api.js'

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
      if (!this.stockData?.symbol || !this.portfolio.holdings) return 0
      const holding = this.portfolio.holdings.find(h => h.symbol === this.stockData.symbol)
      return holding ? holding.shares : 0
    },

    // Can the user execute this trade?
    canTrade() {
      if (!this.stockData || !this.shares || this.shares <= 0) return false

      if (this.tradeType === 'buy') {
        return this.totalAmount <= this.portfolio.cash
      } else {
        return this.shares <= this.currentShares
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
        const response = await marketApi.get(`/stock/${this.stockSymbol.toUpperCase()}`)
        const data = response.data

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
        this.error = err.response?.data?.error || err.message || 'Failed to fetch stock data'
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

        this.successMessage = `Successfully ${this.tradeType === 'buy' ? 'bought' : 'sold'} ${this.shares} shares of ${this.stockData.symbol} at ${this.stockData.price.toFixed(2)} per share`

        // Reset form
        this.shares = 1

      } catch (err) {
        if (err.response?.status === 401) {
          this.$router.push('/login')
          return
        }
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
.trade-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

/* Header Section - Consistent with Dashboard */
.trade-header {
  background: white;
  border-radius: 12px;
  padding: 24px 30px;
  margin-bottom: 30px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-content h1 {
  color: #333;
  margin: 0 0 5px 0;
  font-size: 2.5rem;
  font-weight: 600;
}

.welcome-text {
  color: #666;
  margin: 0;
  font-size: 1.1rem;
}

.nav-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #6c757d;
  color: white;
  border: none;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

.nav-button:hover {
  background: #545b62;
  text-decoration: none;
  color: white;
}

/* Stats Grid - Consistent with Dashboard */
.stats-grid {
  display: grid;
  gap: 25px;
  margin-bottom: 40px;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

.stats-grid.cols-2 {
  grid-template-columns: repeat(2, 1fr);
}

.stat-card {
  background: white;
  padding: 30px;
  border-radius: 15px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 20px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: #dee2e6;
}

.stat-card.primary::before {
  background: #007bff;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.15);
}

.stat-icon {
  font-size: 3rem;
  opacity: 0.8;
}

.stat-content h3 {
  margin: 0 0 8px 0;
  color: #666;
  font-size: 1rem;
  font-weight: 500;
}

.big-number {
  font-size: 2.2rem;
  font-weight: 700;
  color: #333;
  margin: 8px 0;
}

.subtext {
  font-size: 0.85rem;
  color: #888;
  margin-top: 4px;
}

/* Trading Section */
.trading-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  margin-bottom: 40px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 15px;
  border-bottom: 2px solid #f8f9fa;
}

.section-header h2 {
  color: #333;
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.holdings-count-text {
  color: #666;
  font-size: 0.9rem;
}

/* Search Section */
.search-container {
  margin-bottom: 30px;
}

.search-box {
  display: flex;
  gap: 12px;
  margin-bottom: 15px;
}

.search-input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.3s ease;
}

.search-input:focus {
  outline: none;
  border-color: #007bff;
}

.search-btn {
  padding: 12px 20px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

.search-btn:hover:not(:disabled) {
  background: #0056b3;
}

.search-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.error-message {
  background: #f8d7da;
  color: #721c24;
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid #f5c6cb;
  margin-bottom: 20px;
}

/* Stock Trading Card */
.stock-trading-card {
  border: 2px solid #e9ecef;
  border-radius: 12px;
  overflow: hidden;
  background: #f8f9fa;
}

.stock-info-header {
  background: white;
  padding: 25px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 2px solid #e9ecef;
}

.stock-details {
  flex: 1;
}

.stock-symbol {
  font-size: 2rem;
  font-weight: 700;
  color: #007bff;
  margin-bottom: 5px;
}

.stock-name {
  color: #666;
  font-size: 1.1rem;
  margin-bottom: 15px;
}

.stock-price-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.current-price {
  font-size: 2.5rem;
  font-weight: 700;
  color: #333;
}

.price-change {
  font-size: 1.2rem;
  font-weight: 600;
  padding: 6px 12px;
  border-radius: 20px;
}

.price-change.positive {
  background: #d4edda;
  color: #28a745;
}

.price-change.negative {
  background: #f8d7da;
  color: #dc3545;
}

.current-holdings {
  display: flex;
  align-items: center;
}

.holdings-badge {
  background: #e7f3ff;
  color: #0066cc;
  padding: 12px 20px;
  border-radius: 25px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.holdings-count {
  background: white;
  padding: 4px 12px;
  border-radius: 15px;
  font-size: 1.2rem;
  color: #007bff;
}

/* Trading Form */
.trading-form {
  padding: 30px;
}

.form-header h3 {
  margin: 0 0 20px 0;
  color: #333;
  font-size: 1.3rem;
}

.trade-type-selector {
  display: flex;
  gap: 20px;
  margin-bottom: 25px;
  justify-content: center;
}

.trade-type-option {
  cursor: pointer;
}

.trade-type-option input[type="radio"] {
  display: none;
}

.option-button {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px 25px;
  border: 2px solid #dee2e6;
  border-radius: 12px;
  transition: all 0.3s ease;
  font-weight: 600;
  min-width: 120px;
  justify-content: center;
}

.buy-option {
  color: #28a745;
  border-color: #28a745;
}

.sell-option {
  color: #dc3545;
  border-color: #dc3545;
}

.trade-type-option input[type="radio"]:checked + .buy-option {
  background: #28a745;
  color: white;
}

.trade-type-option input[type="radio"]:checked + .sell-option {
  background: #dc3545;
  color: white;
}

.option-icon {
  font-size: 1.2rem;
}

/* Quantity Section */
.quantity-section {
  margin-bottom: 25px;
}

.quantity-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0;
  margin-bottom: 15px;
}

.qty-btn {
  width: 50px;
  height: 50px;
  background: #f8f9fa;
  border: 2px solid #dee2e6;
  cursor: pointer;
  font-size: 1.5rem;
  font-weight: 700;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qty-btn:first-child {
  border-radius: 8px 0 0 8px;
}

.qty-btn:last-child {
  border-radius: 0 8px 8px 0;
}

.qty-btn:hover:not(:disabled) {
  background: #e9ecef;
}

.qty-btn:disabled {
  background: #f8f9fa;
  color: #ccc;
  cursor: not-allowed;
}

.qty-input {
  width: 120px;
  height: 50px;
  text-align: center;
  border: 2px solid #dee2e6;
  border-left: none;
  border-right: none;
  font-size: 1.2rem;
  font-weight: 600;
}

.qty-input:focus {
  outline: none;
  border-color: #007bff;
}

.quick-quantities {
  display: flex;
  justify-content: center;
  gap: 10px;
}

.quick-btn {
  padding: 8px 16px;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.quick-btn:hover {
  background: #e9ecef;
}

/* Order Preview */
.order-preview {
  background: white;
  border: 2px solid #e9ecef;
  border-radius: 12px;
  padding: 20px;
  margin-top: 20px;
}

.preview-header h4 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 1.1rem;
}

.preview-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  padding: 8px 0;
}

.preview-row.total {
  border-top: 2px solid #dee2e6;
  margin-top: 15px;
  padding-top: 15px;
  font-weight: 600;
}

.total-amount {
  font-size: 1.3rem;
  color: #007bff;
  font-weight: 700;
}

.validation-section {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #dee2e6;
}

.validation-row {
  color: #666;
  font-size: 0.9rem;
  margin-bottom: 8px;
}

.validation-error {
  background: #f8d7da;
  color: #721c24;
  padding: 10px 15px;
  border-radius: 6px;
  border: 1px solid #f5c6cb;
  margin-top: 10px;
  font-weight: 500;
}

.trade-execution {
  margin-top: 20px;
}

.execute-btn {
  width: 100%;
  padding: 15px;
  border: none;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.execute-btn.buy-btn {
  background: #28a745;
  color: white;
}

.execute-btn.sell-btn {
  background: #dc3545;
  color: white;
}

.execute-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.execute-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Success Modal */
.success-modal {
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

.modal-content {
  background: white;
  padding: 40px;
  border-radius: 16px;
  text-align: center;
  max-width: 450px;
  margin: 20px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.success-header {
  margin-bottom: 20px;
}

.success-icon {
  font-size: 4rem;
  margin-bottom: 15px;
  color: #28a745;
  font-weight: bold;
}

.success-header h3 {
  margin: 0;
  color: #333;
  font-size: 1.5rem;
}

.success-body {
  margin-bottom: 30px;
  color: #666;
  font-size: 1.1rem;
}

.success-actions {
  display: flex;
  gap: 15px;
}

.continue-btn, .dashboard-btn {
  flex: 1;
  padding: 12px 20px;
  border-radius: 6px;
  font-weight: 600;
  text-decoration: none;
  text-align: center;
  transition: all 0.2s ease;
  border: none;
  cursor: pointer;
}

.continue-btn {
  background: #28a745;
  color: white;
}

.dashboard-btn {
  background: #007bff;
  color: white;
}

.continue-btn:hover, .dashboard-btn:hover {
  transform: translateY(-1px);
  text-decoration: none;
  color: white;
}

/* Holdings Section */
.holdings-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.holdings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.holding-card {
  background: #f8f9fa;
  border: 2px solid #e9ecef;
  border-radius: 12px;
  padding: 20px;
  transition: all 0.3s ease;
}

.holding-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  border-color: #007bff;
}

.holding-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.holding-symbol {
  font-weight: 700;
  font-size: 1.4rem;
  color: #007bff;
}

.holding-shares {
  background: #e7f3ff;
  color: #0066cc;
  padding: 6px 12px;
  border-radius: 15px;
  font-size: 0.9rem;
  font-weight: 600;
}

.holding-details {
  margin-bottom: 15px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 0.95rem;
}

.gain-loss-row {
  padding-top: 8px;
  border-top: 1px solid #dee2e6;
  font-weight: 600;
}

.gain-loss.positive {
  color: #28a745;
}

.gain-loss.negative {
  color: #dc3545;
}

.holding-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  flex: 1;
  padding: 10px 15px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.buy-more-btn {
  background: #28a745;
  color: white;
}

.sell-btn {
  background: #dc3545;
  color: white;
}

.action-btn:hover {
  transform: translateY(-1px);
}

/* Responsive Design */
@media (max-width: 1200px) {
  .stats-grid.cols-2 {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .trade-page {
    padding: 15px;
  }

  .trade-header {
    flex-direction: column;
    gap: 20px;
    text-align: center;
    padding: 25px;
  }

  .header-content h1 {
    font-size: 2rem;
  }

  .stats-grid.cols-2 {
    grid-template-columns: 1fr;
  }

  .stat-card {
    flex-direction: column;
    text-align: center;
    padding: 25px;
  }

  .trading-section {
    padding: 20px;
  }

  .search-box {
    flex-direction: column;
  }

  .stock-info-header {
    flex-direction: column;
    gap: 20px;
  }

  .trade-type-selector {
    flex-direction: column;
    align-items: center;
    gap: 15px;
  }

  .holdings-grid {
    grid-template-columns: 1fr;
  }

  .success-actions {
    flex-direction: column;
  }
}

@media (max-width: 480px) {
  .trade-header {
    padding: 20px;
  }

  .header-content h1 {
    font-size: 1.8rem;
  }

  .stat-card {
    padding: 20px;
  }

  .trading-section {
    padding: 15px;
  }

  .stock-trading-card {
    margin: 0 -5px;
  }
}
</style>