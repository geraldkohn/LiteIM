package push

import (
	"context"

	pbGateway "github.com/geraldkohn/im/internal/api/gateway"
	pbPush "github.com/geraldkohn/im/internal/api/push"
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/db"
	"github.com/geraldkohn/im/pkg/common/logger"
	"google.golang.org/grpc"
)

func (ps grpcPushServer) PushMsg(ctx context.Context, req *pbPush.PushMsgReq) (*pbPush.PushMsgResp, error) {
	// 1. 查找当前用户是否在线
	// 2. 如果不在线则丢弃; 如果在线, 则拿去到与用户生成 websocket 连接的 gateway
	// 3. 发送给对应的 gateway 数据, 等待返回消息
	// 4. 返回是否推送成功的消息
	uid := req.MsgFormat.RecvID
	endpoint, err := db.DB.GetOnlineUserGatewayEndpoint(uid)
	if err != nil {
		logger.Errorf("failed to write to redis | GetOnlineUserGatewayIP() | uid %v | error %v", uid, err)
		return &pbPush.PushMsgResp{UserOnline: false, GatewayEndpoint: "", ErrCode: constant.ErrRedis.ErrCode, ErrMsg: constant.ErrRedis.ErrMsg}, err
	}
	// 用户不在线, 当前的逻辑是直接返回不在线, 可以新增逻辑离线推送
	if endpoint == "" {
		return &pbPush.PushMsgResp{UserOnline: false, GatewayEndpoint: "", ErrCode: constant.OK.ErrCode, ErrMsg: constant.OK.ErrMsg}, nil
	}
	// 用户在线时
	// 调用 gateway grpc 接口发送消息
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Failed to connect to gateway-grpc-server | error %v", err)
		return &pbPush.PushMsgResp{UserOnline: true, GatewayEndpoint: endpoint, ErrCode: constant.ErrConnectionFailed.ErrCode, ErrMsg: constant.ErrConnectionFailed.ErrMsg}, err
	}
	gatewayRPClient := pbGateway.NewPushClient(conn)
	pushMsgReq := &pbGateway.PushMsgReq{MsgFormat: &pbGateway.MsgFormat{
		ChatType:    req.MsgFormat.ChatType,
		SendID:      req.MsgFormat.SendID,
		RecvID:      req.MsgFormat.RecvID,
		GroupID:     req.MsgFormat.GroupID,
		SendTime:    req.MsgFormat.SendTime,
		Sequence:    req.MsgFormat.Sequence,
		ContentType: req.MsgFormat.ContentType,
		Content:     req.MsgFormat.Content,
	}}
	pushMsgResp, err := gatewayRPClient.PushMsg(context.Background(), pushMsgReq)
	if err != nil {
		return &pbPush.PushMsgResp{UserOnline: true, GatewayEndpoint: endpoint, ErrCode: constant.ErrConnectionFailed.ErrCode, ErrMsg: constant.ErrConnectionFailed.ErrMsg}, err
	}
	return &pbPush.PushMsgResp{UserOnline: true, GatewayEndpoint: endpoint, ErrCode: pushMsgResp.ErrCode, ErrMsg: pushMsgResp.ErrMsg}, nil
}
