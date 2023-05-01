package gateway

import (
	"context"
	"net"
	"strconv"

	pbChat "github.com/geraldkohn/im/internal/api/rpc/chat"
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/logger"
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

// 将 Pusher 发来的消息发送到客户端
func (s *GServer) PushMsgToGateway(ctx context.Context, req *pbChat.PushMsgToGatewayRequest) (*pbChat.PushMsgToGatewayResponse, error) {
	msg := req.GetMsgFormat()
	logger.Infof("PushMsg is arriving, recvID: %v", msg.GetRecvID())
	conn := wServer.getUserConn(msg.GetRecvID())
	err := wServer.writeMsg(conn, constant.ActionWSPushMsgToClient, constant.OK.ErrCode, constant.OK.ErrMsg, []byte{})
	if err != nil {
		return &pbChat.PushMsgToGatewayResponse{
			Online:  false,
			ErrCode: constant.ErrConnectionFailed.ErrCode,
			ErrMsg:  constant.ErrConnectionFailed.ErrMsg,
		}, nil
	} else {
		return &pbChat.PushMsgToGatewayResponse{
			Online:  true,
			ErrCode: constant.OK.ErrCode,
			ErrMsg:  constant.OK.ErrMsg,
		}, nil
	}
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
