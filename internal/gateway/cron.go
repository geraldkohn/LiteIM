package gateway

import (
	"fmt"
	"time"

	"github.com/geraldkohn/im/pkg/common/crontab"
	"github.com/geraldkohn/im/pkg/utils"
)

func (ws *WServer) addTask(conn *UserConn) {
	fn := func() error {
		// 先检查上次的 PingMessage 得到回应了么
		ws.connMapLock.RLock()
		if ws.pongTimeMap[conn.userID] != 0 {
			now := utils.GetCurrentTimestampBySecond()
			// 上次没有得到回应, 五分钟发送一次 PingMessage
			if now-ws.pongTimeMap[conn.userID] >= pingFrequency {
				ws.delConnMap(conn)
				ws.delTaskChan <- conn.userID
				return nil
			}
		}
		// 发起 PingMessage
		ws.sendPingMsg(conn, []byte(conn.userID))
		return nil
	}
	minute := time.Now().Minute()
	spec := fmt.Sprintf("%d/%d * * * *", minute, 5)
	task := crontab.NewTask(conn.userID, spec, fn)
	crontab.AddTask(conn.userID, task)
}

// 阻塞方法
func (ws *WServer) handleTask() {
	for {
		select {
		case taskName := <-ws.delTaskChan:
			crontab.DeleteTask(taskName)
		case <-ws.exit:
			return
		}
	}
}
