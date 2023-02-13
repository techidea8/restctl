package core

import (
	"errors"
	"linker/utils"
	"net/http"
)

var DefaultTokenManager TokenManager = TokenManager{
	jwtSecret: "linker@tudance",
}

type TokenManager struct {
	jwtSecret string
}

func NewTokenManager(jwtSecret string) *TokenManager {
	return &TokenManager{
		jwtSecret: jwtSecret,
	}
}

// 注册
func (mgr *TokenManager) ParseToken(token string) (result map[string]interface{}, err error) {
	return utils.ParseToken(token, mgr.jwtSecret)
}

func (mgr *TokenManager) GenerateToken(values map[string]interface{}) (string, error) {
	return utils.GenerateToken(values, mgr.jwtSecret)
}

func GenerateToken(values map[string]interface{}) (string, error) {
	return DefaultTokenManager.GenerateToken(values)
}
func ParseToken(in interface{}) (result map[string]interface{}, err error) {
	if token, ok := in.(string); ok {
		return DefaultTokenManager.ParseToken(token)
	} else if req, ok := in.(*http.Request); ok {
		token := GetTokenStringFromRequest(req)
		return DefaultTokenManager.ParseToken(token)
	} else {
		return nil, errors.New("不支持的数据类型")
	}

}
