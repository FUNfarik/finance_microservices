<template>
  <div class="auth-page">
    <div class="auth-container">
      <h1>{{ isLogin ? 'Welcome Back' : 'Create Account' }}</h1>
      <p class="subtitle">
        {{ isLogin ? 'Sign in to your CS50 Finance account' : 'Join CS50 Finance today' }}
      </p>

      <form @submit.prevent="handleSubmit" class="auth-form">
        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <div v-if="success" class="success-message">
          {{ success }}
        </div>

        <!-- Registration-only fields -->
        <div v-if="!isLogin" class="form-row">
          <div class="form-group">
            <label>First Name:</label>
            <input
                type="text"
                v-model="firstName"
                :required="!isLogin"
                placeholder="Enter your first name"
                autocomplete="given-name"
            >
          </div>
          <div class="form-group">
            <label>Last Name:</label>
            <input
                type="text"
                v-model="lastName"
                :required="!isLogin"
                placeholder="Enter your last name"
                autocomplete="family-name"
            >
          </div>
        </div>

        <div class="form-group">
          <label>Email:</label>
          <input
              type="email"
              v-model="email"
              required
              placeholder="Enter your email"
              autocomplete="email"
          >
        </div>

        <div class="form-group">
          <label>Password:</label>
          <input
              type="password"
              v-model="password"
              required
              :placeholder="isLogin ? 'Enter your password' : 'Create a password'"
              :autocomplete="isLogin ? 'current-password' : 'new-password'"
          >
          <div v-if="!isLogin" class="password-requirements">
            <small>Password must be at least 8 characters long</small>
          </div>
        </div>

        <div v-if="!isLogin" class="form-group">
          <label>Confirm Password:</label>
          <input
              type="password"
              v-model="confirmPassword"
              :required="!isLogin"
              placeholder="Confirm your password"
              autocomplete="new-password"
          >
        </div>

        <button type="submit" class="auth-btn" :disabled="loading">
          {{ loading ? (isLogin ? 'Signing in...' : 'Creating account...') : (isLogin ? 'Sign In' : 'Create Account') }}
        </button>
      </form>

      <div class="auth-switch">
        <p>
          {{ isLogin ? "Don't have an account?" : "Already have an account?" }}
          <button @click="toggleMode" class="link-btn" type="button">
            {{ isLogin ? 'Sign up' : 'Sign in' }}
          </button>
        </p>
      </div>
    </div>
  </div>
</template>

<script>
import authService from '../services/authService.js'

export default {
  name: 'LoginRegister',
  data() {
    return {
      isLogin: true,
      email: '',
      password: '',
      confirmPassword: '',
      firstName: '',
      lastName: '',
      loading: false,
      error: null,
      success: null
    }
  },
  methods: {
    toggleMode() {
      this.isLogin = !this.isLogin
      this.clearMessages()
      this.clearForm()
    },

    clearMessages() {
      this.error = null
      this.success = null
    },

    clearForm() {
      this.email = ''
      this.password = ''
      this.confirmPassword = ''
      this.firstName = ''
      this.lastName = ''
    },

    validateForm() {
      if (!this.isLogin) {
        if (!this.firstName.trim()) {
          this.error = 'First name is required'
          return false
        }
        if (!this.lastName.trim()) {
          this.error = 'Last name is required'
          return false
        }
        if (this.password.length < 8) {
          this.error = 'Password must be at least 8 characters long'
          return false
        }
        if (this.password !== this.confirmPassword) {
          this.error = 'Passwords do not match'
          return false
        }
      }

      if (!this.email.trim()) {
        this.error = 'Email is required'
        return false
      }
      if (!this.password) {
        this.error = 'Password is required'
        return false
      }

      return true
    },

    async handleSubmit() {
      this.clearMessages()

      if (!this.validateForm()) {
        return
      }

      this.loading = true

      try {
        if (this.isLogin) {
          const response = await authService.login(this.email, this.password)
          console.log('Login successful:', response)

          // Store user info for the app
          if (response.user) {
            localStorage.setItem('finance_user', JSON.stringify(response.user))
          }

          // Redirect to dashboard
          this.$router.push('/dashboard')
        } else {
          const userData = {
            email: this.email.trim(),
            password: this.password,
            firstName: this.firstName.trim(),
            lastName: this.lastName.trim()
          }

          const response = await authService.register(userData)
          console.log('Registration successful:', response)

          this.success = 'Account created successfully! Please sign in.'
          this.isLogin = true
          this.clearForm()
        }
      } catch (error) {
        console.error('Auth error:', error)
        this.error = error.message || (this.isLogin ? 'Login failed' : 'Registration failed')
      } finally {
        this.loading = false
      }
    }
  },

  mounted() {
    // Clear any existing error messages when component mounts
    this.clearMessages()

    // Check if user is already logged in
    if (authService.isAuthenticated()) {
      this.$router.push('/dashboard')
    }
  },

  beforeUnmount() {
    // Clear any pending timeouts or intervals if needed
    this.clearMessages()
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.auth-container {
  max-width: 450px;
  width: 100%;
  background: white;
  padding: 40px;
  border-radius: 15px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  text-align: center;
}

h1 {
  color: #333;
  margin-bottom: 10px;
  font-size: 2rem;
  font-weight: 600;
}

.subtitle {
  color: #666;
  margin-bottom: 30px;
  font-size: 1rem;
}

.auth-form {
  text-align: left;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
}

input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.3s ease;
  box-sizing: border-box;
}

input:focus {
  outline: none;
  border-color: #667eea;
}

.password-requirements {
  margin-top: 5px;
}

.password-requirements small {
  color: #666;
  font-size: 0.85rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #fcc;
  font-size: 0.9rem;
}

.success-message {
  background: #efe;
  color: #363;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #cfc;
  font-size: 0.9rem;
}

.auth-btn {
  width: 100%;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: transform 0.2s ease;
  margin-bottom: 20px;
}

.auth-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.auth-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}

.auth-switch {
  text-align: center;
  margin: 20px 0;
  padding: 20px 0;
  border-top: 1px solid #eee;
}

.auth-switch p {
  margin: 0;
  color: #666;
}

.link-btn {
  background: none;
  border: none;
  color: #667eea;
  cursor: pointer;
  font-weight: 500;
  text-decoration: underline;
  margin-left: 5px;
  font-size: inherit;
}

.link-btn:hover {
  color: #764ba2;
}

.link-btn:focus {
  outline: 2px solid #667eea;
  outline-offset: 2px;
}

/* Responsive design */
@media (max-width: 500px) {
  .auth-container {
    padding: 30px 20px;
    margin: 10px;
  }

  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }

  h1 {
    font-size: 1.5rem;
  }
}

/* Loading state improvements */
.auth-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* Focus states for better accessibility */
input:focus,
.link-btn:focus,
.auth-btn:focus {
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}
</style>