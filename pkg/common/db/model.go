package db

import (
	"context"

	"github.com/geraldkohn/im/pkg/common/setting"
	"github.com/golang/glog"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var DB DataBases

type DataBases struct {
	redis redis.ClusterClient
	mysql *gorm.DB
	mongo *mongo.Database
}

func (d *DataBases) Init() {
	d.initRedis()
	d.initMySQL()
	d.initMongoDB()
}

func (d *DataBases) initMySQL() {

}

func (d *DataBases) initRedis() {
	d.redis = *redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    setting.APPSetting.Redis.Addr,
		Username: setting.APPSetting.Redis.Username,
		Password: setting.APPSetting.Redis.Password,
	})

	_, err := d.redis.Ping(context.Background()).Result()
	if err != nil {
		glog.Errorf("Setting up redis cluster client failed, error: %v", err)
	}
}

func (d *DataBases) initMongoDB() {

}
