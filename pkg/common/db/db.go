package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	mongodbDataBase = "mongo"
)

type DataBases struct {
	redis *redis.Client
	mysql *gorm.DB
	mongo *mongo.Client
}

type MysqlConfig struct {
	Addr     string
	Username string
	Password string
}

type RedisConfig struct {
	Addr     string
	Username string
	Password string
}

type MongodbConfig struct {
	Addr     []string
	Username string
	Password string
}

func NewDataBases(mc MysqlConfig, rc RedisConfig, mdc MongodbConfig) *DataBases {
	d := &DataBases{}
	d.initMySQL(mc)
	d.initRedis(rc)
	d.initMongoDB(mdc)
	return d
}

func (d *DataBases) initMySQL(mc MysqlConfig) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=ture&loc=Local", mc.Username, mc.Password, mc.Addr, "mysql")
	d.mysql, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to initialize mysql connection" + err.Error())
	}
}

func (d *DataBases) initRedis(rc RedisConfig) {
	d.redis = redis.NewClient(&redis.Options{
		Addr:     rc.Addr,
		Username: rc.Username,
		Password: rc.Password,
	})

	_, err := d.redis.Ping(context.Background()).Result()
	if err != nil {
		panic("Failed to initialize connection with redis" + err.Error())
	}
}

func (d *DataBases) initMongoDB(mdc MongodbConfig) {
	var err error
	var cred options.Credential
	cred.Username = mdc.Username
	cred.Password = mdc.Password
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	dsn := fmt.Sprintf("mongodb://%s", mdc.Addr[0])
	d.mongo, err = mongo.Connect(ctx, options.Client().ApplyURI(dsn).SetAuth(cred))
	if err != nil {
		panic("Failed to initialize to connect to mongodb" + err.Error())
	}
}
