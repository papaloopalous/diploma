module logger

go 1.24.1

require (
	diploma/logservice v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.71.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace diploma/logservice => ../logservice
