package db

import (
	pbChat "github.com/geraldkohn/im/internal/api/chat"
)

const (
	cChat  = "chat"
	cGroup = "group"
)

// TODO

// 根据序列号范围来拉取消息
func (d *DataBases) GetMsgBySeqRange(uid string, seqBegin, seqEnd int64) ([]*pbChat.MsgFormat, error) {
	return nil, nil
}

// 根据序列号列表来拉取消息
func (d *DataBases) GetMsgBySeqList(uid string, seqList []int64) ([]*pbChat.MsgFormat, error) {
	return nil, nil
}

// 直接存储单个用户消息
func (d *DataBases) SaveSingleChat(uid string, msg *pbChat.MsgFormat) error {
	return nil
}

// 将消息存储到多个用户的信箱
func (d *DataBases) SaveGroupChat(uid []string, msg *pbChat.MsgFormat) error {
	return nil
}
