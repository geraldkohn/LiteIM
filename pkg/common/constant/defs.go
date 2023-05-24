package constant

const (
	// Action Type
	ActionWSGetUserMaxSeq     int32 = 1001
	ActionWSPullMsgBySeqRange int32 = 1002
	ActionWSPullMsgBySeqList  int32 = 1003
	ActionWSPushMsgToServer   int32 = 1004
	ActionWSPushMsgToClient   int32 = 1005

	// Content Type
	ContentText    int32 = 101
	ContentPicture int32 = 102
	ContentVideo   int32 = 103
	ContentFile    int32 = 104

	// Chat Type
	ChatSingle int32 = 11
	ChatGroup  int32 = 12
)
