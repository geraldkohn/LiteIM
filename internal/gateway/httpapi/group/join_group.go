package group

import (
	"net/http"

	database "Lite_IM/internal/gateway/database"
	"Lite_IM/pkg/common/constant"
	"Lite_IM/pkg/common/db"
	"Lite_IM/pkg/common/logger"
	"Lite_IM/pkg/utils"

	"github.com/gin-gonic/gin"
)

type paramsJoinGroup struct {
	GroupUID string `json:"group_uid"`
}

type joinGroupResponse struct {
}

func JoinGroup(c *gin.Context) {
	logger.Infof("http api join_group init ...")
	params := paramsJoinGroup{}
	if err := c.BindJSON(&params); err != nil {
		logger.Errorf("http api join_group bad reqeust")
		c.JSON(http.StatusBadRequest, gin.H{
			"ErrorCode": http.StatusBadRequest,
			"ErrorMsg":  err.Error(),
			"data":      joinGroupResponse{},
		})
		return
	}
	userID := c.GetString("uid")
	err := database.Databases.CreateGroupMember(&db.GroupMember{GroupID: params.GroupUID, UserID: userID, JoinTime: utils.GetCurrentTimestampBySecond()})
	if err != nil {
		logger.Errorf("http api join_group failed to create group member, [params %v], [error %v]", params, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrMysql.ErrCode,
			"ErrorMsg":  constant.ErrMysql.ErrMsg,
			"data":      joinGroupResponse{},
		})
		return
	}
	logger.Infof("http api join_group succeed, [params %v]", params)
	c.JSON(http.StatusOK, gin.H{
		"ErrorCode": constant.OK.ErrCode,
		"ErrorMsg":  constant.OK.ErrMsg,
		"data":      joinGroupResponse{},
	})
}
