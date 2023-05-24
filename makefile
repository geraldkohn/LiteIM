build-proto:
	protoc --go_out=internal/api --go-grpc_out=internal/api internal/api/chat/chat.proto
	
build-gateway-image:
	docker build -t geraldkohn/im-gateway -f dockerfiles/gateway.Dockerfile .