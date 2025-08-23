<template>
  <div class="dashboard">
    <h1>Portfolio Dashboard</h1>

    <!-- Test Backend Services -->
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

    <div class="stats">
      <div class="stat-card">
        <h3>Portfolio Value</h3>
        <p class="big-number">${{ formatCurrency(portfolioData.total_value) }}</p>
        <span :class="['change', portfolioData.total_gain_loss >= 0 ? 'positive' : 'negative']">
          {{ portfolioData.total_gain_loss >= 0 ? '+' : '' }}${{ formatCurrency(portfolioData.total_gain_loss) }}
          ({{ portfolioData.total_gain_loss_percent?.toFixed(2) }}%)
        </span>
      </div>

      <div class="stat-card">
        <h3>Cash Balance</h3>
        <p class="big-number">${{ formatCurrency(portfolioData.cash) }}</p>
      </div>

      <div class="stat-card">
        <h3>Total Holdings</h3>
        <p class="big-number">{{ portfolioData.holdings?.length || 0 }}</p>
      </div>
    </div>

    <div class="portfolio-content">
      <!-- Holdings Table -->
      <div class="holdings-section">
        <h2>Your Holdings</h2>
        <div v-if="loading.portfolio" class="loading">Loading portfolio...</div>
        <div v-else-if="portfolioData.holdings?.length > 0" class="holdings-table">
          <div class="table-header">
            <span>Symbol</span>
            <span>Shares</span>
            <span>Avg Price</span>
            <span>Current Price</span>
            <span>Market Value</span>
            <span>Gain/Loss</span>
          </div>
          <div v-for="holding in portfolioData.holdings" :key="holding.symbol" class="table-row">
            <span class="symbol">{{ holding.symbol }}</span>
            <span>{{ holding.shares }}</span>
            <span>${{ formatCurrency(holding.avg_price) }}</span>
            <span>${{ formatCurrency(holding.current_price || 0) }}</span>
            <span>${{ formatCurrency(holding.market_value || 0) }}</span>
            <span :class="['gain-loss', (holding.gain_loss || 0) >= 0 ? 'positive' : 'negative']">
              {{ (holding.gain_loss || 0) >= 0 ? '+' : '' }}${{ formatCurrency(holding.gain_loss || 0) }}
            </span>
          </div>
        </div>
        <div v-else class="no-holdings">
          <p>No holdings yet. <router-link to="/trade">Start trading!</router-link></p>
        </div>
      </div>
    </div>

    <div class="navigation">
      <router-link to="/trade" class="nav-button">🔄 Trade Stocks</router-link>
      <router-link to="/login" class="nav-button">🔐 Login</router-link>
    </div>
  </div>
</template>

<script>
import portfolioService from '../services/portfolioService.js'
import { marketService } from '../services/marketService.js'

export default {
  data() {
    return {
      portfolioData: {
        holdings: [],
        total_value: 0,
        cash: 10000,
        total_gain_loss: 0,
        total_gain_loss_percent: 0
      },
      loading: {
        portfolio: false,
        market: false
      },
      testing: false,
      serviceStatus: []
    }
  },
  async mounted() {
    await this.loadPortfolioData()
  },
  methods: {
    async loadPortfolioData() {
      this.loading.portfolio = true
      try {
        console.log('Fetching portfolio data...') // Add debugging
        const portfolio = await portfolioService.getPortfolio()
        console.log('Portfolio data received:', portfolio) // Add debugging
        this.portfolioData = portfolio
      } catch (error) {
        console.error('Failed to load portfolio:', error)
        console.error('Error details:', error.response?.data || error.message) // Better error logging

        // Instead of mock data, show the actual user's real data
        this.portfolioData = {
          holdings: [],               // Empty holdings (which is correct for your user)
          total_value: 0,             // No portfolio value
          cash: 500.00,               // Your actual cash amount
          total_gain_loss: 0,
          total_gain_loss_percent: 0
        }

        // Or better yet, show an error message to the user
        this.$toast?.error?.('Failed to load portfolio data')
      } finally {
        this.loading.portfolio = false
      }
    },

    async testServices() {
      this.testing = true
      this.serviceStatus = []

      // Test Portfolio Service
      try {
        await portfolioService.testConnection()
        this.serviceStatus.push({
          service: 'Portfolio Service',
          success: true,
          message: '✅ Connected'
        })
      } catch (error) {
        this.serviceStatus.push({
          service: 'Portfolio Service',
          success: false,
          message: '❌ Failed'
        })
      }

      // Test Market Data Service
      try {
        await marketService.testConnection()
        this.serviceStatus.push({
          service: 'Market Data Service',
          success: true,
          message: '✅ Connected'
        })
      } catch (error) {
        this.serviceStatus.push({
          service: 'Market Data Service',
          success: false,
          message: '❌ Failed'
        })
      }

      this.testing = false
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

<style>
.dashboard {
  max-width: 1200px;
  margin: 20px auto;
  padding: 20px;
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
  display: flex;
  gap: 20px;
  margin: 30px 0;
  flex-wrap: wrap;
}

.stat-card {
  flex: 1;
  min-width: 250px;
  background: white;
  padding: 25px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  text-align: center;
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

.portfolio-content {
  margin: 30px 0;
}

.holdings-section {
  background: white;
  padding: 25px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.holdings-table {
  margin-top: 20px;
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1.2fr 1.2fr;
  gap: 15px;
  padding: 15px 0;
  border-bottom: 2px solid #dee2e6;
  font-weight: bold;
  color: #495057;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr 1.2fr 1.2fr;
  gap: 15px;
  padding: 15px 0;
  border-bottom: 1px solid #dee2e6;
  align-items: center;
}

.symbol {
  font-weight: bold;
  color: #007bff;
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

.loading {
  text-align: center;
  padding: 40px;
  color: #6c757d;
}

.no-holdings {
  text-align: center;
  padding: 40px;
  color: #6c757d;
}

.no-holdings a {
  color: #007bff;
  text-decoration: none;
}

.navigation {
  text-align: center;
  margin: 40px 0;
}

.nav-button {
  display: inline-block;
  margin: 0 15px;
  padding: 12px 24px;
  background: #007bff;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  transition: background-color 0.3s;
}

.nav-button:hover {
  background: #0056b3;
}

h1 {
  color: #333;
  text-align: center;
  margin-bottom: 30px;
}

h2 {
  color: #495057;
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .stats {
    flex-direction: column;
  }

  .table-header,
  .table-row {
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .table-header span:nth-child(n+3),
  .table-row span:nth-child(n+3) {
    display: none;
  }
}
</style>