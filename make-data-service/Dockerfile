FROM rust:1.87

WORKDIR /app
COPY Cargo.toml ./
COPY src/ ./src/
RUN cargo build --release
EXPOSE 8002
CMD ["./target/release/make-data-service"]