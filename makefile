build-proto:
	protoc --go_out=internal/api --go-grpc_out=internal/api internal/api/chat/chat.proto
	protoc --go_out=internal/api --go-grpc_out=internal/api internal/api/push/push.proto
	protoc --go_out=internal/api --go-grpc_out=internal/api internal/api/gateway/gateway.proto
	
	