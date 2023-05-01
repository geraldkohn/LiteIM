package auth

import (
	"github.com/geraldkohn/im/internal/gateway"
	"github.com/geraldkohn/im/pkg/common/logger"
	"github.com/gin-gonic/gin"
)

type paramsUserToken struct {
	UID string `json:"uid"`
}

func UserToken(c *gin.Context) {
	logger.Infof("http api user_token init ...")
}
