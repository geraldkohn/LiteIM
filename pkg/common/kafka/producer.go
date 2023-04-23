package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/geraldkohn/im/pkg/common/logger"
	"github.com/geraldkohn/im/pkg/common/setting"
	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
)

type Producer struct {
	Topic    string
	Producer sarama.SyncProducer
}

func NewKafkaProducer(topic string) *Producer {
	p := Producer{}
	p.Topic = topic
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Return.Errors = true
	if setting.APPSetting.Kafka.SASLUserName != "" && setting.APPSetting.Kafka.SASLPassword != "" {
		producerConfig.Net.SASL.Enable = true
		producerConfig.Net.SASL.User = setting.APPSetting.Kafka.SASLUserName
		producerConfig.Net.SASL.Password = setting.APPSetting.Kafka.SASLPassword
	}
	producer, err := sarama.NewSyncProducer(setting.APPSetting.Kafka.BrokerAddr, producerConfig)
	if err != nil {
		panic(err.Error())
		return nil
	}
	p.Producer = producer
	return &p
}

// 同步发送消息
func (p *Producer) SendMessage(m proto.Message, key string) (partition int32, offset int64, err error) {
	logger.Infof("Send Message To Kafka, message %v, key %v", m, key)
	kMsg := &sarama.ProducerMessage{}
	bMsg, err := proto.Marshal(m)
	if err != nil {
		glog.Errorf("Producer-SendMessage() error: %s", err)
		return -1, -1, err
	}
	kMsg.Topic = p.Topic
	kMsg.Key = sarama.StringEncoder(key)
	kMsg.Value = sarama.ByteEncoder(bMsg)
	return p.Producer.SendMessage(kMsg)
}