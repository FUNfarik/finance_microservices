<template>
  <div class="dashboard">
    <!-- Dashboard Header -->
    <div class="dashboard-header">
      <div class="header-content">
        <h1>Portfolio Dashboard</h1>
        <p class="welcome-text">Welcome back! Here's your portfolio overview</p>
      </div>
      <div class="header-actions">
        <button @click="refreshData" :disabled="isRefreshing" class="refresh-btn">
          <span :class="{ 'spinning': isRefreshing }">🔄</span>
          {{ isRefreshing ? 'Refreshing...' : 'Refresh' }}
        </button>
        <span v-if="lastUpdate" class="last-update">
          Updated: {{ formatTimestamp(lastUpdate) }}
        </span>
      </div>
    </div>

    <!-- Service Status - Development Only -->
    <div v-if="showTesting" class="test-section">
      <button @click="testServices" class="test-btn" :disabled="testing">
        {{ testing ? 'Testing Services...' : 'Test Backend Services' }}
      </button>
      <div v-if="serviceStatus.length > 0" class="service-status">
        <div v-for="status in serviceStatus" :key="status.service"
             :class="['status-item', status.success ? 'success' : 'error']">
          {{ status.service }}: {{ status.message }}
        </div>
      </div>
    </div>

    <!-- Portfolio Stats -->
    <div class="stats-grid cols-4">
      <div class="stat-card primary">
        <div class="stat-icon">💰</div>
        <div class="stat-content">
          <h3>Portfolio Value</h3>
          <p class="big-number">${{ formatCurrency(portfolio.total_value) }}</p>
          <span :class="['change', portfolio.total_gain_loss >= 0 ? 'positive' : 'negative']">
            {{ portfolio.total_gain_loss >= 0 ? '+' : '' }}${{ formatCurrency(portfolio.total_gain_loss) }}
            ({{ portfolio.total_gain_loss_percent?.toFixed(2) }}%)
          </span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💵</div>
        <div class="stat-content">
          <h3>Cash Balance</h3>
          <p class="big-number">${{ formatCurrency(portfolio.cash) }}</p>
          <span class="subtext">Available to trade</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">📈</div>
        <div class="stat-content">
          <h3>Total Net Worth</h3>
          <p class="big-number">${{ formatCurrency(totalNetWorth) }}</p>
          <span class="subtext">Portfolio + Cash</span>
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

    <!-- Portfolio Content -->
    <div class="portfolio-content">
      <!-- Error Display -->
      <div v-if="errors.portfolio" class="error-banner">
        <span>⚠️ Failed to load portfolio data: {{ errors.portfolio }}</span>
        <button @click="clearError('portfolio')" class="error-close">×</button>
      </div>

      <!-- Holdings Table -->
      <div class="holdings-section">
        <div class="section-header">
          <h2>Your Holdings</h2>
          <div class="section-actions">
            <button @click="updatePrices" :disabled="updatingPrices" class="update-prices-btn">
              {{ updatingPrices ? 'Updating...' : 'Update Prices' }}
            </button>
          </div>
        </div>

        <div v-if="loading.portfolio" class="loading">
          <div class="spinner"></div>
          <p>Loading portfolio...</p>
        </div>

        <div v-else-if="portfolio.holdings?.length > 0" class="holdings-table">
          <div class="table-header">
            <span>Symbol</span>
            <span>Shares</span>
            <span>Avg Price</span>
            <span>Current Price</span>
            <span>Market Value</span>
            <span>Gain/Loss</span>
            <span>Actions</span>
          </div>

          <div v-for="holding in portfolio.holdings" :key="holding.symbol" class="table-row">
            <span class="symbol">{{ holding.symbol }}</span>
            <span class="shares">{{ holding.shares }}</span>
            <span class="price">${{ formatCurrency(holding.avg_price) }}</span>
            <span class="current-price">
              ${{ formatCurrency(holding.current_price || 0) }}
              <span v-if="!holding.current_price" class="price-updating">📊</span>
            </span>
            <span class="market-value">${{ formatCurrency(holding.market_value || 0) }}</span>
            <span :class="['gain-loss', (holding.gain_loss || 0) >= 0 ? 'positive' : 'negative']">
              {{ (holding.gain_loss || 0) >= 0 ? '+' : '' }}${{ formatCurrency(holding.gain_loss || 0) }}
            </span>
            <span class="actions">
              <button @click="goToTrade(holding.symbol)" class="action-btn trade-btn">Trade</button>
            </span>
          </div>
        </div>

        <div v-else class="no-holdings">
          <div class="empty-state">
            <div class="empty-icon">📊</div>
            <h3>No Holdings Yet</h3>
            <p>Start building your portfolio by trading stocks</p>
            <router-link to="/trade" class="cta-button">Start Trading</router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions cols-3">
      <router-link to="/trade" class="action-card trade">
        <div class="action-icon">💹</div>
        <h3>Trade Stocks</h3>
        <p>Buy and sell stocks</p>
      </router-link>

      <div class="action-card history" @click="showTransactionHistory = true">
        <div class="action-icon">📋</div>
        <h3>History</h3>
        <p>View past transactions</p>
      </div>

      <div class="action-card refresh" @click="refreshData">
        <div class="action-icon">🔄</div>
        <h3>Refresh</h3>
        <p>Update all data</p>
      </div>
    </div>
  </div>
