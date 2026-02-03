package jwt

import (
	"errors"
	"fmt"
	"web_websocket/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenString string, secret string) (*domain.JWTClaims, error) {
	if tokenString == "" {
		return nil, domain.ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("签名算法不符合预期: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	// jwt.Parse 返回错误：通常表示 token 格式/签名不合法；如果是过期则单独返回“已过期”错误
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrInvalidToken
	}

	// 解析过程无 error，但库仍可能标记 token 为无效（例如签名校验失败等）
	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	// 提取 Claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	jwtClaims := &domain.JWTClaims{
		UserID:   getStringClaim(claims, "user_id"),
		Username: getStringClaim(claims, "username"),
		RoleName: getStringClaim(claims, "role_name"),
		Exp:      int64(getFloatClaim(claims, "exp")),
	}

	// 验证过期时间
	if err := jwtClaims.Valid(); err != nil {
		return nil, err
	}

	return jwtClaims, nil
}

// 辅助函数：安全提取 string 类型 claim
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

// 辅助函数：安全提取 float64 类型 claim
func getFloatClaim(claims jwt.MapClaims, key string) float64 {
	if val, ok := claims[key].(float64); ok {
		return val
	}
	return 0
}
