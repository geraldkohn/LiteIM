package gateway

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	pbChat "LiteIM/internal/api/rpc/chat"
	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/cronjob"
	"LiteIM/pkg/common/kafka"
	"LiteIM/pkg/common/logger"
	"LiteIM/pkg/utils"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type UserConn struct {
	*websocket.Conn
	uid          string // 用户 ID
	drop         int32  // 连接是否已关闭, drop: 0 未关闭; drop: 1 已关闭
	lastPongTime int64  // 客户端上次返回 Pong 的时间点
}

type WServer struct {
	port        int                  // websocket port
	wsUpgrader  *websocket.Upgrader  // websocket upgrader
	connMapLock *sync.RWMutex        // 保护 connMap
	connMap     map[string]*UserConn // 用户 ID -> 连接
	producer    *kafka.Producer      // Kafka Producer
	scheduler   cronjob.Scheduler    // 定时任务调度器
	exit        chan error           // 退出
}

func (ws *WServer) Run() {
	logger.Logger.Infof("Websocket Server Running...")
	go ws.keepAlive()
	ws.serveWs()
}

func (ws *WServer) serveWs() {
	logger.Logger.Infof("Start listening websocket request!")
	http.HandleFunc("/", ws.wsHandler)
	err := http.ListenAndServe(":"+strconv.Itoa(ws.port), nil)
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
		newConn := &UserConn{conn, query["userID"][0], 0, utils.GetCurrentTimestampBySecond()}
		ws.setProfile(newConn)
		logger.Logger.Infof("a new websocket connection is set, userID: %v", newConn.uid)
		go ws.readMsg(newConn)
	}
}

// 新增 用户--Gateway 映射
func (ws *WServer) online(userID string) {
	endpoint := fmt.Sprintf("%s:%d", nodeIP, ws.port)
	_, err := database.Databases.SetOnlineUserGatewayEndpoint(userID, endpoint)
	logger.Logger.Errorf("Failed to bind online user to gateway endpoint | error %v", err)
}

// 删除 用户--Gateway 映射
func (ws *WServer) offline(userID string) {
	_, err := database.Databases.DeleteOnlineUserGatewayEndpoint(userID)
	logger.Logger.Errorf("Failed to Debind online user to gateway endpoint | error %v", err)
}

// 读取 websocket 连接内容
func (ws *WServer) readMsg(conn *UserConn) {
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if err.Error() == websocket.ErrBadHandshake.Error() { // 握手失败, 直接返回
				return
			} else if err.Error() == websocket.ErrReadLimit.Error() {
				continue
			}
			logger.Logger.Errorf("websocket read an error: %v", err)
			break
		}
		if messageType == websocket.CloseMessage {
			glog.Infof("websocket receive close message %s", conn.uid)
			break
		}
		if atomic.LoadInt32(&conn.drop) == 1 { // 连接的状态被设置为已关闭
			glog.Infof("websocket connection has been closed %s", conn.uid)
			break
		}
		ws.parseMsg(conn, msg)
	}
	ws.clearConn(conn)
	logger.Logger.Infof("websocket connection is closed and clear source, conn userID: %v", conn.uid)
}

// 异步逻辑, 处理信息
func (ws *WServer) parseMsg(conn *UserConn, msg []byte) {
	wsReq := &pbChat.WSRequest{}
	err := proto.Unmarshal(msg, wsReq)
	if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
		return
	}
	// 验证 Token
	userID, ok := verifyToken(wsReq.Token)
	if !ok {
		glog.Infof("token is unavilable, token: %s, error: %v", wsReq.Token, err)
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
	case constant.ActionWSPushMsgToServer:
		req := &pbChat.PushMsgRequest{}
		err = proto.Unmarshal(wsReq.Data, req)
		if err = ws.handleUnmarshalError(conn, wsReq.Action, err); err != nil {
			return
		}
		go ws.pushMsg(conn, userID, req)
	}
}

func (ws *WServer) writeMsg(conn *UserConn, action int32, errCode int32, errMsg string, data []byte) error {
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
	return conn.WriteMessage(websocket.BinaryMessage, b)
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
		conn.lastPongTime = utils.GetCurrentTimestampBySecond()
		return nil
	})
}

// 设置 websocket 连接属性
func (ws *WServer) setProfile(newConn *UserConn) {
	ws.connMapLock.Lock()
	ws.connMap[newConn.uid] = newConn
	ws.connMapLock.Unlock()
	ws.online(newConn.uid)     // 用户会话注册到 redis
	ws.addToScheduler(newConn) // 尝试保活
	ws.setPongHandler(newConn) // 设置 PongHandler
}

// 根据用户 uid 获取连接
func (ws *WServer) getUserConn(uid string) *UserConn {
	ws.connMapLock.RLock()
	conn := ws.connMap[uid]
	ws.connMapLock.RUnlock()
	return conn
}

// 退出
func (ws *WServer) Exit() {
	close(ws.exit)
	ws.connMapLock.Lock()
	for _, conn := range ws.connMap {
		conn.Close()
	}
	ws.connMapLock.Unlock()
	ws.scheduler.Stop()
}
