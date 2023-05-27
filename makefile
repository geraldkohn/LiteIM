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

deploy-k8s-components:
	kubectl apply -f deploy_k8s/components/etcd
	kubectl apply -f deploy_k8s/components/kafka
	kubectl apply -f deploy_k8s/components/mongodb
	kubectl apply -f deploy_k8s/components/mysql
	kubectl apply -f deploy_k8s/components/redis

clean-k8s-components:
	kubectl delete sts --all -n geraldkohn
	kubectl delete pvc --all -n geraldkohn
	kubectl delete pv  --all -n geraldkohn

deploy-k8s-all:
	kubectl apply -f deploy_k8s/gateway
	kubectl apply -f deploy_k8s/pusher
	kubectl apply -f deploy_k8s/transfer

deploy-docker-components:
	

test-it:
	docker build -t geraldkohn/im-gateway -f dockerfiles/gateway.Dockerfile .
	docker push geraldkohn/im-gateway
	kubectl apply -f deploy_k8s/gateway
