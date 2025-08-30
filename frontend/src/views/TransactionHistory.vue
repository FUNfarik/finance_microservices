<template>
  <div class="transaction-page">
    <!-- Header -->
    <div class="transaction-header">
      <div class="header-content">
        <h1>Transaction History</h1>
        <p class="welcome-text">View all your trading activity and portfolio changes</p>
      </div>
      <div class="header-actions">
        <router-link to="/dashboard" class="nav-button">
          ← Back to Dashboard
        </router-link>
        <button @click="refreshTransactions" :disabled="loading" class="refresh-btn">
          <span :class="{ 'spinning': loading }">🔄</span>
          {{ loading ? 'Loading...' : 'Refresh' }}
        </button>
      </div>
    </div>

    <!-- Summary Stats -->
    <div class="stats-grid cols-4">
      <div class="stat-card primary">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <h3>Total Transactions</h3>
          <p class="big-number">{{ transactions.length }}</p>
          <span class="subtext">All time</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💹</div>
        <div class="stat-content">
          <h3>Buy Orders</h3>
          <p class="big-number">{{ buyCount }}</p>
          <span class="subtext">Purchases made</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💰</div>
        <div class="stat-content">
          <h3>Sell Orders</h3>
          <p class="big-number">{{ sellCount }}</p>
          <span class="subtext">Sales completed</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💵</div>
        <div class="stat-content">
          <h3>Total Volume</h3>
          <p class="big-number">${{ formatCurrency(totalVolume) }}</p>
          <span class="subtext">Trading volume</span>
        </div>
      </div>
    </div>

    <!-- Filters and Controls -->
    <div class="filter-section">
      <div class="filter-header">
        <h2>Filter Transactions</h2>
      </div>

      <div class="filter-controls">
        <div class="filter-group">
          <label>Transaction Type</label>
          <select v-model="filterType" @change="applyFilters" class="filter-select">
            <option value="">All Types</option>
            <option value="buy">Buy Orders</option>
            <option value="sell">Sell Orders</option>
          </select>
        </div>

        <div class="filter-group">
          <label>Stock Symbol</label>
          <input
              v-model="filterSymbol"
              @input="applyFilters"
              placeholder="Filter by symbol (e.g., AAPL)"
              class="filter-input"
          />
        </div>

        <div class="filter-group">
          <label>Date Range</label>
          <select v-model="filterDate" @change="applyFilters" class="filter-select">
            <option value="">All Time</option>
            <option value="today">Today</option>
            <option value="week">This Week</option>
            <option value="month">This Month</option>
            <option value="year">This Year</option>
          </select>
        </div>

        <button @click="clearFilters" class="clear-filters-btn">
          Clear Filters
        </button>
      </div>
    </div>

    <!-- Transactions Table -->
    <div class="transactions-section">
      <div class="section-header">
        <h2>Transaction History</h2>
        <span class="transaction-count">{{ filteredTransactions.length }} transactions</span>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>Loading transactions...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-banner">
        <span>⚠️ Failed to load transactions: {{ error }}</span>
        <button @click="refreshTransactions" class="retry-btn">Retry</button>
      </div>

      <!-- Empty State -->
      <div v-else-if="transactions.length === 0" class="no-transactions">
        <div class="empty-state">
          <div class="empty-icon">📋</div>
          <h3>No Transactions Yet</h3>
          <p>Start trading to see your transaction history here</p>
          <router-link to="/trade" class="cta-button">Start Trading</router-link>
        </div>
      </div>

      <!-- Transactions Table -->
      <div v-else-if="filteredTransactions.length > 0" class="transactions-table">
        <div class="table-header">
          <span>Date & Time</span>
          <span>Type</span>
          <span>Symbol</span>
          <span>Shares</span>
          <span>Price</span>
          <span>Total Amount</span>
          <span>Status</span>
        </div>

        <div
            v-for="transaction in paginatedTransactions"
            :key="transaction.id"
            class="table-row"
            :class="{ 'buy-row': transaction.trade_type === 'buy', 'sell-row': transaction.trade_type === 'sell' }"
        >
          <span class="transaction-date">
            <div class="date-primary">{{ formatDate(transaction.created_at) }}</div>
            <div class="date-secondary">{{ formatTime(transaction.created_at) }}</div>
          </span>

          <span class="transaction-type" :class="transaction.trade_type">
            <span class="type-badge" :class="transaction.trade_type">
              {{ transaction.trade_type.toUpperCase() }}
            </span>
          </span>

          <span class="symbol">{{ transaction.symbol }}</span>

          <span class="shares">{{ formatNumber(transaction.shares) }}</span>

          <span class="price">${{ formatCurrency(transaction.price) }}</span>

          <span class="total-amount">
            <div class="amount-primary">${{ formatCurrency(transaction.total_amount) }}</div>
          </span>

          <span class="status">
            <span class="status-badge completed">Completed</span>
          </span>
        </div>
      </div>

      <!-- No Results After Filtering -->
      <div v-else class="no-results">
        <div class="empty-state">
          <div class="empty-icon">🔍</div>
          <h3>No Matching Transactions</h3>
          <p>Try adjusting your filters to see more results</p>
          <button @click="clearFilters" class="cta-button">Clear Filters</button>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="filteredTransactions.length > itemsPerPage" class="pagination">
        <button
            @click="previousPage"
            :disabled="currentPage === 1"
            class="pagination-btn"
        >
          ← Previous
        </button>

        <span class="pagination-info">
          Page {{ currentPage }} of {{ totalPages }}
          ({{ filteredTransactions.length }} total)
        </span>

        <button
            @click="nextPage"
            :disabled="currentPage === totalPages"
            class="pagination-btn"
        >
          Next →
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TransactionHistory',

  data() {
    return {
      transactions: [],
      filteredTransactions: [],
      loading: false,
      error: null,

      // Filters
      filterType: '',
      filterSymbol: '',
      filterDate: '',

      // Pagination
      currentPage: 1,
      itemsPerPage: 20
    }
  },

  computed: {
    buyCount() {
      return this.transactions.filter(t => t.trade_type === 'buy').length
    },

    sellCount() {
      return this.transactions.filter(t => t.trade_type === 'sell').length
    },

    totalVolume() {
      return this.transactions.reduce((sum, t) => sum + (t.total_amount || 0), 0)
    },

    totalPages() {
      return Math.ceil(this.filteredTransactions.length / this.itemsPerPage)
    },

    paginatedTransactions() {
      const start = (this.currentPage - 1) * this.itemsPerPage
      const end = start + this.itemsPerPage
      return this.filteredTransactions.slice(start, end)
    }
  },

  async mounted() {
    await this.loadTransactions()
  },

  methods: {
    async loadTransactions() {
      this.loading = true
      this.error = null

      try {
        // Get user token for authentication
        const token = localStorage.getItem('finance_token')
        if (!token) {
          throw new Error('No authentication token found')
        }

        const response = await fetch('/api/portfolio/transactions/', {  // Updated URL
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        })

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        const data = await response.json()

        // Sort transactions by date (newest first)
        this.transactions = (data.transactions || []).sort((a, b) =>
            new Date(b.created_at) - new Date(a.created_at)
        )

        this.applyFilters()

      } catch (err) {
        console.error('Error loading transactions:', err)
        this.error = err.message
      } finally {
        this.loading = false
      }
    },

    async refreshTransactions() {
      await this.loadTransactions()
    },

    applyFilters() {
      let filtered = [...this.transactions]

      // Filter by type
      if (this.filterType) {
        filtered = filtered.filter(t => t.trade_type === this.filterType)
      }

      // Filter by symbol
      if (this.filterSymbol) {
        const symbol = this.filterSymbol.toUpperCase().trim()
        filtered = filtered.filter(t => t.symbol.includes(symbol))
      }

      // Filter by date
      if (this.filterDate) {
        const now = new Date()
        const filterDate = new Date()

        switch (this.filterDate) {
          case 'today':
            filterDate.setHours(0, 0, 0, 0)
            break
          case 'week':
            filterDate.setDate(now.getDate() - 7)
            break
          case 'month':
            filterDate.setMonth(now.getMonth() - 1)
            break
          case 'year':
            filterDate.setFullYear(now.getFullYear() - 1)
            break
        }

        filtered = filtered.filter(t => new Date(t.created_at) >= filterDate)
      }

      this.filteredTransactions = filtered
      this.currentPage = 1 // Reset to first page when filtering
    },

    clearFilters() {
      this.filterType = ''
      this.filterSymbol = ''
      this.filterDate = ''
      this.applyFilters()
    },

    previousPage() {
      if (this.currentPage > 1) {
        this.currentPage--
      }
    },

    nextPage() {
      if (this.currentPage < this.totalPages) {
        this.currentPage++
      }
    },

    formatCurrency(amount) {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(amount || 0)
    },

    formatNumber(num) {
      return new Intl.NumberFormat('en-US').format(num || 0)
    },

    formatDate(dateString) {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      })
    },

    formatTime(dateString) {
      return new Date(dateString).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit'
      })
    }
  }
}
</script>

