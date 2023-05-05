package transfer

import (
	pbChat "Lite_IM/internal/api/rpc/chat"
	"Lite_IM/pkg/common/logger"

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
	tf.sendMsgToPush([]*pbChat.MsgFormat{msgFormat})
	return nil
}

func (tf *Transfer) handleRetryDB(msg []byte, msgKey string) error {
	msgFormat := &pbChat.MsgFormat{}
	err := proto.Unmarshal(msg, msgFormat)
	if err != nil {
		logger.Errorf("failed to unmarshal msg to phChat.MsgFormat, error: %v, msg: %v", err, msg)
		return err
	}
	tf.sendMsgToDB([]*pbChat.MsgFormat{msgFormat})
	return nil
}
