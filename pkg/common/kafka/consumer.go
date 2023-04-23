package kafka

import (
	"sync"

	"github.com/Shopify/sarama"
	"github.com/geraldkohn/im/pkg/common/setting"
)

type Consumer struct {
	WG            sync.WaitGroup
	Topic         string
	PartitionList []int32
	Consumer      sarama.Consumer
}

func NewKafkaConsumer(topic string) *Consumer {
	consumerConfig := sarama.NewConfig()
	if setting.APPSetting.Kafka.SASLUserName != "" && setting.APPSetting.Kafka.SASLPassword != "" {
		consumerConfig.Net.SASL.Enable = true
		consumerConfig.Net.SASL.User = setting.APPSetting.Kafka.SASLUserName
		consumerConfig.Net.SASL.Password = setting.APPSetting.Kafka.SASLPassword
	}

	p := Consumer{}
	p.Topic = topic
	consumer, err := sarama.NewConsumer(setting.APPSetting.Kafka.BrokerAddr, consumerConfig)
	if err != nil {
		panic(err.Error())
		return nil
	}
	p.Consumer = consumer

	partitionList, err := consumer.Partitions(p.Topic)
	if err != nil {
		panic(err.Error())
		return nil
	}
	p.PartitionList = partitionList

	return &p
}
