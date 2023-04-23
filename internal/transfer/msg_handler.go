package transfer

import (
	"context"

	"github.com/Shopify/sarama"
	pbChat "github.com/geraldkohn/im/internal/api/chat"
	pbPush "github.com/geraldkohn/im/internal/api/push"
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/db"
	"github.com/geraldkohn/im/pkg/common/kafka"
	"github.com/geraldkohn/im/pkg/common/logger"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// MonogoDB 消费者组
type MsgConsumerHandler struct {
	topicHandle   map[string]handle
	consumerGroup *kafka.ConsumerGroup
}

func NewMsgConsumerHandler() *MsgConsumerHandler {
	h := &MsgConsumerHandler{
		topicHandle:   make(map[string]handle),
		consumerGroup: kafka.NewConsumerGroup([]string{constant.KafkaChatTopic}, constant.KafkaChatTopic),
	}
	h.topicHandle[constant.KafkaChatTopic] = msgHandle
	return h
}

// 只要发生任何错误, 都需要被重试
func msgHandle(msg []byte, msgKey string) error {
	msgFormat := &pbChat.MsgFormat{}
	err := proto.Unmarshal(msg, msgFormat)
	if err != nil {
		logger.Errorf("failed to unmarshal msg to phChat.MsgFormat, error: %v, msg: %v", err, msg)
		return err
	}
	switch msgFormat.ChatType {
	case constant.ChatSingle:
		strs := []string{msgFormat.SendID, msgFormat.RecvID}
		msgList := make([]*pbChat.MsgFormat, 0)
		for _, uid := range strs {
			seq, err := db.DB.IncrUserSeq(uid)
			if err != nil {
				logger.Errorf("failed to incr user sequence | error %v", err)
				return err
			}
			msg := &pbChat.MsgFormat{
				ChatType:    msgFormat.ChatType,
				SendID:      msgFormat.SendID,
				RecvID:      msgFormat.RecvID,
				GroupID:     msgFormat.GroupID,
				Sequence:    seq,
				SendTime:    msgFormat.SendTime,
				ContentType: msgFormat.ContentType,
				Content:     msgFormat.Content,
			}
			err = db.DB.SaveSingleChat(uid, msg)
			if err != nil {
				logger.Errorf("failed to save single message to monogodb | error %v", err)
				return err
			}
			// 只需要推送到接收者那里
			if uid == msgFormat.SendID {
				continue
			}
			msgList = append(msgList, msg)
		}
		go sendMsgToPush(msgList)
	case constant.ChatGroup:
		users, err := db.DB.GetGroupAllNumber(msgFormat.GroupID)
		if err != nil {
			logger.Errorf("failed to get group number from groupID | groupID %v | error %v", msgFormat.RecvID, err)
			return err
		}
		msgList := make([]*pbChat.MsgFormat, 0)
		for _, u := range users {
			uid := u.UserID
			seq, err := db.DB.IncrUserSeq(uid)
			if err != nil {
				logger.Errorf("failed to incr user sequence | error %v", err)
				return err
			}
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
			err = db.DB.SaveSingleChat(uid, msg)
			if err != nil {
				logger.Errorf("failed to save single message to monogodb | error %v", err)
				return err
			}
			// 只需要推送到接收者那里
			if uid == msgFormat.SendID {
				continue
			}
			msgList = append(msgList, msg)
		}
		go sendMsgToPush(msgList) // 发送给 Push 组件要启用新协程, 防止消费阻塞
	default:
		logger.Errorf("Unavailable Chat Type | ChatType %v", msgFormat.ChatType)
	}
	return nil
}

func (MsgConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (MsgConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *MsgConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Infof("MonogoConsumerHandler get a message %v", msg)
		fn := h.topicHandle[msg.Topic]
		err := fn(msg.Value, string(msg.Key))
		if err == nil {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}

// 发送到 Push 组件, 如果发送失败则写入 Kafka, 下次继续发送
func sendMsgToPush(message []*pbChat.MsgFormat) {
	var err error
	if err != nil {
		logger.Errorf("rpc send to push-element failed | error %v", err)
	}
	// 每次只发送一个, 发送成功后发送下一个, 不成功则写入 Kafka 等待消费.
	endpoint := pushDiscovery.PickOne()
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Failed to connect to push-grpc-server | address %v | error %v", endpoint, err)
		sendMsgToKafka(message)
		return
	}
	pushRPClient := pbPush.NewPushClient(conn)
	for _, m := range message {
		pushMsgReq := &pbPush.PushMsgReq{
			MsgFormat: &pbPush.MsgFormat{
				ChatType:    m.ChatType,
				SendID:      m.SendID,
				RecvID:      m.RecvID,
				GroupID:     m.GroupID,
				SendTime:    m.SendTime,
				Sequence:    m.Sequence,
				ContentType: m.ContentType,
				Content:     m.Content,
			},
		}
		pushMsgResp, err := pushRPClient.PushMsg(context.Background(), pushMsgReq)
		if err != nil {
			sendMsgToKafka([]*pbChat.MsgFormat{m})
			logger.Errorf("Failed to send to push client | error %v", err)
			continue
		}
		if pushMsgResp.ErrCode != 0 {
			logger.Errorf("Push-server failed to push message to gateway or user failed to receive | message %v | error %v", m, err)
			continue
		}
	}
}

// 不断尝试发送给 Kafka 要求重试消息, Kafka 如果出问题了那就不重试了, 等着用户拉取就行
func sendMsgToKafka(message []*pbChat.MsgFormat) {
	for _, m := range message {
		_, _, err := producer.SendMessage(m, m.RecvID)
		if err != nil {
			logger.Errorf("Failed to send retry message to kafka | message %v | error %v", message, err)
		}
	}
}
