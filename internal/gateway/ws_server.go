package gateway

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	pbChat "github.com/geraldkohn/im/internal/api/chat"
	"github.com/geraldkohn/im/pkg/common/constant"
	"github.com/geraldkohn/im/pkg/common/db"
	"github.com/geraldkohn/im/pkg/common/kafka"
	"github.com/geraldkohn/im/pkg/common/setting"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type UserConn struct {
	*websocket.Conn
	lock   *sync.Mutex // 保护 websocket 连接, 顺序写入
	userID string      // 用户 ID
	drop   int32       // 连接是否已关闭, drop: 0 未关闭; drop: 1 已关闭
}

type WServer struct {
	wsUpgrader *websocket.Upgrader  // websocket upgrader
	lock       *sync.RWMutex        // 保护 map
	wsUserConn map[string]*UserConn // 用户 ID -> 连接
	producer   *kafka.Producer      // Kafka Producer
}

func (ws *WServer) onInit() {
	ws.wsUpgrader = &websocket.Upgrader{
		HandshakeTimeout: time.Duration(setting.APPSetting.Websocket.Timeout) * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
	ws.wsUserConn = make(map[string]*UserConn)
	ws.lock = new(sync.RWMutex)
	ws.producer = kafka.NewKafkaProducer(constant.KafkaChatTopic)
}

func (ws *WServer) run() {
	http.HandleFunc("/", ws.wsHandler)
	err := http.ListenAndServe(":"+strconv.Itoa(setting.APPSetting.Websocket.Port), nil)
	if err != nil {
		panic("Websocket Server Listening error:" + err.Error())
	}
}

func (ws *WServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	conn, err := ws.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("update http connection failed %s", err.Error())
		return
	} else {
		newConn := &UserConn{conn, new(sync.Mutex), query["userID"][0], 0}
		go ws.readMsg(newConn)
	}
}

func (ws *WServer) addUserConn(conn *UserConn) error {
	// 更新 用户--Gateway 映射
	endpoint := fmt.Sprintf("%s:%d", IP, setting.APPSetting.RPC.GatewayRPCPort)
	_, err := db.DB.SetOnlineUserGatewayEndpoint(conn.userID, endpoint)
	if err != nil {
		return err
	}
	ws.lock.Lock()
	ws.wsUserConn[conn.userID] = conn
	ws.lock.Unlock()
	return nil
}

func (ws *WServer) delUserConn(conn *UserConn) error {
	atomic.StoreInt32(&conn.drop, 1)
	// 更新 用户--Gateway 映射
	_, err := db.DB.DeleteOnlineUserGatewayEndpoint(conn.userID)
	if err != nil {
		return err
	}
	ws.lock.Lock()
	delete(ws.wsUserConn, conn.userID)
	ws.lock.Unlock()
	return nil
}

func (ws *WServer) readMsg(conn *UserConn) {
	defer glog.Infof("WServer readMsg return, conn: %v", conn)
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			glog.Errorf("read message error: %v", err)
			ws.delUserConn(conn)
			return
		}
		if messageType == websocket.CloseMessage {
			glog.Infof("websocket receive close message %s", conn.userID)
			ws.delUserConn(conn)
			return
		}
		if atomic.LoadInt32(&conn.drop) == 1 {
			glog.Infof("websocket connection has been closed %s", conn.userID)
			return
		}
		if messageType == websocket.PongMessage {
			// 收到客户端的 Pong Message, 需要更新客户端到期时间
			// 这里可以仿照 nsq 来做
			// TODO
			continue
		}
		ws.parseMsg(conn, msg)
	}
}

func (ws *WServer) parseMsg(conn *UserConn, msg []byte) {
	wsReq := &pbChat.WSRequest{}
	err := proto.Unmarshal(msg, wsReq)
	if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
		return
	}
	// 验证 Token
	userID, ok := verifyToken(wsReq.Token)
	if !ok {
		glog.Infof("令牌过期或无效, token: %s, error: %v", wsReq.Token, err)
		ws.writeMsg(conn, wsReq.Action, constant.ErrParseToken.ErrCode, constant.ErrParseToken.ErrMsg, []byte{})
		return
	}
	switch wsReq.Action {
	case constant.ActionWSGetUserMaxSeq:
		req := &pbChat.GetUserMaxSeqRequest{}
		err = proto.Unmarshal(wsReq.Data, req)
		if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
			return
		}
		go ws.getUserMaxSeq(conn, userID, req)
	case constant.ActionWSPullMsgBySeqRange:
		req := &pbChat.PullMsgBySeqRangeRequest{}
		err = proto.Unmarshal(wsReq.Data, req)
		if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
			return
		}
		go ws.pullMsgBySeqRange(conn, userID, req)
	case constant.ActionWSPullMsgBySeqList:
		req := &pbChat.PullMsgBySeqListRequest{}
		err = proto.Unmarshal(wsReq.Data, req)
		if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
			return
		}
		go ws.pullMsgBySeqList(conn, userID, req)
	case constant.ActionWSPushMsg:
		req := &pbChat.PushMsgRequest{}
		err = proto.Unmarshal(wsReq.Data, req)
		if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
			return
		}
		go ws.pushMsg(conn, userID, req)
	}
}

func (ws *WServer) writeMsg(conn *UserConn, action int32, errCode int32, errMsg string, data []byte) {
	if data == nil {
		data = []byte{}
	}
	wsResp := &pbChat.WSResponse{
		Action:    action,
		ErrorCode: errCode,
		ErrorMsg:  errMsg,
		Data:      data,
	}
	b, _ := proto.Marshal(wsResp)
	ws.sendBinaryMsg(conn, b)
}

func (ws *WServer) sendTextMsg(conn *UserConn, obj []byte) {
	conn.WriteMessage(websocket.TextMessage, obj)
}

func (ws *WServer) sendBinaryMsg(conn *UserConn, obj []byte) {
	conn.WriteMessage(websocket.BinaryMessage, obj)
}

func (ws *WServer) sendCloseMsg(conn *UserConn, obj []byte) {
	conn.WriteMessage(websocket.CloseMessage, obj)
}

func (ws *WServer) sendPingMsg(conn *UserConn, obj []byte) {
	conn.WriteMessage(websocket.PingMessage, obj)
}

func (ws *WServer) handleUnmarshalError(conn *UserConn, action int32, err error) error {
	if err != nil {
		glog.Errorf("failed to unmarshal protobuf, error: %v", err)
		ws.writeMsg(conn, action, constant.ErrUnMarshalProto.ErrCode, constant.ErrUnMarshalProto.ErrMsg, []byte{})
	}
	return err
}
