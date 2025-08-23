<template>
  <div id="app">
    <!-- Navigation Header (optional) -->
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

    <!-- Main Content Area - This is where your pages will show -->
    <main class="main-content">
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
  background: #f8f9fa;
}

.main-nav {
  background: #343a40;
  color: white;
  padding: 1rem 0;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
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
  font-weight: bold;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-link {
  color: #adb5bd;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  transition: all 0.3s;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: white;
  background: #495057;
}

.logout-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.logout-btn:hover {
  background: #c82333;
}

.main-content {
  min-height: calc(100vh - 80px);
  padding: 2rem;
}

/* Remove navigation padding when no nav is shown */
body:has(.main-nav) .main-content {
  min-height: calc(100vh - 80px);
}

body:not(:has(.main-nav)) .main-content {
  min-height: 100vh;
  padding: 0;
}
</style>