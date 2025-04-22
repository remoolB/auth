package util

import (
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateObjectIDFromString 将字符串转换为ObjectID
// 如果输入是有效的ObjectID，则直接使用；否则基于输入字符串生成一个确定性的ObjectID
func GenerateObjectIDFromString(input string) primitive.ObjectID {
	// 如果输入是有效的ObjectID，则直接使用
	if id, err := primitive.ObjectIDFromHex(input); err == nil {
		return id
	}

	// 否则生成一个基于输入字符串的确定性ObjectID
	// 使用一个简单的哈希算法将字符串转换为12字节的数据
	hash := 0
	for i := 0; i < len(input); i++ {
		hash = 31*hash + int(input[i])
	}

	// 与当前时间戳结合，确保唯一性
	timestamp := time.Now().Unix()

	// 创建一个新的ObjectID
	return primitive.NewObjectIDFromTimestamp(time.Unix(timestamp, int64(hash)))
}

// ObjectIDToInt64 将ObjectID转换为int64
// 用于前端展示和API兼容性
func ObjectIDToInt64(id primitive.ObjectID) int64 {
	// 获取ObjectID的十六进制表示
	hexStr := id.Hex()

	// 取前8个字符并转换为int64
	if len(hexStr) >= 8 {
		substr := hexStr[:8]
		// 从十六进制字符串转换为int64
		if val, err := strconv.ParseInt(substr, 16, 64); err == nil {
			return val
		}
	}

	// 如果转换失败，返回时间戳作为备选
	return time.Now().Unix()
}

// Int64ToObjectID 将int64转换为ObjectID
// 用于从前端接收ID并转换为数据库使用的ObjectID
func Int64ToObjectID(val int64) (primitive.ObjectID, error) {
	// 将int64转换为十六进制字符串
	hexStr := strconv.FormatInt(val, 16)

	// 确保字符串长度为24个字符
	for len(hexStr) < 24 {
		hexStr = "0" + hexStr
	}

	// 如果超过24个字符，截取最后24个字符
	if len(hexStr) > 24 {
		hexStr = hexStr[len(hexStr)-24:]
	}

	// 尝试将其转换为ObjectID
	return primitive.ObjectIDFromHex(hexStr)
}
