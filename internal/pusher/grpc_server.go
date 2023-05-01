package pusher

import (
	"net"
	"strconv"

	pbChat "github.com/geraldkohn/im/internal/api/rpc/chat"
	"github.com/geraldkohn/im/pkg/common/logger"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pbChat.UnimplementedPusherServer
	Port int
}

func newGrpcServer(port int) *grpcServer {
	ps := &grpcServer{
		Port: port,
	}
	return ps
}

func (s *grpcServer) Run() {
	address := "127.0.0.1:" + strconv.Itoa(s.Port)
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

	pbChat.RegisterPusherServer(srv, s)
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

// TODO
func (s *grpcServer) Exit() {
}
