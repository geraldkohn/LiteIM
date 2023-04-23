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
	"github.com/geraldkohn/im/pkg/common/logger"
	"github.com/geraldkohn/im/pkg/common/setting"
	"github.com/geraldkohn/im/pkg/utils"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

const (
	pingFrequency int64 = 300 // 300 seconds, 5 minutes
)

type UserConn struct {
	*websocket.Conn
	lock   *sync.Mutex // 保护 websocket 连接, 顺序写入
	userID string      // 用户 ID
	drop   int32       // 连接是否已关闭, drop: 0 未关闭; drop: 1 已关闭
}

type WServer struct {
	wsUpgrader   *websocket.Upgrader  // websocket upgrader
	connMapLock  *sync.RWMutex        // 保护 connMap
	connMap      map[string]*UserConn // 用户 ID -> 连接
	pongTimeLock *sync.RWMutex        // 保护 pongTimeMap
	pongTimeMap  map[string]int64     // 用户 ID -> 上次返回 PongMessage 的时间
	producer     *kafka.Producer      // Kafka Producer
	delTaskChan  chan string          // 终止定时任务
	exit         chan error           // 退出
}

func (ws *WServer) onInit() {
	ws.wsUpgrader = &websocket.Upgrader{
		HandshakeTimeout: time.Duration(setting.APPSetting.Websocket.Timeout) * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
	ws.connMap = make(map[string]*UserConn)
	ws.connMapLock = new(sync.RWMutex)
	ws.pongTimeLock = new(sync.RWMutex)
	ws.pongTimeMap = make(map[string]int64)
	ws.producer = kafka.NewKafkaProducer(constant.KafkaChatTopic)
	ws.delTaskChan = make(chan string, 32)
	ws.exit = make(chan error)
}

func (ws *WServer) Run() {
	go func() {
		ws.run()
	}()
	go func() {
		ws.handleTask()
	}()
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
		ws.setProfile(newConn)
		go ws.readMsg(newConn)
	}
}

// 新增 用户--Gateway 映射
func (ws *WServer) addMapping(conn *UserConn) {
	endpoint := fmt.Sprintf("%s:%d", IP, setting.APPSetting.RPC.GatewayRPCPort)
	_, err := db.DB.SetOnlineUserGatewayEndpoint(conn.userID, endpoint)
	logger.Errorf("Failed to bind online user to gateway endpoint | error %v", err)
}

// 删除 用户--Gateway 映射
func (ws *WServer) delMapping(conn *UserConn) {
	_, err := db.DB.DeleteOnlineUserGatewayEndpoint(conn.userID)
	logger.Errorf("Failed to Debind online user to gateway endpoint | error %v", err)
}

func (ws *WServer) addConnMap(conn *UserConn) {
	ws.connMapLock.Lock()
	ws.connMap[conn.userID] = conn
	ws.connMapLock.Unlock()
}

func (ws *WServer) delConnMap(conn *UserConn) {
	atomic.StoreInt32(&conn.drop, 1)
	ws.connMapLock.Lock()
	delete(ws.connMap, conn.userID)
	ws.connMapLock.Unlock()
}

func (ws *WServer) readMsg(conn *UserConn) {
	defer glog.Infof("WServer readMsg return, conn: %v", conn)
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			glog.Errorf("read message error: %v", err)
			ws.delConnMap(conn)
			ws.delMapping(conn)
			ws.sendCloseMsg(conn, []byte{}) // 关闭 websocket
			return
		}
		if messageType == websocket.CloseMessage {
			glog.Infof("websocket receive close message %s", conn.userID)
			ws.delConnMap(conn)
			ws.delMapping(conn)
			return
		}
		if atomic.LoadInt32(&conn.drop) == 1 {
			glog.Infof("websocket connection has been closed %s", conn.userID)
			return
		}
		// if messageType == websocket.PongMessage {
		// 	// 收到客户端的 Pong Message, 需要更新客户端到期时间
		// 	// 这里可以仿照 nsq 来做
		// 	// TODO
		// 	continue
		// }
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

// 添加 websocket 连接 PongHanler
func (ws *WServer) setPongHandler(conn *UserConn) {
	conn.Conn.SetPongHandler(func(appData string) error {
		now := utils.GetCurrentTimestampBySecond()
		ws.pongTimeLock.Lock()
		ws.pongTimeMap[appData] = now
		ws.pongTimeLock.Unlock()
		return nil
	})
}

// 设置 websocket 连接属性
func (ws *WServer) setProfile(newConn *UserConn) {
	ws.addConnMap(newConn)     // 加入 Map
	ws.addMapping(newConn)     // 注册 redis
	ws.addTask(newConn)        // 添加定时任务, 定时发送 Ping Message
	ws.setPongHandler(newConn) // 设置 PongHandler
}

// 退出
func (ws *WServer) Exit() {
	close(ws.exit)
	ws.connMapLock.Lock()
	for _, v := range ws.connMap {
		ws.sendCloseMsg(v, []byte("close"))
	}
	ws.connMapLock.Unlock()
}
