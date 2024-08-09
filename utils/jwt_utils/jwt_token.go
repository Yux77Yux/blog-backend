package jwt_utils

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/yux77yux/blog-backend/utils/log_utils"
	"github.com/yux77yux/blog-backend/utils/redis_utils"
)

func GenerateJWT(uid string) (string, error) {
	var timeatamp int64
	const maxRetries = 8
	const length = 32
	signingKey := make([]byte, length)
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	for i := 0; i < maxRetries; i++ {
		// 生成随机字符串
		bytes := make([]byte, length)
		_, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}
		timeatamp = time.Now().UnixNano()

		if isExists, err := redis_utils.IsJtiBlacklisted(fmt.Sprintf("%d", timeatamp)); err != nil {
			_ = isExists
			log_utils.Logger.Printf("Error: GenerateJWT: %v", err)
		} else if isExists {
			if i < maxRetries-1 {
				continue
			}
			return "", fmt.Errorf("failed to generate JWT after %d retries", maxRetries)
		}

		break
	}

	// Set claims
	claims := jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(time.Hour * 48).Unix(),
		"iat": time.Now().Unix(),
		"jti": fmt.Sprintf("%d", timeatamp),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	isExist, err := redis_utils.ExistsSigningKeyInRedix(uid)
	if err != nil {
		fmt.Println("Error: GenerateJWT ExistsSigningKeyInRedix : ", err)
		log_utils.Logger.Printf("Error: GenerateJWT ExistsSigningKeyInRedix : %v", err)
	}

	if isExist {
		signingKey, err = redis_utils.RetrieveSigningKeyInRedis(uid)
		if err != nil {
			log_utils.Logger.Printf("Error: GenerateJWT RetrieveSigningKeyInRedis : %v", err)
		}
	} else {
		_, err := rand.Read(signingKey)
		if err != nil {
			return "", fmt.Errorf("GenerateJWT rand.Read: %v", err)
		}

		go func() {
			if err := redis_utils.StoreSigningKeyInRedis(uid, signingKey); err != nil {
				log_utils.Logger.Printf("Error: GenerateJWT StoreSigningKeyInRedis: %v", err)
			}
		}()
	}

	// Sign the token with our secret
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		log_utils.Logger.Printf("Error: GenerateJWT SignedString : %v", err)
		return "", fmt.Errorf("GenerateJWT SignedString : %v", err)
	}

	return tokenString, nil
}

func ParseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证 token 的签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 从 JWT 的 claims 中提取 UID
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid claims")
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			return nil, fmt.Errorf("sub claim is missing or not a string")
		}

		// 从 Redis 获取签名密钥
		signingKey, err := redis_utils.RetrieveSigningKeyInRedis(sub)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve signing key: %v", err)
		}

		return signingKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("ParseJWT Parse: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("ParseJWT Invalid token")
	}

	// Validate the token's claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check expiration
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, fmt.Errorf("token is expired")
			}
		} else {
			// Handle the case where "exp" is missing or not a float64
			return nil, fmt.Errorf("missing or invalid exp claim")
		}

		// Optionally, check JWT ID
		if jti, ok := claims["jti"].(string); ok {
			// Perform custom validation for jti if needed
			_ = jti // Placeholder for actual validation
		} else {
			// Handle the case where "exp" is missing or not a float64
			return nil, fmt.Errorf("missing or invalid jti claim")
		}

	} else {
		return nil, fmt.Errorf("invalid token claims")
	}

	return token, nil
}
