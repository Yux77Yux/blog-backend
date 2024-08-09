package redis_utils

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yux77yux/blog-backend/config"
	"github.com/yux77yux/blog-backend/internal/model"
	"github.com/yux77yux/blog-backend/utils/log_utils"
)

/*
func ExistsFieldInRedix(key string) (bool, error) {
	val, err := config.RDB.Exists(config.CTX, key).Result()
	if err != nil {
		return false, fmt.Errorf("checking key is not exists: %v", err)
	}
	return val > 0, nil
}

func StoreFieldInRedis(uid string, key string, value interface{}) error {
	return config.RDB.Set(config.CTX, key, value, 0).Err()
}

func RetrieveFieldInRedis(uid string, key string) (string, error) {
	value, err := config.RDB.Get(config.CTX, key).Result()
	if err != nil {
		return "", fmt.Errorf("error retrieving Field value from Redis: %v", err)
	}
	return value, nil
}*/

func RemoveExpiredJti() error {
	// 获取当前时间的时间戳
	now := float64(time.Now().Unix())

	// 删除分数小于当前时间戳的所有元素
	_, err := config.RDB.ZRemRangeByScore(config.CTX, "blacklist", "-inf", fmt.Sprintf("%f", now)).Result()
	return err
}

func ScheduleCleanup() {
	ticker := time.NewTicker(time.Minute * 40)
	defer ticker.Stop()

	for range ticker.C {
		if err := RemoveExpiredJti(); err != nil {
			log_utils.Logger.Printf("Error: redis_utils cleaning up expired members: %v", err)
		}
	}
}

func AddJtiToBlacklist(jti string, expiration float64) error {
	listName := "blacklist"

	z := &redis.Z{
		Score:  expiration,
		Member: jti,
	}

	_, err := config.RDB.ZAdd(config.CTX, listName, z).Result()
	if err != nil {
		return fmt.Errorf("redis_utils AddJtiToBlacklist: %v", err)
	}

	return nil
}

func IsJtiBlacklisted(jti string) (bool, error) {
	listName := "blacklist"

	// 检查 jti 是否在有序集合中
	score, err := config.RDB.ZScore(config.CTX, listName, jti).Result()
	if err == redis.Nil {
		return false, nil // jti 不在黑名单中
	}
	if err != nil {
		return false, fmt.Errorf("redis_utils IsJtiBlacklisted ZScore: %v", err)
	}

	// 检查 jti 是否过期
	return score > float64(time.Now().Unix()), nil
}

func ExistsSigningKeyInRedix(uid string) (bool, error) {
	field := fmt.Sprintf("signingKey_%s", uid)
	val, err := config.RDB.Exists(config.CTX, field).Result()
	if err != nil {
		return false, fmt.Errorf("redis_utils checking key is not exists: %v", err)
	}
	return val > 0, nil
}

func StoreSigningKeyInRedis(uid string, signingKey []byte) error {
	field := fmt.Sprintf("signingKey_%s", uid)
	encodedKey := base64.StdEncoding.EncodeToString(signingKey)
	return config.RDB.Set(config.CTX, field, encodedKey, 0).Err()
}

func RetrieveSigningKeyInRedis(uid string) ([]byte, error) {
	field := fmt.Sprintf("signingKey_%s", uid)
	value, err := config.RDB.Get(config.CTX, field).Result()
	if err != nil {
		return nil, fmt.Errorf("redis_utils retrieving signing key from Redis: %v", err)
	}

	signingKey, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("redis_utils DecodeString: %v", err)
	}
	return signingKey, nil
}

func ModifyUserField(uid string, field string, value interface{}) error {
	key := fmt.Sprintf("user:%s", uid)

	_, err := config.RDB.HSet(config.CTX, key, field, value).Result()
	if err != nil {
		return fmt.Errorf("redis_utils UpdateUserFieldInRedis HSet: %v", err)
	}
	return nil
}

func StoreUserInRedis(user *model.UserIncidental) error {
	// 使用用户 ID 作为 Redis Hash 的键
	key := fmt.Sprintf("user:%s", user.Uid)

	// 使用 HSET 命令存储用户数据
	_, err := config.RDB.HSet(config.CTX, key, map[string]interface{}{
		"Uid":        user.Uid,
		"Name":       user.Name,
		"Profile":    user.Profile,
		"Bio":        user.Bio,
		"Popularity": user.Popularity,
	}).Result()

	if err != nil {
		return fmt.Errorf("redis_utils StoreUserInRedis HSet: %v", err)
	}

	return nil
}

func GetUserFromRedis(uid string) (*model.UserIncidental, error) {
	key := fmt.Sprintf("user:%s", uid)

	// 使用 HGETALL 命令获取用户数据
	data, err := config.RDB.HGetAll(config.CTX, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis_utils GetUserFromRedis HGetAll: %v", err)
	}

	// 将数据反序列化为 UserIncidental 对象
	popularity, err := strconv.ParseFloat(data["Popularity"], 32)
	if err != nil {
		return nil, fmt.Errorf("redis_utils GetUserFromRedis ParseFloat: %v", err)
	}

	status, err := GetUserOnline(uid)
	if err != nil {
		return nil, fmt.Errorf("redis_utils GetUserFromRedis GetUserOnline: %v", err)
	}

	// 将数据反序列化为 UserIncidental 对象
	user := &model.UserIncidental{
		Uid:        data["Uid"],
		Name:       data["Name"],
		Profile:    data["Profile"],
		Bio:        data["Bio"],
		Status:     status,
		Popularity: float32(popularity),
	}

	return user, nil
}

func SetUserOnline(uid string, online bool) error {
	uid_int, err := strconv.Atoi(uid)
	if err != nil {
		return fmt.Errorf("redis_utils Format transform failure: %v", err)
	}
	uid_int64 := int64(uid_int)

	key := "user_online_status"
	bit := 0
	if online {
		bit = 1
	}
	// 设置用户的位
	return config.RDB.SetBit(config.CTX, key, uid_int64, bit).Err()
}

func GetUserOnline(uid string) (bool, error) {
	uid_int, err := strconv.Atoi(uid)
	if err != nil {
		return false, fmt.Errorf("redis_utils Format transform failure: %v", err)
	}
	uid_int64 := int64(uid_int)

	key := "user_online_status"
	status, err := config.RDB.GetBit(config.CTX, key, uid_int64).Result()
	if err != nil {
		return false, fmt.Errorf("redis_utils GetUserOnline GetBit: %v", err)
	}
	return status == 1, nil
}
