package grpcclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/FUNfarik/finance_microservices/proto/go/market"
)

type MarketClient struct {
	conn   *grpc.ClientConn
	client pb.MarketDataServiceClient
}

// Connect establishes gRPC connection to Market Data Service
func Connect() (*MarketClient, error) {
	conn, err := grpc.NewClient("localhost:8005", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to market data service: %w", err)
	}

	client := pb.NewMarketDataServiceClient(conn)

	fmt.Println("Connected to Market Data Service via gRPC on port 8005")
	return &MarketClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *MarketClient) Close() error {
	return c.conn.Close()
}

// GetStockPrice calls your actual rust gRPC service
func (c *MarketClient) GetStockPrice(ctx context.Context, symbol string) (float64, string, error) {
	request := &pb.GetStockPriceRequest{
		Symbol: symbol,
	}

	response, err := c.client.GetStockPrice(ctx, request)
	if err != nil {
		return 0, "", fmt.Errorf("gRPC call failed: %w", err)
	}

	if !response.Success {
		return 0, "", fmt.Errorf("service error: %s", response.ErrorMessage)
	}

	fmt.Printf("Retrieved price for %s (%s): $%.2f\n", symbol, response.Name, response.CurrentPrice)
	return response.CurrentPrice, response.Name, nil
}

// GetMultipleStockPrices calls your GetMultipleStocks gRPC method
func (c *MarketClient) GetMultipleStockPrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	request := &pb.GetMultipleStocksRequest{
		Symbols: symbols,
	}

	response, err := c.client.GetMultipleStocks(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("service error: %s", response.ErrorMessage)
	}

	prices := make(map[string]float64)
	for _, stock := range response.Stocks {
		if stock.Success {
			prices[stock.Symbol] = stock.CurrentPrice
		}
	}

	fmt.Printf("Retrieved prices for %d symbols\n", len(prices))
	return prices, nil
}

// ValidateSymbol checks if a stock symbol exists
func (c *MarketClient) ValidateSymbol(ctx context.Context, symbol string) (bool, error) {
	_, _, err := c.GetStockPrice(ctx, symbol)
	return err == nil, nil
}
