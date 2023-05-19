package auth

import (
	"net/http"

	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/logger"
	"LiteIM/pkg/utils"

	"github.com/gin-gonic/gin"
)

type paramsUserRegister struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type userRegisterResponse struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func UserRegister(c *gin.Context) {
	logger.Logger.Infof("HTTP API UserRegister init ...")
	params := paramsUserRegister{}
	if err := c.BindJSON(&params); err != nil {
		logger.Logger.Errorf("http api user_register bind json failed, [params %v], [error %v]", params, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"ErrorCode": http.StatusBadRequest,
			"ErrorMsg":  err.Error(),
			"data":      userRegisterResponse{UID: "", Name: params.Name},
		})
		return
	}
	user, err := database.Databases.GetUserByName(params.Name)
	if err != nil {
		logger.Logger.Errorf("http api user_register call mysql failed, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrMysql.ErrCode,
			"ErrorMsg":  constant.ErrMysql.ErrMsg,
			"data":      userRegisterResponse{UID: "", Name: params.Name},
		})
		return
	}
	if user.Name == params.Name { // 名字需要保持唯一
		logger.Logger.Infof("http api user_register try to register an existed name, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrAccountExists.ErrCode,
			"ErrorMsg":  constant.ErrAccountExists.ErrMsg,
			"data":      userRegisterResponse{UID: "", Name: params.Name},
		})
		return
	}
	// 生成 uid
	uid := utils.GenerateUID()
	user.UID = uid
	user.Name = params.Name
	user.Password = params.Password
	user.CreateTime = utils.GetCurrentTimestampBySecond()
	// 创建用户
	err = database.Databases.CreateUser(user)
	if err != nil {
		logger.Logger.Errorf("http api user_register create user is failed, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrAccountExists.ErrCode,
			"ErrorMsg":  constant.ErrAccountExists.ErrMsg,
			"data":      userRegisterResponse{UID: "", Name: params.Name},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ErrorCode": constant.ErrAccountExists.ErrCode,
		"ErrorMsg":  constant.ErrAccountExists.ErrMsg,
		"data":      userRegisterResponse{UID: uid, Name: params.Name},
	})
}
