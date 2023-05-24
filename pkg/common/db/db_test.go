package db

import (
	"LiteIM/pkg/utils"
	"testing"
)

func TestMysqlUserModel(t *testing.T) {
	var d DataBases
	d.initMySQL(MysqlConfig{
		Addr:     "127.0.0.1:3306",
		Username: "root",
		Password: "root",
	})
	uid := utils.GenerateUID()
	u := &User{
		UID:        uid,
		Name:       "user-name",
		Password:   "user-password",
		CreateTime: utils.GetCurrentTimestampByMill(),
	}
	err := d.CreateUser(u)
	if err != nil {
		t.FailNow()
	}
	user, err := d.GetUserByUid(uid)
	if err != nil {
		t.FailNow()
	}
	if user.Name != u.Name {
		t.FailNow()
	}
}
