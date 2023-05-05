package gateway

import (
	pbChat "Lite_IM/internal/api/rpc/chat"
	"Lite_IM/pkg/common/constant"
	"Lite_IM/pkg/common/logger"
	"context"
)

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
