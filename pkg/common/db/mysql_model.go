package db

type User struct {
	UserID string `json:"user_id"`
}

type Group struct {
	GroupID    string `json:"group_id" grom:"group_id"`
	CreateTime int64  `json:"create_time"`
}

// TODO
func (d *DataBases) GetGroupAllNumber(groupID string) ([]*User, error) {
	return nil, nil
}
