use serde::{Deserialize, Serialize};
use warp::Filter;
use std::env;
use serde_json;
use tonic::{transport::Server, Request, Response, Status};
use futures::future::join_all;

pub mod market_proto {
    tonic::include_proto!("market");
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
// Include the generated protobuf code
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
                let response = GetStockPriceResponse
                {
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
async fn get_stock_data(symbol: String) -> Result<impl warp::Reply, warp::Rejection> {
    match fetch_stock_data_internal(symbol).await {
        Ok(stock) => {
            println!("Sending response: {:?}", stock);
            Ok(warp::reply::json(&stock))
        }
        Err(e) => {
            println!("Error: {}", e);
            Err(warp::reject::not_found())
        }
    }
}


#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();
    
    println!("🚀Starting Market Data Service...🚀");

    let stock_route = warp::path!("stock" / String)
        .and_then(get_stock_data);

    let grpc_service = MyMarketDataService::default();
    let grpc_server = MarketDataServiceServer::new(grpc_service);
    
    
    let http_handle = tokio::spawn(async move {
        println!("📡 HTTP server running on port 8002");
        warp::serve(stock_route)
            .run(([0, 0, 0, 0], 8002))
            .await;
    });
    
    let grpc_handle = tokio::spawn(async move {
        println!("⚡ gRPC server running on port 8005");
        Server::builder()
        .add_service(grpc_server)
        .serve(([0, 0, 0, 0], 8005).into())
        .await
        .expect("gRPC server failed");
    });
    
    tokio::select! {
    _ = http_handle => println!("HTTP Server shutdown"),
    _ = grpc_handle => println!("GRPC Server shutdown"),
    }
}