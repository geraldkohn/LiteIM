package group

import (
	"net/http"

	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/logger"
	"LiteIM/pkg/utils"

	"github.com/gin-gonic/gin"
)

type paramsCreateGroup struct {
	GroupName string `json:"group_name"`
}

type createGroupResponse struct {
	GroupUID string `json:"group_uid`
}

func CreateGroup(c *gin.Context) {
	logger.Logger.Infof("http api create_group init ...")
	params := paramsCreateGroup{}
	if err := c.BindJSON(&params); err != nil {
		logger.Logger.Errorf("http api create_group bad request")
		c.JSON(http.StatusBadRequest, gin.H{
			"ErrorCode": http.StatusBadRequest,
			"ErrorMsg":  err.Error(),
			"data":      createGroupResponse{},
		})
		return
	}
	groupID := utils.GenerateUID()
	err := database.Databases.CreateGroup(&db.Group{GroupID: groupID, Name: params.GroupName, CreateTime: utils.GetCurrentTimestampBySecond()})
	if err != nil {
		logger.Logger.Errorf("http api create_group failed to create group, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrMysql.ErrCode,
			"ErrorMsg":  constant.ErrMysql.ErrMsg,
			"data":      createGroupResponse{},
		})
		return
	}
	logger.Logger.Infof("http api create_group succeed, [params %v]", params)
	c.JSON(http.StatusOK, gin.H{
		"ErrorCode": constant.OK.ErrCode,
		"ErrorMsg":  constant.OK.ErrMsg,
		"data":      createGroupResponse{GroupUID: groupID},
	})
}
