build-proto:
	cd internal/api/rpc && protoc --go_out=. --go-grpc_out=. chat/chat.proto
	
build-images: # 默认版本为latest
	docker build -t geraldkohn/im-gateway -f dockerfiles/gateway.Dockerfile .
	docker build -t geraldkohn/im-pusher -f dockerfiles/pusher.Dockerfile .
	docker build -t geraldkohn/im-transfer -f dockerfiles/transfer.Dockerfile .

push-images: # 默认版本为latest
	docker push geraldkohn/im-gateway
	docker push geraldkohn/im-pusher
	docker push geraldkohn/im-transfer

deploy-components:
	kubectl apply -f deploy/components/etcd
	kubectl apply -f deploy/components/kafka
	kubectl apply -f deploy/components/mongodb
	kubectl apply -f deploy/components/mysql
	kubectl apply -f deploy/components/redis

deploy-source:
	kubectl apply -f deploy/gateway
	kubectl apply -f deploy/pusher
	kubectl apply -f deploy/transfer
