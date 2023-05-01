package gateway

import (
	"fmt"

	pbChat "github.com/geraldkohn/im/internal/api/rpc/chat"
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/logger"
	"github.com/geraldkohn/im/pkg/utils"
	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
)

func (ws *WServer) getUserMaxSeq(conn *UserConn, userID string, req *pbChat.GetUserMaxSeqRequest) {
	logger.Infof("ask to get user max sequence, userID: %s", userID)
	seq, err := database.GetUserMaxSeq(userID)
	if err != nil {
		glog.Errorf("failed to get user max sequence, error: %v", err)
		ws.writeMsg(conn, constant.ActionWSGetUserMaxSeq, constant.ErrRedis.ErrCode, constant.ErrRedis.ErrMsg, []byte{})
		return
	}
	resp := &pbChat.GetUserMaxSeqResponse{Seq: seq}
	b, _ := proto.Marshal(resp)
	ws.writeMsg(conn, constant.ActionWSGetUserMaxSeq, constant.OK.ErrCode, constant.OK.ErrMsg, b)
}

func (ws *WServer) pullMsgBySeqRange(conn *UserConn, userID string, req *pbChat.PullMsgBySeqRangeRequest) {
	logger.Infof("ask to pull msg by seq range, userID: %v, range: %v-%v", userID, req.SeqBegin, req.SeqEnd)
	result, err := database.GetMsgBySeqRange(userID, req.SeqBegin, req.SeqEnd)
	if err != nil {
		ws.writeMsg(conn, constant.ActionWSPullMsgBySeqRange, constant.ErrMongo.ErrCode, constant.ErrMongo.ErrMsg, []byte{})
		logger.Errorf("failed to get message by seq range from monogoDB, error: %v", err)
		return
	}
	list := &pbChat.MsgFormatList{}
	list.MsgFormats = append(list.MsgFormats, result...)
	b, _ := proto.Marshal(list)
	ws.writeMsg(conn, constant.ActionWSPullMsgBySeqRange, constant.OK.ErrCode, constant.OK.ErrMsg, b)
}

func (ws *WServer) pullMsgBySeqList(conn *UserConn, userID string, req *pbChat.PullMsgBySeqListRequest) {
	logger.Infof("ask to pull msg by seq list, userID: %v, reqList: %v", userID, req.SeqList)
	result, err := database.GetMsgBySeqList(userID, req.SeqList)
	if err != nil {
		ws.writeMsg(conn, constant.ActionWSPullMsgBySeqList, constant.ErrMongo.ErrCode, constant.ErrMongo.ErrMsg, []byte{})
		logger.Errorf("failed to get message by seq range from monogoDB, error: %v", err)
		return
	}
	list := &pbChat.MsgFormatList{}
	list.MsgFormats = append(list.MsgFormats, result...)
	b, _ := proto.Marshal(list)
	ws.writeMsg(conn, constant.ActionWSPullMsgBySeqList, constant.OK.ErrCode, constant.OK.ErrMsg, b)
}

// 客户端向服务端推送消息
func (ws *WServer) pushMsg(conn *UserConn, userID string, req *pbChat.PushMsgRequest) {
	logger.Infof("user pull message to gateway server, userID: %v", userID)
	if req.MsgFormat.SendTime == 0 {
		req.MsgFormat.SendTime = utils.GetCurrentTimestampByNano() // 设置消息发送时间
	}
	// 回调逻辑
	// TODO

	// ChatType
	switch req.MsgFormat.ChatType {
	case constant.ChatSingle:
		// 单聊的时候 key 设置为 sendID|recvID (sendID<recvID) 这样可以保证用户A与用户B单聊的信息被放入同一个分区, 这样就可以被一个协程处理, 而不是多个协程.
		key := fmt.Sprintf("%s|%s", req.MsgFormat.SendID, req.MsgFormat.RecvID)
		err := ws.sendMsgToKafka(req.MsgFormat, key)
		if err != nil {
			logger.Errorf("chatType: ChatSingle | failed to send msg to kafka | error: %v", err)
			ws.writeMsg(conn, constant.ActionWSPushMsgToServer, constant.ErrKafka.ErrCode, constant.ErrKafka.ErrMsg, []byte{})
			return
		}
		ws.writeMsg(conn, constant.ActionWSPushMsgToServer, constant.OK.ErrCode, constant.OK.ErrMsg, []byte{})
		return
	case constant.ChatGroup:
		// key 设置为 groupID, 这样可以保证一个群聊的消息被放入同一个分区.
		key := req.MsgFormat.GroupID
		err := ws.sendMsgToKafka(req.MsgFormat, key)
		if err != nil {
			logger.Errorf("failed to send msg to kafka, key: %v, msg: %v, error: %v", key, req.MsgFormat, err)
			ws.writeMsg(conn, constant.ActionWSPushMsgToServer, constant.ErrKafka.ErrCode, constant.ErrKafka.ErrMsg, []byte{})
			return
		}
		ws.writeMsg(conn, constant.ActionWSPushMsgToServer, constant.OK.ErrCode, constant.OK.ErrMsg, []byte{})
	default:
		ws.writeMsg(conn, constant.ActionWSPushMsgToServer, constant.ErrChatType.ErrCode, constant.ErrChatType.ErrMsg, []byte{})
	}
}

