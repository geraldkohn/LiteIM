package db

import (
	"time"
)

type User struct {
	UID        string    `gorm:"column:uid;primaryKey"`
	Name       string    `gorm:"column:name"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (u *User) TableName() string {
	return "user"
}

func (d *DataBases) CreateUser(u *User) error {
	return create(d.mysql, u)
}

func (d *DataBases) GetUserByUid(uid string) *User {
	u := new(User)
	conds := map[string]interface{}{"uid": uid}
	getObjectByCond(d.mysql, u, "", conds)
	return u
}
