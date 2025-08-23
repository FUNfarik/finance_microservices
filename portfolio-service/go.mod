module portfolio-service

go 1.24.5

require (
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.75.0
)

require google.golang.org/protobuf v1.36.8 // indirect

require (
	github.com/FUNfarik/finance_microservices/proto/go v0.0.0
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
)

replace github.com/FUNfarik/finance_microservices/proto/go => ./proto/go
