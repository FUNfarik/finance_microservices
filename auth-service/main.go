package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// JSON request structures
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// enable CORS for browser to pass the policy
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			// Default for development
			allowedOrigins = "http://localhost:3000"
		}

		// You can support multiple origins separated by commas
		origins := strings.Split(allowedOrigins, ",")
		origin := r.Header.Get("Origin")

		// Check if the request origin is allowed
		allowOrigin := ""
		for _, allowedOrigin := range origins {
			if strings.TrimSpace(allowedOrigin) == origin {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		userID := int(claims["user_id"].(float64))
		username := claims["username"].(string)

		fmt.Printf("User %s (ID: %d) authenticated\n", username, userID)
		next(w, r)
	}
}

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "default-secret-key"
	}
	return secret
}

func connectDB() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "admin"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "admin"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "finance_db"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL")
	return db, nil
}

func generateJWT(userID int, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func main() {
	db, err := connectDB()
	if err != nil {
		fmt.Printf("Couldn't connect to PostgreSQL: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("Auth Service is running...")

	// Protected profile endpoint
	http.HandleFunc("/profile", enableCORS(jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Protected endpoint", "status": "success", "note": "You are authenticated!"}`))
	})))

	// Login endpoint - Updated to handle JSON
	http.HandleFunc("/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			w.Write([]byte(`{"message": "Login endpoint", "method": "POST"}`))
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error": "Method not allowed", "status": "error"}`))
			return
		}

		// Read and parse JSON request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Could not read request body", "status": "error"}`))
			return
		}

		var loginReq LoginRequest
		err = json.Unmarshal(body, &loginReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid JSON format", "status": "error"}`))
			return
		}

		// Validate input
		if loginReq.Email == "" || loginReq.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Email and password are required", "status": "error"}`))
			return
		}

		// Query user by email
		var username, passwordHash string
		var userID int
		var cash float64

		err = db.QueryRow("SELECT id, username, password_hash, cash FROM users WHERE email = $1", loginReq.Email).Scan(&userID, &username, &passwordHash, &cash)
		if err != nil {
			fmt.Printf("Login attempt failed for email: %s - %v\n", loginReq.Email, err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid email or password", "status": "error"}`))
			return
		}

		// Check password
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginReq.Password))
		if err != nil {
			fmt.Printf("Password verification failed for user: %s\n", username)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid email or password", "status": "error"}`))
			return
		}

		// Generate JWT token
		token, err := generateJWT(userID, username)
		if err != nil {
			fmt.Printf("JWT generation failed for user %s: %v\n", username, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Could not generate token", "status": "error"}`))
			return
		}

		// Success response
		fmt.Printf("User %s (ID: %d) logged in successfully\n", username, userID)

		response := fmt.Sprintf(`{
			"message": "Login successful", 
			"status": "success",
			"token": "%s",
			"user": {
				"id": %d,
				"username": "%s",
				"email": "%s",
				"cash": %.2f
			}
		}`, token, userID, username, loginReq.Email, cash)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))

	// Register endpoint - Updated to handle JSON
	http.HandleFunc("/register", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			w.Write([]byte(`{"message": "Register endpoint", "method": "POST"}`))
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error": "Method not allowed", "status": "error"}`))
			return
		}

		// Read and parse JSON request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Could not read request body", "status": "error"}`))
			return
		}

		var registerReq RegisterRequest
		err = json.Unmarshal(body, &registerReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid JSON format", "status": "error"}`))
			return
		}

		// Validate input
		if registerReq.Username == "" || registerReq.Email == "" || registerReq.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Username, email and password are required", "status": "error"}`))
			return
		}

		// Check if user already exists
		var existingUsername string
		err = db.QueryRow("SELECT username FROM users WHERE username = $1 OR email = $2", registerReq.Username, registerReq.Email).Scan(&existingUsername)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error": "Username or email already exists", "status": "error"}`))
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerReq.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Password hashing failed", "status": "error"}`))
			return
		}

		// Insert new user with default cash amount
		var newUserID int
		err = db.QueryRow("INSERT INTO users (username, password_hash, email, cash, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id",
			registerReq.Username, hashedPassword, registerReq.Email, 10000.00).Scan(&newUserID)
		if err != nil {
			fmt.Printf("Failed to create user: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to create user", "status": "error"}`))
			return
		}

		fmt.Printf("✅ New user registered: %s (ID: %d)\n", registerReq.Username, newUserID)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "Successfully registered", "status": "success"}`))
	}))

	// Health check endpoint
	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "OK", "service": "auth-service", "port": 8001, "database": "connected"}`))
	}))

	fmt.Println("Server is running on http://localhost:8001")
	fmt.Println("Available endpoints:")
	fmt.Println("- http://localhost:8001/login")
	fmt.Println("- http://localhost:8001/register")
	fmt.Println("- http://localhost:8001/health")
	fmt.Println("- http://localhost:8001/profile")

	http.ListenAndServe(":8001", nil)
}
