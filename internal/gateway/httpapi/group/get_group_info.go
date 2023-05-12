package group

import (
	"net/http"

	database "LiteIM/internal/gateway/database"
	"LiteIM/pkg/common/constant"
	"LiteIM/pkg/common/db"
	"LiteIM/pkg/common/logger"

	"github.com/gin-gonic/gin"
)

type paramsGetGroupInfo struct {
	GroupIDList []string `json:"group_id_list"`
}

type getGroupInfoResponse struct {
	GroupList []db.Group `json:"group_list"`
}

func GetGroupInfo(c *gin.Context) {
	logger.Infof("http api get_group_info init ...")
	params := paramsGetGroupInfo{}
	if err := c.BindJSON(&params); err != nil {
		logger.Errorf("http api get_group_info bad request")
		c.JSON(http.StatusBadRequest, gin.H{
			"ErrorCode": http.StatusBadRequest,
			"ErrorMsg":  err.Error(),
			"data":      getGroupInfoResponse{},
		})
		return
	}
	groupsInfo := make([]db.Group, 0)
	for _, gid := range params.GroupIDList {
		groupInfo, err := database.Databases.GetGroupByGroupID(gid)
		if err != nil {
			logger.Errorf("http api get_group_info failed to get group by groupid, [params %v], [error %v]", params, err)
			c.JSON(http.StatusBadRequest, gin.H{
				"ErrorCode": constant.ErrMysql.ErrCode,
				"ErrorMsg":  constant.ErrMysql.ErrMsg,
				"data":      getGroupInfoResponse{},
			})
			return
		}
		groupsInfo = append(groupsInfo, *groupInfo)
	}
	logger.Infof("http api get_group_info succeed, [params %v]", params)
	c.JSON(http.StatusBadRequest, gin.H{
		"ErrorCode": constant.OK.ErrCode,
		"ErrorMsg":  constant.OK.ErrMsg,
		"data":      getGroupInfoResponse{GroupList: groupsInfo},
	})
}
