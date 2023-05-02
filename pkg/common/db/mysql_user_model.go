package db

type User struct {
	UID        string `gorm:"column:uid;primaryKey"`
	Name       string `gorm:"column:name"`
	Password   string `gorm:"column:password"`
	CreateTime int64  `gorm:"column:create_time;comment:创建时间, 以秒为单位"`
}

func (u *User) TableName() string {
	return "user"
}

func (d *DataBases) CreateUser(u *User) error {
	return create(d.mysql, u)
}

func (d *DataBases) GetUserByUid(uid string) (*User, error) {
	u := new(User)
	conds := map[string]interface{}{"uid": uid}
	err := getObjectByCond(d.mysql, u, "", conds)
	return u, err
}

func (d *DataBases) GetUserByName(name string) (*User, error) {
	u := new(User)
	conds := map[string]interface{}{"name": name}
	err := getObjectByCond(d.mysql, u, "", conds)
	return u, err
}
