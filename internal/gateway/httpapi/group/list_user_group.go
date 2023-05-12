package group

import (
	"net/http"

	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/logger"

	"github.com/gin-gonic/gin"
)

type paramsListUserGroup struct {
}

type listUserGroupResponse struct {
	GroupIDList []string `json:"group_list"`
}

func ListUserGroup(c *gin.Context) {
	logger.Infof("http api list_user_group init ...")
	userID := c.GetString("uid")
	groupmember, err := database.Databases.GetGroupByUserID(userID)
	if err != nil {
		logger.Errorf("http api list_user_group failed to create group member, [params %v], [error %v]", userID, err)
		c.JSON(http.StatusOK, gin.H{
			"ErrorCode": constant.ErrMysql.ErrCode,
			"ErrorMsg":  constant.ErrMysql.ErrMsg,
			"data":      listUserGroupResponse{},
		})
		return
	}
	groupIDList := make([]string, 0)
	for _, gm := range groupmember {
		groupIDList = append(groupIDList, gm.GroupID)
	}
	logger.Infof("http api list_user_group succeed, [params %v]", userID)
	c.JSON(http.StatusOK, gin.H{
		"ErrorCode": constant.OK.ErrCode,
		"ErrorMsg":  constant.OK.ErrMsg,
		"data":      listUserGroupResponse{GroupIDList: groupIDList},
	})
}
