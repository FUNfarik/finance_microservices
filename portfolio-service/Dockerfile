FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY proto/ ./proto/
COPY portfolio-service/ ./portfolio-service/

WORKDIR /app/portfolio-service

RUN go mod download

RUN go get github.com/FUNfarik/finance_microservices/proto/go
RUN go mod download

RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/portfolio-service/main .
CMD ["./main"]