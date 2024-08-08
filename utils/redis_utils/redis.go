package redis_utils

import (
	"fmt"
	"strconv"

	"github.com/yux77yux/blog-backend/config"
	"github.com/yux77yux/blog-backend/internal/model"
)

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
		"Id":         user.Id,
		"Uid":        user.Uid,
		"Name":       user.Name,
		"Profile":    user.Profile,
		"Bio":        user.Bio,
		"Popularity": user.Popularity,
	}).Result()

	return fmt.Errorf("redis_utils StoreUserInRedis HSet: %v", err)
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

	offset, err := strconv.Atoi(uid)
	if err != nil {
		return nil, fmt.Errorf("redis_utils GetUserFromRedis Format error: %v", err)
	}

	status, err := GetUserOnline(int32(offset - 100000000))
	if err != nil {
		return nil, fmt.Errorf("redis_utils GetUserFromRedis GetUserOnline: %v", err)
	}

	// 将数据反序列化为 UserIncidental 对象
	user := &model.UserIncidental{
		Id:         int32(offset - 100000000),
		Uid:        data["Uid"],
		Name:       data["Name"],
		Profile:    data["Profile"],
		Bio:        data["Bio"],
		Status:     status,
		Popularity: float32(popularity),
	}

	return user, nil
}

func SetUserOnline(userID int32, online bool) error {
	key := "user_online_status"
	bit := 0
	if online {
		bit = 1
	}
	// 设置用户的位
	return config.RDB.SetBit(config.CTX, key, int64(userID), bit).Err()
}

func GetUserOnline(userID int32) (bool, error) {
	key := "user_online_status"
	status, err := config.RDB.GetBit(config.CTX, key, int64(userID)).Result()
	if err != nil {
		return false, fmt.Errorf("redis_utils GetUserOnline GetBit: %v", err)
	}
	return status == 1, nil
}
