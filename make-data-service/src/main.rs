use serde::{Deserialize, Serialize};
use warp::Filter;
use std::env;
use serde_json;

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

async fn get_stock_data(symbol: String) -> Result<impl warp::Reply, warp::Rejection> {
    let api_key = env::var("ALPHA_API").unwrap_or_else(|_| "demo".to_string());
    let url = format!("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol={}&apikey={}", 
    symbol, api_key
    );

    println!("Fetching data from: {}", url);
    let response = reqwest::get(&url)
        .await
        .map_err(|e| {
            println!("HTTP error: {}", e);
            warp::reject::not_found()
        })?;
    
    let raw_text = response.text()
        .await
        .map_err(|e| {
            println!("Text error: {}", e);
            warp::reject::not_found()
        })?;
    
    println!("Raw JSON response:)");
    println!("{}", raw_text);

    let alpha_data: AlphaVantageResponse = serde_json::from_str(&raw_text)
        .map_err(|e| {
            println!("JSON error: {}", e);
            warp::reject::not_found()
        })?;

    println!("Received data:");
    println!("  Symbol: {}", alpha_data.global_quote.symbol);
    println!("  Price: {}", alpha_data.global_quote.price);
    println!("  Change: {}", alpha_data.global_quote.change_percent);

    let price: f64 = alpha_data.global_quote.price.parse().unwrap_or(0.0);
    let change_str = alpha_data.global_quote.change_percent.trim_end_matches('%');
    let change_percent: f64 = change_str.parse().unwrap_or(0.0);
    
    let stock = Stock {
        symbol: alpha_data.global_quote.symbol,
        name: format!("{} Corp", symbol),
        price,
        change_percent,
    };
    
    println!("Sending response: {:?}", stock);
    Ok(warp::reply::json(&stock))
}

#[tokio::main]
async fn main() {
    let stock_route = warp::path!("stock" / String)
        .and_then(get_stock_data);
    
    println!("Market Data Service is running on port 8002...");
    warp::serve(stock_route)
        .run(([0, 0, 0, 0], 8002))
        .await;
}