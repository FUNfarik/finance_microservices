package main

import (
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

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

		// После успешной проверки токена:
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
	http.HandleFunc("/profile", jwtMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Protected endpoint", "status": "success", "note": "You are authenticated!"}`))
	}))

	db, err := connectDB()
	if err != nil {
		fmt.Printf("Couldn't connect to PostgreSQL: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("Auth Service is running...")
	//  Implement login
	http.HandleFunc("/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Login page", "action": "login"}`))
		if r.Method == "POST" {
			login := r.FormValue("login")
			password := r.FormValue("password")
			// bcrypt.GenerateFromPassword
			var username, passwordHash string
			var userID int
			err := db.QueryRow("SELECT username, password_hash, id FROM users WHERE username = $1", login).Scan(&username, &passwordHash, &userID)
			if err != nil {
				fmt.Println("Username and Password are invalid")
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
			if err != nil {
				fmt.Printf("Couldn't login: %v\n", err)
				return
			}
			fmt.Printf("%s: Successfully logged in", username)
			token, err := generateJWT(userID, username)
			if err != nil {
				fmt.Printf("Couldn't generate JWT for %s : %v\n", username, err)
				return
			}
			w.Write([]byte(fmt.Sprintf(`{"message": "Successfully logged in", "token": "%s", "status": "success"}`, token)))
		}
	}))

	http.HandleFunc("/register", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Register page", "action": "register"}`))
		if r.Method == "POST" {
			// Getting the users registration JSON
			email := r.FormValue("email")
			login := r.FormValue("login")
			password := r.FormValue("password")

			// Checking that user can be created
			var username string
			err := db.QueryRow("SELECT username FROM users WHERE username = $1", login).Scan(&username)
			if err == nil {
				fmt.Println("This login is already taken")
				return
			}
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				fmt.Println("Password hashing failed")
				return
			}

			_, err = db.Exec("INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)", login, hashedPassword, email)
			if err != nil {
				fmt.Println("Failed to create user")
				return
			}
			w.Write([]byte(`{"message": "Successfully registered", "action": "login"}`))
		}
	}))

	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "OK", "service": "auth-service", "port": 8001, "database": "connected"}`))
	}))

	fmt.Println("Server is running on http://localhost:8001")
	fmt.Println("Available endpoints:")
	fmt.Println("- http://localhost:8001/login")
	fmt.Println("- http://localhost:8001/register")
	fmt.Println("- http://localhost:8001/health")

	http.ListenAndServe(":8001", nil)
}
