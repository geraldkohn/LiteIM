package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
)

type ConsumerGroup struct {
	sarama.ConsumerGroup
	groupID string
	topics  []string
}

func NewConsumerGroup(topics []string, groupID string) *ConsumerGroup {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 未找到组消费位移的时候从哪边开始消费
	consumerGroup, err := sarama.NewConsumerGroup(topics, groupID, config)
	if err != nil {
		glog.Errorf("NewConsumerGroup() failed to create consumer group, error: %s", err)
	}
	return &ConsumerGroup{
		consumerGroup,
		groupID,
		topics,
	}
}

func (cg *ConsumerGroup) RegisterHandleAndConsumer(handler sarama.ConsumerGroupHandler) {
	ctx := context.Background()
	for {
		err := cg.ConsumerGroup.Consume(ctx, cg.topics, handler)
		if err != nil {
			glog.Errorf("ConsumerGroup-RegisterHandleAndConsumer() return a unexpected error: %s", err)
		}
	}
}

// // 消费者组处理结构体
// type consumerGroupHandler struct {
// }

// func (consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
// 	return nil
// }
// func (consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// // ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// func (consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
// 	for msg := range claim.Messages() { // 读消息 channel
// 		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", msg.Value, msg.Timestamp, msg.Topic)
// 		session.MarkMessage(msg, "") // 将 Kafka 消息标记为已消费
// 	}
// 	return nil
// }
