package userlogin

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func GetToken(username string, id int64) (string, error) {
	// 定义密钥，可以自行设置一个安全的密钥
	secretKey := []byte("your-secret-key")

	// 创建一个新的 token
	token := jwt.New(jwt.SigningMethodHS256)

	// 设置 claims
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["userid"] = id
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // 设置 token 过期时间，例如一天后

	// 使用密钥进行签名
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func CheckToken(tokenString string) (*jwt.Token, error) {
	// 定义密钥，与生成 token 时使用的密钥保持一致
	secretKey := []byte("your-secret-key")

	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