</template>

<script>
import { useUserStore } from '../stores/user'

export default {
  name: 'Dashboard',
  setup() {
    const userStore = useUserStore()
    return { userStore }
  },

  data() {
    return {
      // Testing flag - set to false for production, true for development
      showTesting: false, // Change this to true when you need testing

      testing: false,
      serviceStatus: [],
      isRefreshing: false,
      updatingPrices: false,
      showTransactionHistory: false
    }
  },

  computed: {
    portfolio() { return this.userStore.portfolio },
    loading() { return this.userStore.loading },
    errors() { return this.userStore.errors },
    lastUpdate() { return this.userStore.lastUpdate },
    totalNetWorth() { return this.userStore.totalNetWorth }
  },

  async mounted() {
    // Initialize demo user for development
    this.userStore.initDemoUser()

    // Load portfolio data on mount
    await this.userStore.loadPortfolio()
  },

  methods: {
    async refreshData() {
      this.isRefreshing = true
      try {
        await this.userStore.refreshPortfolio()
      } finally {
        this.isRefreshing = false
      }
    },

    async updatePrices() {
      this.updatingPrices = true
      try {
        await this.userStore.updateHoldingsPrices()
      } finally {
        this.updatingPrices = false
      }
    },

    async testServices() {
      this.testing = true
      this.serviceStatus = []

      // Test Portfolio Service
      try {
        await this.userStore.loadPortfolio()
        this.serviceStatus.push({
          service: 'Portfolio Service',
          success: true,
          message: '✅ Connected'
        })
      } catch (error) {
        this.serviceStatus.push({
          service: 'Portfolio Service',
          success: false,
          message: '❌ Failed: ' + error.message
        })
      }

      // Test Market Data Service
      try {
        const response = await fetch('http://localhost:8002/health')
        if (response.ok) {
          this.serviceStatus.push({
            service: 'Market Data Service',
            success: true,
            message: '✅ Connected'
          })
        } else {
          throw new Error('Health check failed')
        }
      } catch (error) {
        this.serviceStatus.push({
          service: 'Market Data Service',
          success: false,
          message: '❌ Failed: ' + error.message
        })
      }

      this.testing = false
    },

    clearError(type) {
      this.userStore.clearError(type)
    },

    goToTrade(symbol = '') {
      this.$router.push({
        path: '/trade',
        query: symbol ? { symbol } : {}
      })
    },

    formatCurrency(amount) {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(amount || 0)
    },

    formatTimestamp(timestamp) {
      return new Date(timestamp).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit'
      })
    }
  }
}
</script>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.dashboard-header {
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 15px;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #0056b3;
}

.refresh-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.last-update {
  font-size: 0.9rem;
  color: #666;
}

