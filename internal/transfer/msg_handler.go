package transfer

import (
	"context"

	pbChat "Lite_IM/internal/api/rpc/chat"
	"Lite_IM/pkg/common/constant"
	"Lite_IM/pkg/common/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

// 只要发生任何错误, 都需要被重试
func (tf *Transfer) handleMsg(msg []byte, msgKey string) error {
	msgFormat := &pbChat.MsgFormat{}
	err := proto.Unmarshal(msg, msgFormat)
	if err != nil {
		logger.Errorf("failed to unmarshal msg to phChat.MsgFormat, error: %v, msg: %v", err, msg)
		return err
	}
	switch msgFormat.ChatType {
	case constant.ChatSingle:
		uids := []string{msgFormat.SendID, msgFormat.RecvID}
		msgToPushList := make([]*pbChat.MsgFormat, 0)
		msgToDBList := make([]*pbChat.MsgFormat, 0)
		for _, uid := range uids {
			// 获取最新序列号
			seq, err := tf.db.IncrUserSeq(uid)
			if err != nil {
				logger.Errorf("failed to incr user sequence | error %v", err)
				return err
			}
			// 复制
			msg := &pbChat.MsgFormat{
				ChatType:    msgFormat.ChatType,
				SendID:      msgFormat.SendID,
				RecvID:      uid, // 发送到发信人和收信人的收信箱
				GroupID:     msgFormat.GroupID,
				Sequence:    seq,
				SendTime:    msgFormat.SendTime,
				ContentType: msgFormat.ContentType,
				Content:     msgFormat.Content,
			}
			// 只需要推送到接收者那里
			if uid == msgFormat.SendID {
				msgToDBList = append(msgToDBList, msg)
			} else {
				msgToPushList = append(msgToPushList, msg)
				msgToDBList = append(msgToDBList, msg)
			}
		}
		err := tf.sendMsgToDB(msgToDBList)
		go tf.sendMultiMsgToPush(msgToPushList)
		if err != nil {
			logger.Errorf("Failed to sendMsgToDB | error %v", err)
			return err
		}
	case constant.ChatGroup:
		groupMemberList, err := tf.db.GetGroupMemberByGroupID(msgFormat.GroupID)
		if err != nil {
			logger.Errorf("failed to get group number from groupID | groupID %v | error %v", msgFormat.RecvID, err)
			return err
		}
		msgToPushList := make([]*pbChat.MsgFormat, 0)
		msgToDBList := make([]*pbChat.MsgFormat, 0)
		for _, gm := range groupMemberList {
			uid := gm.UserID
			// 获取最新序列号
			seq, err := tf.db.IncrUserSeq(uid)
			if err != nil {
				logger.Errorf("failed to incr user sequence | error %v", err)
				return err
			}
			// 复制
			msg := &pbChat.MsgFormat{
				ChatType:    msgFormat.ChatType,
				SendID:      msgFormat.SendID,
				RecvID:      uid, // 将收信箱设置为群组中对应的个人用户
				GroupID:     msgFormat.GroupID,
				Sequence:    seq,
				SendTime:    msgFormat.SendTime,
				ContentType: msgFormat.ContentType,
				Content:     msgFormat.Content,
			}
			// 只需要推送到接收者那里
			if uid == msgFormat.SendID {
				continue
			}
			msgToPushList = append(msgToPushList, msg)
		}
		err = tf.sendMsgToDB(msgToDBList)
		if err != nil {
			logger.Errorf("Failed to sendMsgToDB | error %v", err)
			return err
		}
		go tf.sendMultiMsgToPush(msgToPushList) // 发送给 Push 组件要启用新协程, 防止消费阻塞
	default:
		logger.Errorf("Unavailable Chat Type | ChatType %v", msgFormat.ChatType)
	}
	return nil
}

// 发送到 Push 组件, 如果发送失败则写入 Kafka, 下次继续发送
func (tf *Transfer) sendMultiMsgToPush(message []*pbChat.MsgFormat) {
	var err error
	if err != nil {
		logger.Errorf("rpc send to push-element failed | error %v", err)
	}
	// 每次只发送一个, 发送成功后发送下一个, 不成功则写入 Kafka 等待消费.
	endpoint := tf.pushDiscovery.PickOne()
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("Failed to connect to push-grpc-server | address %v | error %v", endpoint, err)
		tf.retryMultiPush(message)
		return
	}
	pushRPClient := pbChat.NewPusherClient(conn)
	for _, m := range message {
		pushMsgReq := &pbChat.PushMsgToPusherRequest{
			MsgFormat: m,
		}
		pushMsgResp, err := pushRPClient.PushMsgToPusher(context.Background(), pushMsgReq)
		if err != nil {
			tf.retrySinglePush(m)
			logger.Errorf("Failed to send to push client | error %v", err)
			continue
		}
		if pushMsgResp.ErrCode != 0 {
			logger.Errorf("Push-server failed to push message to gateway or user failed to receive | message %v | error %v", m, err)
			continue
		}
	}
}

// 发送到 Push 组件, 如果发送失败则写入 Kafka, 下次继续发送
func (tf *Transfer) sendSingleMsgToPush(message *pbChat.MsgFormat) {
	var err error
	endpoint := tf.pushDiscovery.PickOne()
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("Failed to connect to push-grpc-server | address %v | error %v", endpoint, err)
		tf.retrySinglePush(message)
		return
	}
	pushRPClient := pbChat.NewPusherClient(conn)
	pushMsgReq := &pbChat.PushMsgToPusherRequest{
		MsgFormat: message,
	}
	pushMsgResp, err := pushRPClient.PushMsgToPusher(context.Background(), pushMsgReq)
	if err != nil {
		tf.retrySinglePush(message)
		logger.Errorf("Failed to send to push client | error %v", err)
	}
	if pushMsgResp.ErrCode != 0 {
		logger.Errorf("Push-server failed to push message to gateway or user failed to receive | message %v | error %v", message, err)
	}
}

// 将消息持久化到 MongoDB
func (tf *Transfer) sendMsgToDB(msgList []*pbChat.MsgFormat) error {
	uidList := make([]*string, len(msgList))
	for i := 0; i < len(msgList); i++ {
		uidList[i] = &msgList[i].RecvID
	}
	return tf.db.SaveMultiChat(uidList, msgList)
}

// 不断尝试发送给 Kafka 要求重新推送到 Push, Kafka 如果出问题了那就不重试了, 等着用户拉取就行
func (tf *Transfer) retryMultiPush(message []*pbChat.MsgFormat) {
	for _, m := range message {
		tf.retrySinglePush(m)
	}
}

// 不断尝试发送给 Kafka 要求重新推送到 Push, Kafka 如果出问题了那就不重试了, 等着用户拉取就行
func (tf *Transfer) retrySinglePush(message *pbChat.MsgFormat) {
	_, _, err := tf.retryPushProducer.SendMessage(message, message.RecvID)
	if err != nil {
		logger.Errorf("Failed to send retry message to kafka | message %v | error %v", message, err)
	}
}
