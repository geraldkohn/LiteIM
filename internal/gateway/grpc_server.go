package gateway

import (
	"net"
	"strconv"

	pbChat "Lite_IM/internal/api/rpc/chat"
	"Lite_IM/pkg/common/logger"

	"google.golang.org/grpc"
)

type GServer struct {
	pbChat.UnimplementedGatewayServer
	port int
}

func NewGrpcPushServer(port int) *GServer {
	s := &GServer{
		port: port,
	}
	return s
}

func (s *GServer) Run() {
	address := "127.0.0.1:" + strconv.Itoa(s.port)
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

	pbChat.RegisterGatewayServer(srv, s)
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
