<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Portfolio Dashboard</h1>
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

    <!-- Service Status -->
    <div class="test-section">
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
    <div class="stats">
      <div class="stat-card">
        <h3>Portfolio Value</h3>
        <p class="big-number">${{ formatCurrency(portfolio.total_value) }}</p>
        <span :class="['change', portfolio.total_gain_loss >= 0 ? 'positive' : 'negative']">
          {{ portfolio.total_gain_loss >= 0 ? '+' : '' }}${{ formatCurrency(portfolio.total_gain_loss) }}
          ({{ portfolio.total_gain_loss_percent?.toFixed(2) }}%)
        </span>
      </div>

      <div class="stat-card">
        <h3>Cash Balance</h3>
        <p class="big-number">${{ formatCurrency(portfolio.cash) }}</p>
        <span class="subtext">Available to trade</span>
      </div>

      <div class="stat-card">
        <h3>Total Net Worth</h3>
        <p class="big-number">${{ formatCurrency(totalNetWorth) }}</p>
        <span class="subtext">Portfolio + Cash</span>
      </div>

      <div class="stat-card">
        <h3>Holdings</h3>
        <p class="big-number">{{ portfolio.holdings?.length || 0 }}</p>
        <span class="subtext">Unique positions</span>
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
            <span>{{ holding.shares }}</span>
            <span>${{ formatCurrency(holding.avg_price) }}</span>
            <span class="current-price">
              ${{ formatCurrency(holding.current_price || 0) }}
              <span v-if="!holding.current_price" class="price-updating">📊</span>
            </span>
            <span>${{ formatCurrency(holding.market_value || 0) }}</span>
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
            <h3>No Holdings Yet</h3>
            <p>Start building your portfolio by trading stocks</p>
            <router-link to="/trade" class="cta-button">Start Trading</router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions">
      <router-link to="/trade" class="action-card">
        <div class="action-icon">🔄</div>
        <h3>Trade Stocks</h3>
        <p>Buy and sell stocks</p>
      </router-link>

      <div class="action-card" @click="showTransactionHistory = true">
        <div class="action-icon">📊</div>
        <h3>History</h3>
        <p>View past transactions</p>
      </div>

      <div class="action-card" @click="refreshData">
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
  max-width: 1200px;
  margin: 20px auto;
  padding: 20px;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
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
  padding: 8px 16px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s;
}

.refresh-btn:hover:not(:disabled) {
  background: #0056b3;
}

.refresh-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.last-update {
  font-size: 0.9em;
  color: #6c757d;
}

.test-section {
  margin-bottom: 30px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 10px;
  text-align: center;
}

.test-btn {
  padding: 10px 20px;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.3s;
}

.test-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.service-status {
  margin-top: 15px;
}

.status-item {
  margin: 5px 0;
  padding: 8px;
  border-radius: 4px;
}

.status-item.success {
  background: #d4edda;
  color: #155724;
}

.status-item.error {
  background: #f8d7da;
  color: #721c24;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin: 30px 0;
}

.stat-card {
  background: white;
  padding: 25px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  text-align: center;
  transition: transform 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.big-number {
  font-size: 32px;
  font-weight: bold;
  color: #007bff;
  margin: 10px 0;
}

.change {
  font-size: 14px;
  font-weight: bold;
}

.change.positive {
  color: #28a745;
}

.change.negative {
  color: #dc3545;
}

.subtext {
  font-size: 12px;
  color: #6c757d;
}

.error-banner {
  background: #f8d7da;
  color: #721c24;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.error-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #721c24;
}

.holdings-section {
  background: white;
  padding: 25px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.update-prices-btn {
  padding: 6px 12px;
  background: #17a2b8;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.loading {
  text-align: center;
  padding: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #007bff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.holdings-table {
  margin-top: 20px;
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1.2fr 1.2fr 0.8fr;
  gap: 15px;
  padding: 15px 0;
  border-bottom: 2px solid #dee2e6;
  font-weight: bold;
  color: #495057;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1.2fr 1.2fr 0.8fr;
  gap: 15px;
  padding: 15px 0;
  border-bottom: 1px solid #dee2e6;
  align-items: center;
}

.symbol {
  font-weight: bold;
  color: #007bff;
}

.current-price {
  display: flex;
  align-items: center;
  gap: 5px;
}

.price-updating {
  font-size: 12px;
  opacity: 0.7;
}

.gain-loss {
  font-weight: bold;
}

.gain-loss.positive {
  color: #28a745;
}

.gain-loss.negative {
  color: #dc3545;
}

.actions {
  display: flex;
  gap: 5px;
}

.action-btn {
  padding: 4px 8px;
  font-size: 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
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
  padding: 60px 20px;
}

.empty-state h3 {
  color: #495057;
  margin-bottom: 10px;
}

.empty-state p {
  color: #6c757d;
  margin-bottom: 20px;
}

.cta-button {
  display: inline-block;
  padding: 12px 24px;
  background: #28a745;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  font-weight: bold;
  transition: all 0.3s;
}

.cta-button:hover {
  background: #218838;
  text-decoration: none;
  transform: translateY(-2px);
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-top: 40px;
}

.action-card {
  background: white;
  padding: 25px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  text-decoration: none;
  color: inherit;
}

.action-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 20px rgba(0,0,0,0.15);
  text-decoration: none;
  color: inherit;
}

.action-icon {
  font-size: 48px;
  margin-bottom: 15px;
}

.action-card h3 {
  margin: 0 0 10px 0;
  color: #333;
}

.action-card p {
  margin: 0;
  color: #6c757d;
  font-size: 14px;
}

h1 {
  color: #333;
  margin: 0;
}

h2 {
  color: #495057;
  margin: 0;
}

@media (max-width: 768px) {
  .dashboard {
    padding: 10px;
  }

  .dashboard-header {
    flex-direction: column;
    gap: 15px;
    text-align: center;
  }

  .stats {
    grid-template-columns: 1fr;
  }

  .table-header,
  .table-row {
    grid-template-columns: 1fr 1fr 1fr;
    gap: 10px;
  }

  .table-header span:nth-child(n+4),
  .table-row span:nth-child(n+4) {
    display: none;
  }

  .quick-actions {
    grid-template-columns: 1fr;
  }
}
</style>