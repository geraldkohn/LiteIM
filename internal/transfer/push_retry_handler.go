package transfer

import (
	"github.com/Shopify/sarama"
	pbChat "github.com/geraldkohn/im/internal/api/chat"
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/kafka"
	"github.com/geraldkohn/im/pkg/common/logger"
	"google.golang.org/protobuf/proto"
)

// Push Message 消费者
type PushConsumerHandler struct {
	topicHandle   map[string]handle
	consumerGroup *kafka.ConsumerGroup
}

func NewPushConsumerGroupHandler() *PushConsumerHandler {
	h := &PushConsumerHandler{
		topicHandle:   make(map[string]handle),
		consumerGroup: kafka.NewConsumerGroup([]string{constant.KafkaChatTopic}, constant.KafkaPushConsumerGroupID),
	}
	h.topicHandle[constant.KafkaChatTopic] = pushHandle
	return h
}

// 重试, 将消息发送到 Push 组件
func pushHandle(msg []byte, msgKey string) error {
	msgFormat := &pbChat.MsgFormat{}
	err := proto.Unmarshal(msg, msgFormat)
	if err != nil {
		logger.Errorf("failed to unmarshal msg to phChat.MsgFormat, error: %v, msg: %v", err, msg)
		return err
	}
	sendMsgToPush([]*pbChat.MsgFormat{msgFormat})
	return nil
}

func (PushConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (PushConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *PushConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Infof("PushConsumerHandler get a message %v", msg)
		fn := h.topicHandle[msg.Topic]
		err := fn(msg.Value, string(msg.Key))
		if err == nil {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}
