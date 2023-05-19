package gateway

import (
	"sync/atomic"
	"time"

	"LiteIM/pkg/common/cronjob"
	"LiteIM/pkg/common/logger"

	"github.com/gorilla/websocket"
)

const (
	heartBeat = 60  // 心跳时长: 一分钟
	aliveTime = 300 // 存活时长: 三分钟
)

// 初始化定时任务调度器
func (ws *WServer) keepAlive() {
	logger.Logger.Infof("Start keeping alive!")
	ws.scheduler.Start() // 阻塞方法
}

// 与客户端保持连接, 如长时间连接失败则自动清理
func (ws *WServer) addToScheduler(conn *UserConn) {
	dofunc := func() {
		if time.Now().Unix()-conn.lastPongTime >= aliveTime {
			ws.scheduler.Remove(conn.uid)
			return
		}
		conn.WriteMessage(websocket.PingMessage, []byte(conn.uid)) // 忽略掉错误
	}
	job := cronjob.NewJober(conn.uid, dofunc, time.Duration(heartBeat*time.Second), 0)
	logger.Logger.Infof("job of usrID %v is added to scheduler", conn.uid)
	ws.scheduler.Add(job) // 加入调度
}

func (ws *WServer) clearConn(conn *UserConn) {
	// 只在连接为被关闭时执行, 防止被关闭两次
	if atomic.LoadInt32(&conn.drop) == 0 {
		logger.Logger.Infof("close and clear websocket connection, conn's userID: %v", conn.uid)
		atomic.StoreInt32(&conn.drop, 1) // 设置连接的状态为已关闭
		ws.connMapLock.Lock()
		delete(ws.connMap, conn.uid) // 删除连接池
		ws.connMapLock.Unlock()
		conn.Close()                  // 关闭连接
		ws.scheduler.Remove(conn.uid) // 停止定时任务
		ws.offline(conn.uid)          // 清除 redis 用户会话
	}
}
