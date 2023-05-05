package db

import (
	"context"
	"fmt"

	pbChat "Lite_IM/internal/api/rpc/chat"

	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

const (
	collectionChat  = "chat"
	collectionGroup = "group"
)

type UserChat struct {
	UID string `bson:"uid"`
	Seq int64  `bson:"seq"`
	Msg []byte `bson:"msg"`
}

// 根据序列号范围来拉取消息
func (d *DataBases) GetMsgBySeqRange(uid string, seqBegin, seqEnd int64) ([]*pbChat.MsgFormat, error) {
	filter := bson.M{"uid": uid, "seq": bson.M{"$gt": seqBegin, "$lt": seqEnd}}
	return d.getMsgByFilter(filter)
}

// 根据序列号列表来拉取消息
func (d *DataBases) GetMsgBySeqList(uid string, seqList []int64) ([]*pbChat.MsgFormat, error) {
	filter := bson.M{"uid": uid, "seq": bson.M{"$in": seqList}}
	return d.getMsgByFilter(filter)
}

func (d *DataBases) getMsgByFilter(filter bson.M) ([]*pbChat.MsgFormat, error) {
	var err error
	collection := d.mongo.Database(mongodbDataBase).Collection(collectionChat)
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())
	// 遍历结果
	var userChats []UserChat
	var msgFormats []*pbChat.MsgFormat
	if err = cursor.All(context.Background(), &userChats); err != nil {
		return nil, err
	}
	// msgFormat 在 UserChat 中的 Msg 字段中
	for _, uc := range userChats {
		var mf pbChat.MsgFormat
		err = proto.Unmarshal(uc.Msg, &mf)
		if err != nil {
			return nil, err
		}
		msgFormats = append(msgFormats, &mf)
	}
	return msgFormats, nil
}

// 直接存储单个用户消息
func (d *DataBases) SaveSingleChat(uid string, msg *pbChat.MsgFormat) error {
	var err error
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	collection := d.mongo.Database(mongodbDataBase).Collection(collectionChat)
	userChat := UserChat{
		UID: uid,
		Seq: msg.Sequence,
		Msg: b,
	}
	_, err = collection.InsertOne(context.TODO(), userChat)
	if err != nil {
		return err
	}
	return nil
}

// 直接存储多个用户消息
func (d *DataBases) SaveMultiChat(uidList []*string, msgList []*pbChat.MsgFormat) error {
	var err error
	if len(uidList) != len(msgList) {
		return fmt.Errorf("length of uidList is not as same as length of msgList")
	}
	length := len(uidList)
	userChatList := make([]interface{}, length)
	for i := 0; i < length; i++ {
		b, err := proto.Marshal(msgList[i])
		if err != nil {
			return err
		}
		userChatList[i] = UserChat{
			UID: *uidList[i],
			Seq: msgList[i].Sequence,
			Msg: b,
		}
	}
	collection := d.mongo.Database(mongodbDataBase).Collection(collectionChat)
	_, err = collection.InsertMany(context.TODO(), userChatList, nil)
	if err != nil {
		return err
	}
	return nil
}
