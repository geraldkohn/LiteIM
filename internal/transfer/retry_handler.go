package transfer

import (
	pbChat "LiteIM/internal/api/rpc/chat"
	"LiteIM/pkg/common/logger"

	"google.golang.org/protobuf/proto"
)

// 重试, 将消息发送到 Push 组件
func (tf *Transfer) handleRetryPush(msg []byte, msgKey string) error {
	msgFormat := &pbChat.MsgFormat{}
	err := proto.Unmarshal(msg, msgFormat)
	if err != nil {
		logger.Errorf("failed to unmarshal msg to phChat.MsgFormat, error: %v, msg: %v", err, msg)
		return err
	}
	tf.sendSingleMsgToPush(msgFormat)
	return nil
}