.test-section {
  background: rgba(255, 255, 255, 0.95);
  margin-bottom: 30px;
  padding: 25px;
  border-radius: 15px;
  text-align: center;
  backdrop-filter: blur(10px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.test-btn {
  padding: 10px 20px;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

.test-btn:hover:not(:disabled) {
  background: #1e7e34;
}

.test-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.service-status {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.status-item {
  padding: 12px 16px;
  border-radius: 8px;
  font-weight: 500;
}

.status-item.success {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.status-item.error {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

/* Grid System for Cards */
.stats-grid,
.quick-actions {
  display: grid;
  gap: 25px;
  margin-bottom: 40px;
}

/* Default: Auto-fit responsive grid */
.stats-grid,
.quick-actions {
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

/* 1 Column Layout */
.stats-grid.cols-1,
.quick-actions.cols-1 {
  grid-template-columns: 1fr;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

/* 2 Column Layout */
.stats-grid.cols-2,
.quick-actions.cols-2 {
  grid-template-columns: repeat(2, 1fr);
}

/* 3 Column Layout */
.stats-grid.cols-3,
.quick-actions.cols-3 {
  grid-template-columns: repeat(3, 1fr);
}

/* 4 Column Layout */
.stats-grid.cols-4,
.quick-actions.cols-4 {
  grid-template-columns: repeat(4, 1fr);
}

/* Responsive behavior for column layouts */
@media (max-width: 1200px) {
  .stats-grid.cols-4,
  .quick-actions.cols-4 {
    grid-template-columns: repeat(2, 1fr);
  }

  .stats-grid.cols-3,
  .quick-actions.cols-3 {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid.cols-4,
  .stats-grid.cols-3,
  .stats-grid.cols-2,
  .quick-actions.cols-4,
  .quick-actions.cols-3,
  .quick-actions.cols-2 {
    grid-template-columns: 1fr;
  }
}

/* Ensure cards have consistent minimum widths */
.stat-card,
.action-card {
  min-width: 250px;
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
  transform: translateY(-8px);
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

.change {
  font-size: 0.9rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.change.positive {
  color: #28a745;
}

.change.negative {
  color: #dc3545;
}

.subtext {
  font-size: 0.85rem;
  color: #888;
  margin-top: 4px;
}

.error-banner {
  background: #f8d7da;
  color: #721c24;
  padding: 16px 20px;
  border-radius: 10px;
  margin-bottom: 25px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid #f5c6cb;
}

.error-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #721c24;
  padding: 0 5px;
}

.holdings-section {
  background: white;
  padding: 30px;
  border-radius: 15px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  margin-bottom: 40px;
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

.update-prices-btn {
  padding: 8px 14px;
  background: #17a2b8;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  font-size: 14px;
  transition: background-color 0.2s ease;
}

.update-prices-btn:hover:not(:disabled) {
  background: #138496;
}

.loading {
  text-align: center;
  padding: 60px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.holdings-table {
  overflow-x: auto;
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1.2fr 1.2fr 0.8fr;
  gap: 20px;
  padding: 20px 0;
  border-bottom: 2px solid #e9ecef;
  font-weight: 600;
  color: #495057;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 10px;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1.2fr 1.2fr 0.8fr;
  gap: 20px;
  padding: 20px 0;
  border-bottom: 1px solid #f1f3f4;
  align-items: center;
  transition: all 0.2s ease;
}

.table-row:hover {
  background: #f8f9fa;
  border-radius: 8px;
}

.symbol {
  font-weight: 700;
  color: #667eea;
  font-size: 1.1rem;
}

.shares, .price, .market-value {
  font-weight: 500;
  color: #495057;
}

.current-price {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.price-updating {
  font-size: 0.8rem;
  opacity: 0.7;
}

.gain-loss {
  font-weight: 600;
  font-size: 1rem;
}

.gain-loss.positive {
  color: #28a745;
}

.gain-loss.negative {
  color: #dc3545;
}

.actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  padding: 8px 12px;
  font-size: 0.85rem;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.trade-btn {
  background: #007bff;
  color: white;
}

.trade-btn:hover {
  background: #0056b3;
}

.no-holdings {
  text-align: center;
  padding: 80px 20px;
}

.empty-state {
  max-width: 400px;
  margin: 0 auto;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 20px;
  opacity: 0.6;
}

.empty-state h3 {
  color: #495057;
  margin-bottom: 12px;
  font-size: 1.5rem;
}

.empty-state p {
  color: #6c757d;
  margin-bottom: 30px;
  font-size: 1.1rem;
}

.cta-button {
  display: inline-block;
  padding: 12px 24px;
  background: #28a745;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 1rem;
  transition: background-color 0.2s ease;
}

.cta-button:hover {
  text-decoration: none;
  color: white;
  background: #1e7e34;
}

.action-card {
  background: white;
  padding: 30px;
  border-radius: 15px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  text-decoration: none;
  color: inherit;
  position: relative;
  overflow: hidden;
}

.action-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: #dee2e6;
}

.action-card.trade::before {
  background: #007bff;
}

.action-card.history::before {
  background: #6c757d;
}

.action-card.refresh::before {
  background: #28a745;
}

.action-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.15);
  text-decoration: none;
  color: inherit;
}

.action-icon {
  font-size: 3rem;
  margin-bottom: 15px;
  opacity: 0.8;
}

.action-card h3 {
  margin: 0 0 10px 0;
  color: #333;
  font-size: 1.3rem;
  font-weight: 600;
}

.action-card p {
  margin: 0;
  color: #666;
  font-size: 1rem;
}

/* Responsive Design */
@media (max-width: 768px) {
  .dashboard {
    padding: 15px;
  }

  .dashboard-header {
    flex-direction: column;
    gap: 20px;
    text-align: center;
    padding: 25px;
  }

  .header-content h1 {
    font-size: 2rem;
  }

  .header-actions {
    flex-direction: row;
  }

  .stat-card {
    flex-direction: column;
    text-align: center;
    padding: 25px;
  }

  .holdings-section {
    padding: 20px;
  }

  .table-header,
  .table-row {
    grid-template-columns: 1fr 1fr 1fr;
    gap: 15px;
  }

  .table-header span:nth-child(n+4),
  .table-row span:nth-child(n+4) {
    display: none;
  }
}

@media (max-width: 480px) {
  .dashboard-header {
    padding: 20px;
  }

  .header-content h1 {
    font-size: 1.8rem;
  }

  .stat-card {
    padding: 20px;
  }

  .holdings-section {
    padding: 15px;
  }
}
</style>