func (ws *WServer) sendMsgToKafka(m proto.Message, key string) error {
	partition, offset, err := ws.producer.SendMessage(m, key)
	if err != nil {
		logger.Infof("kafka send failed, key %v, send data %v, partition %v, offset %v, error %v", key, m, partition, offset, err)
		return err
	}
	return nil
}

func verifyToken(token string) (string, bool) {
	userID, err := utils.ParseToken(token)
	if err != nil {
		glog.Errorf("parse token error: %v", err)
		return "", false
	}
	if userID == "" {
		glog.Infof("unavailable token: %v", token)
		return "", false
	}
	glog.Infof("available token: %v", token)
	return userID, true
}

// // 获取
// func (ws *WServer) getUserMaxSeq(conn *UserConn, req *Req) {
// 	seq, err := ws.db.GetUserMaxSeq(req.UserID)
// 	if err != nil {
// 		glog.Errorf("无法获取用户最新序列号, error: %v", err)
// 		return
// 	}
// 	resp := &Resp{
// 		ReqIdentifier: req.ReqIdentifier,
// 		ErrCode:       constant.OK.ErrCode,
// 		ErrMsg:        constant.OK.ErrMsg,
// 	}
// 	d := RespDataMaxSeq{Sequence: seq}
// 	resp.Data, err = json.Marshal(d)
// 	if err != nil {
// 		glog.Errorf("json marshal error: %v", err)
// 	}
// 	obj, err := json.Marshal(resp)
// 	if err != nil {
// 		glog.Errorf("json marshal error: %v", err)
// 	}
// 	ws.sendTextMsg(conn, obj)
// }

// // 拉取离线消息直接拉 monogoDB 中的即可
// // 拉取消息返回的都是 protobuf
// // 根据首尾序列号拉取消息
// func (ws *WServer) pullMsgBySeqRange(conn *UserConn, req *Req) {
// 	seqRange := &ReqDataSeqPullRange{}
// 	err := json.Unmarshal(req.Data, seqRange)
// 	if err != nil {
// 		glog.Errorf("json unmarshal error: %v", err)
// 		return
// 	}
// 	// 从 mongoDB 中拉取消息
// 	msgList, _, _, err := ws.db.GetMsgBySeqRange(req.UserID, seqRange.Begin, seqRange.End)
// 	// 写入 websocket
// 	b, err := proto.Marshal(msgList)
// 	if err != nil {
// 		glog.Errorf("proto marshal error: %v", err)
// 	}
// 	ws.sendBinaryMsg(conn, b)
// }

// func (ws *WServer) pullMsgBySeqList(conn *UserConn, req *Req) {
// 	// 解析请求
// 	// 从 mongoDB 中拉取消息
// 	// 写入 websocket
// 	seqList := &ReqDataSeqPullList{}
// 	err := json.Unmarshal(req.Data, seqList)
// 	if err != nil {
// 		glog.Errorf("json unmarshal error: %v", err)
// 		return
// 	}
// 	// 从 mongoDB 中拉取消息
// 	msgList, _, _, err := ws.db.GetMsgBySeqList(req.UserID, seqList.List)
// 	// 写入 websocket
// 	b, err := proto.Marshal(msgList)
// 	if err != nil {
// 		glog.Errorf("proto marshal error: %v", err)
// 	}
// 	ws.sendBinaryMsg(conn, b)
// }

// // 客户端发送消息, 发送值为 json
// func (ws *WServer) pushMsg(conn *UserConn, req *Req) {
// 	// 解析为 protobuf
// 	// 发送到 Kafka 中, transfer 取出, 然后写序列号
// 	// 1. 写入 MonogoDB
// 	// 2. 写入 Push 组件
// 	// 1 2 同步进行
// 	// 2.1 Push 组件查看是否在线, 如果不在线则丢弃, 在线则写入对应的 gateway grpc 接口中.

// 	ws.producer.SendMessage()

// 	// 客户端不进行 ACK 确认. 通过指定拉取某个间隔序列号的信息, 客户端可以选择适时同步收到的 seq, 然后更新用户的 minSeq, 这样 monogo 中的消息就可以定时清理了, 或者是作为冷备份.
// }

// // 返回给客户端的接口应该使用这个, 现在就是实现一遍, 先不用
// func (ws *WServer) writeWSResp(conn *UserConn, resp *Resp) {

// }
