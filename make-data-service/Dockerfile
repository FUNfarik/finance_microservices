FROM rust:1.87

# Install protobuf compiler
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the service-specific files
COPY make-data-service/Cargo.toml ./make-data-service/
COPY make-data-service/build.rs ./make-data-service/
COPY make-data-service/src/ ./make-data-service/src/

# Copy the proto files to the expected location relative to make-data-service/
COPY proto/ ./proto/

# Set working directory to the service directory
WORKDIR /app/make-data-service

# Build the application
RUN cargo build --release

EXPOSE 8002

CMD ["./target/release/make-data-service"]