package pusher

import (
	"context"

	pbChat "LiteIM/internal/api/rpc/chat"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (ps *grpcServer) PushMsgToPusher(ctx context.Context, req *pbChat.PushMsgToPusherRequest) (*pbChat.PushMsgToPusherResponse, error) {
	// 1. 查找当前用户是否在线
	// 2. 如果不在线则丢弃; 如果在线, 则拿去到与用户生成 websocket 连接的 gateway
	// 3. 发送给对应的 gateway 数据, 等待返回消息
	// 4. 返回是否推送成功的消息
	uid := req.MsgFormat.RecvID
	endpoint, err := p.db.GetOnlineUserGatewayEndpoint(uid)
	if err != nil {
		logger.Errorf("failed to write to redis | GetOnlineUserGatewayIP() | uid %v | error %v", uid, err)
		return &pbChat.PushMsgToPusherResponse{UserOnline: false, GatewayEndpoint: "", ErrCode: constant.ErrRedis.ErrCode, ErrMsg: constant.ErrRedis.ErrMsg}, err
	}
	// 用户不在线, 当前的逻辑是直接返回不在线, 可以新增逻辑离线推送
	if endpoint == "" {
		return &pbChat.PushMsgToPusherResponse{UserOnline: false, GatewayEndpoint: "", ErrCode: constant.OK.ErrCode, ErrMsg: constant.OK.ErrMsg}, nil
	}
	// 用户在线时
	// 调用 gateway grpc 接口发送消息
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("Failed to connect to gateway-grpc-server | error %v", err)
		return &pbChat.PushMsgToPusherResponse{UserOnline: true, GatewayEndpoint: endpoint, ErrCode: constant.ErrConnectionFailed.ErrCode, ErrMsg: constant.ErrConnectionFailed.ErrMsg}, err
	}
	gatewayRPClient := pbChat.NewGatewayClient(conn)
	pushMsgReq := &pbChat.PushMsgToGatewayRequest{MsgFormat: req.MsgFormat}
	pushMsgResp, err := gatewayRPClient.PushMsgToGateway(context.Background(), pushMsgReq)
	if err != nil {
		return &pbChat.PushMsgToPusherResponse{UserOnline: true, GatewayEndpoint: endpoint, ErrCode: constant.ErrConnectionFailed.ErrCode, ErrMsg: constant.ErrConnectionFailed.ErrMsg}, err
	}
	return &pbChat.PushMsgToPusherResponse{UserOnline: true, GatewayEndpoint: endpoint, ErrCode: pushMsgResp.ErrCode, ErrMsg: pushMsgResp.ErrMsg}, nil
}
