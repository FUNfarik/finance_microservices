use serde::{Deserialize, Serialize};
use warp::Filter;
use std::env;
use serde_json;
use tonic::{transport::Server, Request, Response, Status};
use futures::future::join_all;
use std::convert::Infallible;

pub mod market_proto {
    tonic::include_proto!("market");
}

// Create CORS filter for Warp
fn cors_filter() -> warp::filters::cors::Cors {
    let allowed_origins = env::var("ALLOWED_ORIGINS")
        .unwrap_or_else(|_| "http://localhost:3000".to_string());

    if allowed_origins == "*" {
        // For development - allow all origins
        warp::cors()
            .allow_any_origin()
            .allow_methods(vec!["GET", "POST", "PUT", "DELETE", "OPTIONS"])
            .allow_headers(vec!["content-type", "authorization", "accept", "origin"])
            .build()
    } else {
        // For production - only allowed origins
        let origins: Vec<&str> = allowed_origins
            .split(',')
            .map(|s| s.trim())
            .filter(|s| !s.is_empty())
            .collect();

        warp::cors()
            .allow_origins(origins)
            .allow_methods(vec!["GET", "POST", "PUT", "DELETE", "OPTIONS"])
            .allow_headers(vec!["content-type", "authorization", "accept", "origin"])
            .allow_credentials(true)
            .build()
    }
}

// Import the server components and message types from the module
use market_proto::{
    market_data_service_server::{MarketDataService, MarketDataServiceServer},
    GetStockPriceRequest, GetStockPriceResponse,
    GetMultipleStocksRequest, GetMultipleStocksResponse,
};

#[derive(Serialize, Deserialize, Debug)]
struct Stock {
    symbol: String,
    name: String,
    price: f64,
    change_percent: f64,
}

#[derive(Deserialize, Debug)]
struct AlphaVantageResponse {
    #[serde(rename = "Global Quote")]
    global_quote: GlobalQuote,
}

#[derive(Deserialize, Debug)]
struct GlobalQuote {
    #[serde(rename = "01. symbol")]
    symbol: String,
    #[serde(rename = "05. price")]
    price: String,
    #[serde(rename = "10. change percent")]
    change_percent: String,
}

#[derive(Default)]
pub struct MyMarketDataService;

#[tonic::async_trait]
impl MarketDataService for MyMarketDataService {
    async fn get_stock_price(
        &self,
        request: Request<GetStockPriceRequest>,
    ) -> Result<Response<GetStockPriceResponse>, Status> {
        let symbol = request.into_inner().symbol;
        println!("gRPC request for symbol {}", symbol);

        match fetch_stock_data_internal(symbol.clone()).await {
            Ok(stock) => {
                let response = GetStockPriceResponse {
                    symbol: stock.symbol,
                    name: stock.name,
                    current_price: stock.price,
                    change_percent: stock.change_percent,
                    success: true,
                    error_message: String::new()
                };
                Ok(Response::new(response))
            }
            Err(e) => {
                let response = GetStockPriceResponse {
                    symbol,
                    name: String::new(),
                    current_price: 0.0,
                    change_percent: 0.0,
                    success: false,
                    error_message: e,
                };
                Ok(Response::new(response))
            }
        }
    }

    async fn get_multiple_stocks(
        &self,
        request: Request<GetMultipleStocksRequest>,
    ) -> Result<Response<GetMultipleStocksResponse>, Status> {
        let symbols = request.into_inner().symbols;
        println!("gRPC request for symbols {:?}", symbols);

        let futures: Vec<_> = symbols.into_iter()
            .map(|symbol| {
                let symbol_clone = symbol.clone();
                async move {
                    let symbol_for_error = symbol.clone();
                    match fetch_stock_data_internal(symbol_clone).await {
                        Ok(stock) => GetStockPriceResponse {
                            symbol: stock.symbol,
                            name: stock.name,
                            current_price: stock.price,
                            change_percent: stock.change_percent,
                            success: true,
                            error_message: String::new(),
                        },
                        Err(e) => GetStockPriceResponse {
                            symbol: symbol_for_error,
                            name: String::new(),
                            current_price: 0.0,
                            change_percent: 0.0,
                            success: false,
                            error_message: e,
                        }
                    }
                }
            })
            .collect();

        let stocks = join_all(futures).await;

        let response = GetMultipleStocksResponse {
            stocks,
            success: true,
            error_message: String::new(),
        };

        Ok(Response::new(response))
    }
}

