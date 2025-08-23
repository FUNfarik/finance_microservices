<template>
  <div class="login-page">
    <h1>Login to CS50 Finance</h1>

    <!-- Test Backend Connection -->
    <div class="test-section">
      <button @click="testBackend" class="test-btn">
        Test Backend Connection
      </button>
      <p v-if="connectionStatus" :class="connectionClass">
        {{ connectionStatus }}
      </p>
    </div>

    <form @submit.prevent="handleLogin" class="login-form">
      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <div class="form-group">
        <label>Email:</label>
        <input
            type="email"
            v-model="email"
            required
            placeholder="Enter your email"
        >
      </div>

      <div class="form-group">
        <label>Password:</label>
        <input
            type="password"
            v-model="password"
            required
            placeholder="Enter your password"
        >
      </div>

      <button type="submit" class="login-btn" :disabled="loading">
        {{ loading ? 'Logging in...' : 'Login' }}
      </button>
    </form>

    <p>
      <router-link to="/dashboard">Go to Dashboard (for testing)</router-link>
    </p>
  </div>
</template>

<script>
import authService from '../services/authService.js'

export default {
  data() {
    return {
      email: '',
      password: '',
      loading: false,
      error: null,
      connectionStatus: null
    }
  },
  computed: {
    connectionClass() {
      return {
        'success-text': this.connectionStatus && this.connectionStatus.includes('✅'),
        'error-text': this.connectionStatus && this.connectionStatus.includes('❌')
      }
    }
  },
  methods: {
    async testBackend() {
      try {
        this.connectionStatus = 'Testing connection...'
        await authService.testConnection()
        this.connectionStatus = 'Backend connection successful!'
      } catch (error) {
        this.connectionStatus = `Backend connection failed: ${error.message || 'Unknown error'}`
      }
    },

    async handleLogin() {
      this.loading = true
      this.error = null

      try {
        const response = await authService.login(this.email, this.password)
        console.log('Login successful:', response)
        this.$router.push('/dashboard')
      } catch (error) {
        this.error = error.message || 'Login failed'
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style>
.login-page {
  max-width: 400px;
  margin: 50px auto;
  padding: 20px;
  text-align: center;
}

.test-section {
  margin-bottom: 30px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.test-btn {
  padding: 10px 20px;
  background: #28a745;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  margin-bottom: 10px;
}

.test-btn:hover {
  background: #218838;
}

.success-text {
  color: #28a745;
  font-weight: bold;
}

.error-text {
  color: #dc3545;
  font-weight: bold;
}

.error-message {
  background: #f8d7da;
  color: #721c24;
  padding: 10px;
  border-radius: 5px;
  margin-bottom: 20px;
  border: 1px solid #f5c6cb;
}

.login-form {
  background: white;
  padding: 30px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.form-group {
  margin-bottom: 20px;
  text-align: left;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 16px;
  box-sizing: border-box;
}

.login-btn {
  width: 100%;
  padding: 12px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 5px;
  font-size: 16px;
  cursor: pointer;
}

.login-btn:hover:not(:disabled) {
  background: #0056b3;
}

.login-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

h1 {
  color: #333;
  margin-bottom: 30px;
}
</style>