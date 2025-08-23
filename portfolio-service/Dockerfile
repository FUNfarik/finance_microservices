FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy proto files first (now accessible from root context)
COPY proto/ ./proto/

# Copy portfolio service files
COPY portfolio-service/go.mod portfolio-service/go.sum ./
RUN go mod download

COPY portfolio-service/ .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8003
CMD ["./main"]