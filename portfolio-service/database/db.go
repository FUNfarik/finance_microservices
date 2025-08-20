package database

import (
	"context"
	"database/sql"
	"fmt"
	"portfolio-service/models"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func Connect() (*DB, error) {
	connStr := "postgres://admin:admin@localhost:5432/finance_db?sslmode=disable"

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	err = conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	fmt.Println("✅ Portfolio Service connected to database")
	return &DB{conn: conn}, nil
}

// Close db connection

func (db *DB) Close() error {
	return db.conn.Close()
}

// Get users cash
func (db *DB) GetUserCash(userID int) (float64, error) {
	var cash float64
	err := db.conn.QueryRow("SELECT cash FROM users WHERE id = $1", userID).Scan(&cash)
	if err != nil {
		return 0, fmt.Errorf("error getting cash from user: %v", err)
	}
	return cash, nil
}

func (db *DB) UpdateUserCash(userID int, newCash float64) error {
	_, err := db.conn.Exec("UPDATE users SET cash = $1 WHERE id = $2", newCash, userID)
	if err != nil {
		return fmt.Errorf("error updating cash to user: %v", err)
	}
	return nil
}

func (db *DB) GetUserHoldings(userID int, symbol string) (*models.Holding, error) {
	var holding models.Holding
	query := "SELECT symbol, shares, avg_price FROM holdings WHERE user_id = $1 AND symbol = $2"

	err := db.conn.QueryRow(query, userID, symbol).Scan(&holding.Symbol, &holding.Shares, &holding.AvgPrice)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting holdings from user: %v", err)
	}

	return &holding, nil
}

func (db *DB) UpdateUserHoldings(userID int, symbol string, shares int, avgPrice float64) error {
	query := `
       INSERT INTO holdings (user_id, symbol, shares, avg_price)
       VALUES ($1, $2, $3, $4)
       ON CONFLICT (user_id, symbol)
       DO UPDATE SET
           shares = $3,
           avg_price = $4
`
	_, err := db.conn.Exec(query, userID, symbol, shares, avgPrice)
	if err != nil {
		return fmt.Errorf("error updating holdings in user: %v", err)
	}
	return nil
}

func (db *DB) CreateTransaction(userID int, symbol string, shares int, price float64, transactionType string, totalAmount float64) error {
	query := `
       INSERT INTO transactions (user_id, symbol, shares, price, transaction_type, total_amount, created_at)
       VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.conn.Exec(query, userID, symbol, shares, price, transactionType, totalAmount, time.Now())
	if err != nil {
		return fmt.Errorf("error creating transaction: %v", err)
	}
	return nil
}

func (db *DB) GetUserTransaction(userID int) ([]models.Transaction, error) {
	query := `
		SELECT id, user_id, symbol, shares, price, transaction_type, total_amount, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
`

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting transactions from user: %v", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Symbol, &tx.Shares, &tx.Price,
			&tx.TransactionType, &tx.TotalAmount, &tx.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, nil)
}

// GetAllUserHoldings retrieves all stock holdings for a user
func (db *DB) GetAllUserHoldings(userID int) ([]models.Holding, error) {
	query := `
		SELECT symbol, shares, avg_price 
		FROM holdings 
		WHERE user_id = $1 AND shares > 0
	`

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query holdings: %w", err)
	}
	defer rows.Close()

	var holdings []models.Holding
	for rows.Next() {
		var holding models.Holding
		err := rows.Scan(&holding.Symbol, &holding.Shares, &holding.AvgPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}
		holdings = append(holdings, holding)
	}

	return holdings, nil
}

// GetUserHolding gets specific stock holding for user (rename your existing function)
func (db *DB) GetUserHolding(userID int, symbol string) (*models.Holding, error) {
	return db.GetUserHoldings(userID, symbol)
}

// UpsertHolding - rename your existing UpdateUserHoldings
func (db *DB) UpsertHolding(userID int, symbol string, shares int, avgPrice float64) error {
	return db.UpdateUserHoldings(userID, symbol, shares, avgPrice)
}

// GetUserTransactions - rename your existing function
func (db *DB) GetUserTransactions(userID int) ([]models.Transaction, error) {
	return db.GetUserTransaction(userID)
}
