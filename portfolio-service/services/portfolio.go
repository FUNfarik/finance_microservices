package services

import (
	"context"
	"fmt"
	"portfolio-service/database"
	grpcclient "portfolio-service/grpc-client"
	"portfolio-service/models"
)

type PortfolioService struct {
	db           *database.DB
	marketClient *grpcclient.MarketClient
}

func NewPortfolioService(db *database.DB, marketClient *grpcclient.MarketClient) *PortfolioService {
	return &PortfolioService{
		db:           db,
		marketClient: marketClient,
	}
}

// GetPortfolio retrieves a user's complete portfolio with current prices
func (s *PortfolioService) GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	// Get user's cash balance
	cash, err := s.db.GetUserCash(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user cash: %w", err)
	}

	// Get user's stock holdings
	holdings, err := s.db.GetAllUserHoldings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user holdings: %w", err)
	}

	// If no holdings, return portfolio with just cash
	if len(holdings) == 0 {
		return &models.Portfolio{
			UserID:     userID,
			Cash:       cash,
			TotalValue: cash,
			Holdings:   []models.Holding{},
		}, nil
	}

	// Get symbols for price lookup
	symbols := make([]string, len(holdings))
	for i, holding := range holdings {
		symbols[i] = holding.Symbol
	}

	// Get current prices from Market Data Service via gRPC
	prices, err := s.marketClient.GetMultipleStockPrices(ctx, symbols)
	if err != nil {
		fmt.Printf("Failed to get prices from market service: %v, falling back to database prices\n", err)
		// Fall back to database prices
		prices, err = s.db.GetStockPrices(symbols)
		if err != nil {
			fmt.Printf("Failed to get prices from database: %v, using average prices\n", err)
			prices = make(map[string]float64) // Empty map will cause fallback to avg prices
		}
	}

	// Calculate portfolio values
	totalValue := cash
	for i := range holdings {
		currentPrice, exists := prices[holdings[i].Symbol]
		if !exists {
			currentPrice = holdings[i].AvgPrice // Fallback to average price
		}

		holdings[i].CurrentPrice = currentPrice
		holdings[i].TotalValue = float64(holdings[i].Shares) * currentPrice
		holdings[i].GainLoss = holdings[i].TotalValue - (float64(holdings[i].Shares) * holdings[i].AvgPrice)

		totalValue += holdings[i].TotalValue
	}

	return &models.Portfolio{
		UserID:     userID,
		Cash:       cash,
		TotalValue: totalValue,
		Holdings:   holdings,
	}, nil
}

// BuyStock processes a stock purchase
func (s *PortfolioService) BuyStock(ctx context.Context, userID string, symbol string, shares int) error {
	if shares <= 0 {
		return fmt.Errorf("invalid shares amount: %d", shares)
	}

	// Validate symbol exists
	valid, err := s.marketClient.ValidateSymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to validate symbol: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid stock symbol: %s", symbol)
	}

	// Get current stock price
	price, _, err := s.marketClient.GetStockPrice(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get stock price: %w", err)
	}

	totalCost := float64(shares) * price

	// Check if user has enough cash
	cash, err := s.db.GetUserCash(userID)
	if err != nil {
		return fmt.Errorf("failed to get user cash: %w", err)
	}

	if cash < totalCost {
		return fmt.Errorf("insufficient funds: have $%.2f, need $%.2f", cash, totalCost)
	}

	// Start database transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update cash balance
	newCash := cash - totalCost
	err = s.db.UpdateUserCash(userID, newCash)
	if err != nil {
		return fmt.Errorf("failed to update cash: %w", err)
	}

	// Get existing holding
	existingHolding, err := s.db.GetUserHolding(userID, symbol)
	if err != nil {
		return fmt.Errorf("failed to get existing holding: %w", err)
	}

	// Calculate new holding values
	var newShares int
	var newAvgPrice float64

	if existingHolding == nil {
		// New holding
		newShares = shares
		newAvgPrice = price
	} else {
		// Update existing holding
		totalValue := (float64(existingHolding.Shares) * existingHolding.AvgPrice) + totalCost
		newShares = existingHolding.Shares + shares
		newAvgPrice = totalValue / float64(newShares)
	}

	// Update holdings
	err = s.db.UpsertHolding(userID, symbol, newShares, newAvgPrice)
	if err != nil {
		return fmt.Errorf("failed to update holding: %w", err)
	}

	// Record transaction
	err = s.db.CreateTransaction(userID, symbol, shares, price, "BUY", totalCost)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Successfully bought %d shares of %s for $%.2f\n", shares, symbol, totalCost)
	return nil
}

// SellStock processes a stock sale
func (s *PortfolioService) SellStock(ctx context.Context, userID string, symbol string, shares int) error {
	if shares <= 0 {
		return fmt.Errorf("invalid shares amount: %d", shares)
	}

	// Get current holding
	holding, err := s.db.GetUserHolding(userID, symbol)
	if err != nil {
		return fmt.Errorf("failed to get holding: %w", err)
	}
	if holding == nil {
		return fmt.Errorf("no holdings found for symbol: %s", symbol)
	}
	if holding.Shares < shares {
		return fmt.Errorf("insufficient shares: have %d, trying to sell %d", holding.Shares, shares)
	}

	// Get current stock price
	price, _, err := s.marketClient.GetStockPrice(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get stock price: %w", err)
	}

	totalReceived := float64(shares) * price

	// Start database transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update cash balance
	cash, err := s.db.GetUserCash(userID)
	if err != nil {
		return fmt.Errorf("failed to get user cash: %w", err)
	}

	newCash := cash + totalReceived
	err = s.db.UpdateUserCash(userID, newCash)
	if err != nil {
		return fmt.Errorf("failed to update cash: %w", err)
	}

	// Update holdings
	newShares := holding.Shares - shares
	if newShares == 0 {
		// If selling all shares, set to 0
		err = s.db.UpsertHolding(userID, symbol, 0, holding.AvgPrice)
	} else {
		// Keep same average price when selling
		err = s.db.UpsertHolding(userID, symbol, newShares, holding.AvgPrice)
	}
	if err != nil {
		return fmt.Errorf("failed to update holding: %w", err)
	}

	// Record transaction
	err = s.db.CreateTransaction(userID, symbol, shares, price, "SELL", totalReceived)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Successfully sold %d shares of %s for $%.2f\n", shares, symbol, totalReceived)
	return nil
}

// GetTransactions retrieves user's transaction history
func (s *PortfolioService) GetTransactions(ctx context.Context, userID string) ([]models.Transaction, error) {
	return s.db.GetUserTransactions(userID)
}
