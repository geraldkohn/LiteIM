package gateway

import (
	"net"
	"strconv"

	pbGateway "github.com/geraldkohn/im/internal/api/gateway"
	"github.com/geraldkohn/im/pkg/common/logger"
	"google.golang.org/grpc"
)

type grpcPushServer struct {
	pbGateway.UnimplementedPushServer
	Port int
}

func NewGrpcPushServer(port int) *grpcPushServer {
	ps := &grpcPushServer{
		Port: port,
	}
	return ps
}

func (ps *grpcPushServer) Run() {
	address := "127.0.0.1:" + strconv.Itoa(ps.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorf("listen network failed, err = %s, address = %s", err.Error(), address)
		return
	}
	logger.Infof("listen network success, address = %s", address)

	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd

	pbGateway.RegisterPushServer(srv, ps)
	if err != nil {
		logger.Errorf("register rpc get_token to etcd failed, err = %s", err.Error())
		return
	}

	err = srv.Serve(listener)
	if err != nil {
		logger.Infof("rpc get_token fail, err = %s", err.Error())
		return
	}
	logger.Infof("rpc get_token init success")
}
