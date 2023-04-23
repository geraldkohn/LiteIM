package constant

const (
	// Service Name
	GatewayServiceName  string = "gateway"
	TransferServiceName string = "transfer"
	PushServiceName     string = "push"

	// Action Type
	ActionWSGetUserMaxSeq     int32 = 1001
	ActionWSPullMsgBySeqRange int32 = 1002
	ActionWSPullMsgBySeqList  int32 = 1003
	ActionWSPushMsg           int32 = 1004

	// Content Type
	ContentText    int32 = 101
	ContentPicture int32 = 102
	ContentVideo   int32 = 103
	ContentFile    int32 = 104

	// Chat Type
	ChatSingle int32 = 11
	ChatGroup  int32 = 12

	KafkaChatTopic string = "chat"
	// KafkaChatMonogoRetryTopic  string = "chat_monogo"
	KafkaChatPushRetryTopic string = "chat_push"
	// KafkaMonogoConsumerGroupID string = "monogo_consumer_group"
	KafkaPushConsumerGroupID string = "push_consumer_group"
)
