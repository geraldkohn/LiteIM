package db

import (
	"context"
	"strconv"
	"time"
)

const (
	prefixUserIncrSeq  = "REDIS_USER_INCR_SEQ:" // user incr seq
	prefixUserMinSeq   = "REDIS_USER_MIN_SEQ:"
	prefixUserPlatform = "REDIS_USER_PLATFORM:"
)

// Perform seq auto-increment operation of user messages
// 更新用户最新消息序列号
func (d *DataBases) IncrUserSeq(uid string) (int64, error) {
	key := prefixUserIncrSeq + uid
	// Redis中的 INCR 命令可以将指定的key对应的值增加1.
	// 关于 INCR 命令的最大值, 它受Redis中整数类型的最大值限制, 也就是32位有符号整数的最大值, 即2147483647
	return d.redis.Incr(context.Background(), key).Result()
}

// Get the largest Seq
// 获取用户最新的消息序列号
func (d *DataBases) GetUserMaxSeq(uid string) (int64, error) {
	key := prefixUserIncrSeq + uid
	val, err := d.redis.Get(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	seq, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return int64(seq), nil
}

// Set the user's minimum seq
// 设置用户最小的消息序列号, 表示 < minSeq 的序列号都已经被消费
func (d *DataBases) SetUserMinSeq(uid string, minSeq int64) error {
	key := prefixUserMinSeq + uid
	_, err := d.redis.Set(context.Background(), key, strconv.Itoa(int(minSeq)), 0).Result()
	return err
}

// Get the smallest Seq
// 设置用户最小的消息序列号, 还没有被消息的消息
func (d *DataBases) GetUserMinSeq(uid string) (int64, error) {
	key := prefixUserMinSeq + uid
	val, err := d.redis.Get(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	seq, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return int64(seq), nil
}

// platform 就是和用户形成 Websocket 连接的服务器的唯一识别值
// 这样就将服务器和用户绑定在一起
// Store userid and platform class to redis, ttl is second
func (d *DataBases) SetUserIDAndPlatform(userID, platform string, ttl int64) error {
	key := prefixUserPlatform + userID
	_, err := d.redis.Set(context.Background(), key, platform, time.Duration(ttl)*time.Second).Result()
	return err
}

// Check exists userid and platform class from redis
func (d *DataBases) ExistsUserIDAndPlatform(userID, platformClass string) (interface{}, error) {
	key := prefixUserPlatform + platformClass
	return d.redis.Exists(context.Background(), key).Result() // 是否过期, platform 需要不断续约
}

// Get gateway entpoint binding to online user
func (d *DataBases) GetOnlineUserGatewayEndpoint(uid string) (string, error) {
	key := prefixUserPlatform + uid
	return d.redis.Get(context.Background(), key).Result()
}

// Bind gateway endpoint to online user
func (d *DataBases) SetOnlineUserGatewayEndpoint(uid string, endpoint string) (string, error) {
	key := prefixUserPlatform + uid
	return d.redis.Set(context.Background(), key, endpoint, 0).Result()
}

// Delete gateway entpoint to online user
func (d *DataBases) DeleteOnlineUserGatewayEndpoint(uid string) (int64, error) {
	key := prefixUserPlatform + uid
	return d.redis.Del(context.Background(), key).Result()
}

// uid->最大序列号
// uid->最小未消费序列号
// uid->gatewayIP 用户与哪个服务器连接了
