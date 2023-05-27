package kafka

import (
	"LiteIM/pkg/common/config"
	"LiteIM/pkg/common/logger"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
)

type Producer struct {
	Topic    string
	Producer sarama.SyncProducer
}

type KafkaProducerConfig struct {
	BrokerAddr   []string
	SASLUsername string
	SASLPassword string
	Topic        string
}

func NewKafkaProducer(kc KafkaProducerConfig) *Producer {
	p := Producer{}
	p.Topic = kc.Topic
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Return.Errors = true
	if config.Conf.Kafka.KafkaSASLUsername != "" && config.Conf.Kafka.KafkaSASLPassword != "" {
		producerConfig.Net.SASL.Enable = true
		producerConfig.Net.SASL.User = kc.SASLUsername
		producerConfig.Net.SASL.Password = kc.SASLPassword
	}
	producer, err := sarama.NewSyncProducer(kc.BrokerAddr, producerConfig)
	if err != nil {
		panic(err.Error())
		return nil
	}
	p.Producer = producer
	return &p
}

// 同步发送消息
func (p *Producer) SendMessage(m proto.Message, key string) (partition int32, offset int64, err error) {
	logger.Logger.Infof("Send Message To Kafka, message %v, key %v", m, key)
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