<style scoped>
.transaction-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

/* Header Section */
.transaction-header {
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
  gap: 12px;
  align-items: center;
}

.nav-button, .refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  transition: background-color 0.2s ease;
  cursor: pointer;
}

.nav-button {
  background: #6c757d;
  color: white;
}

.nav-button:hover {
  background: #545b62;
  text-decoration: none;
  color: white;
}

.refresh-btn {
  background: #007bff;
  color: white;
}

.refresh-btn:hover:not(:disabled) {
  background: #0056b3;
}

.refresh-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Stats Grid */
.stats-grid {
  display: grid;
  gap: 25px;
  margin-bottom: 40px;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

.stats-grid.cols-4 {
  grid-template-columns: repeat(4, 1fr);
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

/* Filter Section */
.filter-section {
  background: white;
  border-radius: 15px;
  padding: 25px 30px;
  margin-bottom: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.filter-header h2 {
  color: #333;
  margin: 0 0 20px 0;
  font-size: 1.3rem;
  font-weight: 600;
}

.filter-controls {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  align-items: end;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  font-weight: 500;
  color: #555;
  font-size: 0.9rem;
}

.filter-select, .filter-input {
  padding: 10px 12px;
  border: 2px solid #e1e5e9;
  border-radius: 6px;
  font-size: 0.95rem;
  transition: border-color 0.3s ease;
}

.filter-select:focus, .filter-input:focus {
  outline: none;
  border-color: #007bff;
}

.clear-filters-btn {
  padding: 10px 20px;
  background: #6c757d;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s ease;
  height: fit-content;
}

.clear-filters-btn:hover {
  background: #545b62;
}

/* Transactions Section */
.transactions-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
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

.transaction-count {
  color: #666;
  font-size: 0.9rem;
}

/* Loading and Error States */
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

.retry-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

.retry-btn:hover {
  background: #c82333;
}

/* Empty States */
.no-transactions, .no-results {
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
  border: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.cta-button:hover {
  text-decoration: none;
  color: white;
  background: #1e7e34;
}

/* Transactions Table */
.transactions-table {
  overflow-x: auto;
}

.table-header {
  display: grid;
  grid-template-columns: 1.5fr 0.8fr 0.8fr 0.8fr 1fr 1.2fr 0.8fr;
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
  grid-template-columns: 1.5fr 0.8fr 0.8fr 0.8fr 1fr 1.2fr 0.8fr;
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

.table-row.buy-row {
  border-left: 4px solid #28a745;
  padding-left: 16px;
}

.table-row.sell-row {
  border-left: 4px solid #dc3545;
  padding-left: 16px;
}

.transaction-date {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.date-primary {
  font-weight: 600;
  color: #333;
  font-size: 0.95rem;
}

.date-secondary {
  font-size: 0.85rem;
  color: #666;
}

.type-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
}

.type-badge.buy {
  background: #d4edda;
  color: #28a745;
}

.type-badge.sell {
  background: #f8d7da;
  color: #dc3545;
}

.symbol {
  font-weight: 700;
  color: #667eea;
  font-size: 1.1rem;
}

.shares, .price {
  font-weight: 500;
  color: #495057;
}

.total-amount {
  font-weight: 600;
  color: #333;
  font-size: 1rem;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.status-badge.completed {
  background: #d4edda;
  color: #28a745;
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #e9ecef;
}

.pagination-btn {
  padding: 10px 20px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}

.pagination-btn:hover:not(:disabled) {
  background: #0056b3;
}

.pagination-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.pagination-info {
  color: #666;
  font-weight: 500;
}

/* Responsive Design */
@media (max-width: 1200px) {
  .stats-grid.cols-4 {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .transaction-page {
    padding: 15px;
  }

  .transaction-header {
    flex-direction: column;
    gap: 20px;
    text-align: center;
    padding: 25px;
  }

  .header-content h1 {
    font-size: 2rem;
  }

  .stats-grid.cols-4 {
    grid-template-columns: 1fr;
  }

  .stat-card {
    flex-direction: column;
    text-align: center;
    padding: 25px;
  }

  .filter-controls {
    grid-template-columns: 1fr;
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

  .pagination {
    flex-direction: column;
    gap: 15px;
  }
}

@media (max-width: 480px) {
  .transaction-header {
    padding: 20px;
  }

  .header-content h1 {
    font-size: 1.8rem;
  }

  .stat-card {
    padding: 20px;
  }

  .transactions-section {
    padding: 15px;
  }
}
</style>