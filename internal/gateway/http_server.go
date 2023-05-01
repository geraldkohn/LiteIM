package gateway

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/geraldkohn/im/pkg/common/setting"
	"github.com/geraldkohn/im/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type HServer struct {
	server *http.Server
}

func (hs *HServer) onInit() {
	router := gin.Default()
	initRouter(router)
	hs.server = &http.Server{
		Addr:         ":" + strconv.Itoa(viper.GetInt("HTTPPort")),
		Handler:      router,
		ReadTimeout:  time.Duration(viper.GetInt("HTTPReadTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("HTTPWriteTimeout")) * time.Second,
	}
}

func (hs *HServer) Run() {
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

func initRouter(r *gin.Engine) {
	r.Use(utils.CorsHandler())
	// user routing group, which handles user registration and login services
	userRouterGroup := r.Group("/user")
	{
		userRouterGroup.POST("/update_user_info", user.UpdateUserInfo)
		userRouterGroup.POST("/get_user_info", user.GetUserInfo)
	}
	//friend routing group
	friendRouterGroup := r.Group("/friend")
	{
		friendRouterGroup.POST("/get_friends_info", friend.GetFriendsInfo)
		friendRouterGroup.POST("/add_friend", friend.AddFriend)
		friendRouterGroup.POST("/get_friend_apply_list", friend.GetFriendApplyList)
		friendRouterGroup.POST("/get_self_apply_list", friend.GetSelfApplyList)
		friendRouterGroup.POST("/get_friend_list", friend.GetFriendList)
		friendRouterGroup.POST("/add_blacklist", friend.AddBlacklist)
		friendRouterGroup.POST("/get_blacklist", friend.GetBlacklist)
		friendRouterGroup.POST("/remove_blacklist", friend.RemoveBlacklist)
		friendRouterGroup.POST("/delete_friend", friend.DeleteFriend)
		friendRouterGroup.POST("/add_friend_response", friend.AddFriendResponse)
		friendRouterGroup.POST("/set_friend_comment", friend.SetFriendComment)
		friendRouterGroup.POST("/is_friend", friend.IsFriend)
		friendRouterGroup.POST("/import_friend", friend.ImportFriend)
	}
	//group related routing group
	groupRouterGroup := r.Group("/group")
	{
		groupRouterGroup.POST("/create_group", group.CreateGroup)
		groupRouterGroup.POST("/set_group_info", group.SetGroupInfo)
		groupRouterGroup.POST("join_group", group.JoinGroup)
		groupRouterGroup.POST("/quit_group", group.QuitGroup)
		groupRouterGroup.POST("/group_application_response", group.ApplicationGroupResponse)
		groupRouterGroup.POST("/transfer_group", group.TransferGroupOwner)
		groupRouterGroup.POST("/get_group_applicationList", group.GetGroupApplicationList)
		groupRouterGroup.POST("/get_groups_info", group.GetGroupsInfo)
		groupRouterGroup.POST("/kick_group", group.KickGroupMember)
		groupRouterGroup.POST("/get_group_member_list", group.GetGroupMemberList)
		groupRouterGroup.POST("/get_group_all_member_list", group.GetGroupAllMember)
		groupRouterGroup.POST("/get_group_members_info", group.GetGroupMembersInfo)
		groupRouterGroup.POST("/invite_user_to_group", group.InviteUserToGroup)
		groupRouterGroup.POST("/get_joined_group_list", group.GetJoinedGroupList)
	}
	//certificate
	authRouterGroup := r.Group("/auth")
	{
		authRouterGroup.POST("/user_register", apiAuth.UserRegister)
		authRouterGroup.POST("/user_token", apiAuth.UserToken)
	}
	//Third service
	thirdGroup := r.Group("/third")
	{
		thirdGroup.POST("/tencent_cloud_storage_credential", apiThird.TencentCloudStorageCredential)
	}
	//Message
	chatGroup := r.Group("/chat")
	{
		chatGroup.POST("/newest_seq", apiChat.UserGetSeq)
		chatGroup.POST("/pull_msg", apiChat.UserPullMsg)
		chatGroup.POST("/send_msg", apiChat.UserSendMsg)
		chatGroup.POST("/pull_msg_by_seq", apiChat.UserPullMsgBySeqList)
	}
	//Manager
	managementGroup := r.Group("/manager")
	{
		managementGroup.POST("/delete_user", manage.DeleteUser)
		managementGroup.POST("/send_msg", manage.ManagementSendMsg)
		managementGroup.POST("/get_all_users_uid", manage.GetAllUsersUid)
	}
}
