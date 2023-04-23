package gateway

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/geraldkohn/im/pkg/common/setting"
	"github.com/gin-gonic/gin"
)

type HServer struct {
	server *http.Server
}

func (hs *HServer) onInit() {
	router := gin.New()
	initRouter(router)
	hs.server = &http.Server{
		Addr:         ":" + strconv.Itoa(setting.APPSetting.HTTP.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(setting.APPSetting.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(setting.APPSetting.HTTP.WriteTimeout) * time.Second,
	}
}

func (hs *HServer) run() {
	addr := ":" + strconv.Itoa(setting.APPSetting.HTTP.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic("HTTP Server Listen failed: " + addr)
	}
	err = hs.server.Serve(l)
	if err != nil {
		panic("HTTP Server: set up error: " + err.Error())
	}
}

func initRouter(router *gin.Engine) {
	api := router.Group("api")
	api.Use()
	{

	}
}