async fn fetch_stock_data_internal(symbol: String) -> Result<Stock, String> {
    let api_key = env::var("ALPHA_API").unwrap_or_else(|_| "demo".to_string());
    let url = format!("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol={}&apikey={}",
                      symbol, api_key
    );

    println!("Fetching data from {}", url);

    let response = reqwest::get(&url)
        .await
        .map_err(|e| format!("HTTP error: {}", e))?;

    let raw_text = response.text()
        .await
        .map_err(|e| format!("HTTP error: {}", e))?;

    println!("Raw JSON Response: {}", raw_text);

    let alpha_data: AlphaVantageResponse = serde_json::from_str(&raw_text)
        .map_err(|e| format!("JSON parse error: {}", e))?;

    let price: f64 = alpha_data.global_quote.price.parse().unwrap_or(0.0);
    let change_str = alpha_data.global_quote.change_percent.trim_end_matches('%');
    let change_percent: f64 = change_str.parse().unwrap_or(0.0);

    let stock = Stock {
        symbol: alpha_data.global_quote.symbol,
        name: format!("{} Corp", symbol),
        price,
        change_percent,
    };

    Ok(stock)
}

async fn get_stock_data(symbol: String) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    match fetch_stock_data_internal(symbol).await {
        Ok(stock) => {
            println!("Sending response: {:?}", stock);
            Ok(Box::new(warp::reply::json(&stock)))
        }
        Err(e) => {
            println!("Error: {}", e);
            Err(warp::reject::not_found())
        }
    }
}

// Add another endpoint for multiple stock requests via HTTP
async fn get_multiple_stocks_http(symbols: String) -> Result<Box<dyn warp::Reply>, warp::Rejection> {
    let symbol_list: Vec<String> = symbols
        .split(',')
        .map(|s| s.trim().to_string())
        .filter(|s| !s.is_empty())
        .collect();

    if symbol_list.is_empty() {
        return Err(warp::reject::not_found());
    }

    let futures: Vec<_> = symbol_list.into_iter()
        .map(|symbol| async move {
            fetch_stock_data_internal(symbol).await
        })
        .collect();

    let results = join_all(futures).await;
    let stocks: Vec<Stock> = results.into_iter()
        .filter_map(|result| result.ok())
        .collect();

    Ok(Box::new(warp::reply::json(&stocks)))
}

#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();

    println!("Starting Market Data Service...");

    // Create CORS filter
    let cors = cors_filter();

    let stock_route = warp::path!("stock" / String)
        .and_then(get_stock_data);

    // New route for multiple requests
    let multiple_stocks_route = warp::path!("stocks" / String)
        .and_then(get_multiple_stocks_http);

    // Add OPTIONS handler for preflight requests
    let cors_preflight = warp::options()
        .map(|| -> Result<Box<dyn warp::Reply>, Infallible> {
            Ok(Box::new(warp::reply::with_status("", warp::http::StatusCode::OK)))
        })
        .and_then(|result| async { result });

    // Combine all routes
    let routes = stock_route
        .or(multiple_stocks_route)
        .or(cors_preflight)
        .with(cors);

    let grpc_service = MyMarketDataService::default();
    let grpc_server = MarketDataServiceServer::new(grpc_service);

    // Start both servers concurrently without tokio::spawn
    println!("HTTP server running on port 8002");
    println!("Available endpoints:");
    println!("  GET /stock/AAPL - Get single stock data");
    println!("  GET /stocks/AAPL,GOOGL,MSFT - Get multiple stocks data");
    println!("gRPC server running on port 8005");

    let http_server = warp::serve(routes)
        .run(([0, 0, 0, 0], 8002));

    let grpc_server_future = Server::builder()
        .add_service(grpc_server)
        .serve(([0, 0, 0, 0], 8005).into());

    // Run both servers concurrently
    tokio::select! {
        _ = http_server => println!("HTTP Server shutdown"),
        result = grpc_server_future => {
            if let Err(e) = result {
                eprintln!("gRPC server failed: {}", e);
            } else {
                println!("gRPC Server shutdown");
            }
        }
    }
}