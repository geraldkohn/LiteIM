package gateway

import (
	"net/http"

	"LiteIM/internal/gateway/httpapi/auth"
	"LiteIM/internal/gateway/httpapi/group"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/logger"
	"LiteIM/pkg/utils"

	"github.com/gin-gonic/gin"
)

type HServer struct {
	server *http.Server
}

func (hs *HServer) Run() {
	err := hs.server.ListenAndServe()
	if err != nil {
		panic("HTTP Server: set up error: " + err.Error())
	}
}

func initRouter(r *gin.Engine) {
	r.Use(utils.CorsHandler())
	r.Use(TokenHandler())

	// user routing group, which handles user registration and login services
	// userRouterGroup := r.Group("/user")

	//friend routing group
	// friendRouterGroup := r.Group("/friend")

	//group related routing group
	groupRouterGroup := r.Group("/group")
	{
		groupRouterGroup.POST("create_group", group.CreateGroup)
		groupRouterGroup.POST("join_group", group.JoinGroup)
		groupRouterGroup.POST("get_group_info", group.GetGroupInfo)
		groupRouterGroup.POST("list_user_group", group.ListUserGroup)
	}
	//certificate
	authRouterGroup := r.Group("/auth")
	{
		authRouterGroup.POST("/user_register", auth.UserRegister)
		authRouterGroup.POST("/user_login", auth.UserLogin)
	}
}

func TokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			logger.Logger.Errorf("token handler failed to get Authorization")
			c.JSON(http.StatusOK, gin.H{
				"ErrorCode": constant.ErrGetToken.ErrCode,
				"ErrorMsg":  constant.ErrGetToken.ErrMsg,
				"data":      nil,
			})
			return
		}
		claim, err := utils.ParseToken(tokenStr)
		if err != nil {
			logger.Logger.Errorf("token handler failed to parse Authorization")
			c.JSON(http.StatusOK, gin.H{
				"ErrorCode": constant.ErrParseToken.ErrCode,
				"ErrorMsg":  constant.ErrParseToken.ErrMsg,
				"data":      nil,
			})
			return
		}
		if claim == "" {
			logger.Logger.Infof("token handler parsed an unavailble token")
			c.JSON(http.StatusOK, gin.H{
				"ErrorCode": constant.ErrUnavailableToken.ErrCode,
				"ErrorMsg":  constant.ErrUnavailableToken.ErrMsg,
				"data":      nil,
			})
			return
		}
		logger.Logger.Infof("token handler parsed an unavailble token")
		c.Set("uid", claim)
		c.Next()
	}
}
