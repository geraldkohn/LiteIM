package db

type Group struct {
	GroupID    string `gorm:"column:group_id"`
	Name       string `gorm:"name"`
	CreateTime int64  `gorm:"column:create_time"`
}

type GroupMember struct {
	GroupID  string `gorm:"column:group_id"`
	UserID   string `gorm:"column:user_id"`
	JoinTime int64  `gorm:"column:join_time"`
}

func (g *Group) TableName() string {
	return "group"
}

func (d *DataBases) CreateGroup(g *Group) error {
	return create(d.mysql, g)
}

func (d *DataBases) GetGroupByGroupID(groupID string) (*Group, error) {
	g := new(Group)
	conds := map[string]interface{}{"group_id": groupID}
	err := getObjectByCond(d.mysql, g, "", conds)
	return g, err
}

func (d *DataBases) CreateGroupMember(gm *GroupMember) error {
	return create(d.mysql, gm)
}

func (d *DataBases) GetGroupMemberByGroupID(groupID string) ([]*GroupMember, error) {
	groupMemberList := make([]*GroupMember, 0)
	conds := map[string]interface{}{"group_id": groupID}
	err := getListBatch(d.mysql, groupMemberList, "", -1, -1, conds)
	return groupMemberList, err
}

func (d *DataBases) GetGroupByUserID(userID string) ([]*GroupMember, error) {
	groupMemberList := make([]*GroupMember, 0)
	conds := map[string]interface{}{"user_id": userID}
	err := getListBatch(d.mysql, groupMemberList, "", -1, -1, conds)
	return groupMemberList, err
}
