package auth

import (
	"net/http"

	database "Lite_IM/internal/gateway/database"
	"Lite_IM/pkg/common/constant"
	"Lite_IM/pkg/common/logger"
	"Lite_IM/pkg/utils"

	"github.com/gin-gonic/gin"
)

type paramsUserLogin struct {
	Name     string `json:"uid"`
	Password string `json:"password"`
}

type userLoginResponse struct {
	UID   string `json:"uid"`
	Token string `json:"token"`
}

func UserLogin(c *gin.Context) {
	logger.Infof("http api user_login init ...")
	params := paramsUserLogin{}
	if err := c.BindJSON(&params); err != nil {
		logger.Errorf("http api user_login bind json failed, [params %v], [error %v]", params, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"ErrorCode": http.StatusBadRequest,
			"ErrorMsg":  err.Error(),
			"data":      userLoginResponse{},
		})
		return
	}
	user, err := database.Databases.GetUserByName(params.Name)
	if err != nil {
		logger.Errorf("http api user_login failed to connect to mysql, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrMysql.ErrCode,
			"ErrorMsg":  constant.ErrMysql.ErrMsg,
			"data":      userLoginResponse{},
		})
		return
	}
	token, err := utils.GenerateToken(user.UID)
	if err != nil {
		logger.Errorf("http api user_login failed to generate token, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrCreateToken.ErrCode,
			"ErrorMsg":  constant.ErrCreateToken.ErrMsg,
			"data":      userLoginResponse{UID: user.UID},
		})
		return
	}
	logger.Infof("http api user_login succeed to generate token, [params %v], [error %v]", params, err)
	c.JSON(http.StatusOK, gin.H{
		"ErrorCode": constant.OK.ErrCode,
		"ErrorMsg":  constant.OK.ErrMsg,
		"data":      userLoginResponse{UID: user.UID, Token: token},
	})
}
