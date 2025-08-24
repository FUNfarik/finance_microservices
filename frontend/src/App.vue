<template>
  <div id="app">
    <!-- Navigation Header -->
    <nav v-if="showNavigation" class="main-nav">
      <div class="nav-container">
        <h1 class="app-title">CS50 Finance</h1>
        <div class="nav-links">
          <router-link to="/dashboard" class="nav-link">Dashboard</router-link>
          <router-link to="/trade" class="nav-link">Trade</router-link>
          <button @click="logout" class="logout-btn">Logout</button>
        </div>
      </div>
    </nav>

    <!-- Main Content Area -->
    <main :class="['main-content', { 'with-nav': showNavigation, 'full-screen': !showNavigation }]">
      <router-view />
    </main>
  </div>
</template>

<script>
export default {
  name: 'App',
  computed: {
    // Hide navigation on login page
    showNavigation() {
      return this.$route.path !== '/login'
    }
  },
  methods: {
    logout() {
      // Clear token and redirect to login
      localStorage.removeItem('finance_token')
      localStorage.removeItem('finance_user')
      this.$router.push('/login')
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  min-height: 100vh;
}

.main-nav {
  background: #343a40;
  color: white;
  padding: 1rem 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: relative;
  z-index: 1000;
}

.nav-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.app-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-link {
  color: #adb5bd;
  text-decoration: none;
  padding: 0.75rem 1.25rem;
  border-radius: 8px;
  transition: all 0.3s ease;
  font-weight: 500;
  position: relative;
}

.nav-link:hover {
  color: white;
  background: rgba(255, 255, 255, 0.1);
}

.nav-link.router-link-active {
  color: white;
  background: #007bff;
}

.logout-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 0.75rem 1.25rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

.logout-btn:hover {
  background: #c82333;
}

/* Main Content Styles */
.main-content {
  position: relative;
}

/* When navigation is present */
.main-content.with-nav {
  min-height: calc(100vh - 80px);
  padding-top: 0;
}

/* Full screen for login page */
.main-content.full-screen {
  min-height: 100vh;
  padding: 0;
}

/* Special handling for dashboard route */
.main-content.with-nav:has(.dashboard) {
  padding: 0;
  background: transparent;
}

/* Default background for non-dashboard pages with navigation */
.main-content.with-nav {
  background: #f8f9fa;
  padding: 2rem;
}

/* Override for dashboard specifically */
:deep(.dashboard) {
  margin: 0;
  padding-top: 20px;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .nav-container {
    padding: 0 1rem;
    flex-direction: column;
    gap: 1rem;
  }

  .app-title {
    font-size: 1.3rem;
  }

  .nav-links {
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.5rem;
  }

  .nav-link,
  .logout-btn {
    padding: 0.5rem 1rem;
    font-size: 0.9rem;
  }

  .main-content.with-nav {
    min-height: calc(100vh - 120px);
    padding: 1rem;
  }
}

@media (max-width: 480px) {
  .nav-container {
    padding: 0 0.5rem;
  }

  .main-content.with-nav {
    padding: 0.5rem;
  }
}

/* Ensure dashboard gets proper styling */
body {
  margin: 0;
  padding: 0;
}
</style>