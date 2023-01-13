package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
)

/*
* auth-info USE-CASES
strategy=cookie,secret=static-token,token=kore3lab
strategy=cookie,secret=static-user,username=admin,password=kore3lab
strategy=cookie,secret=basic-user,location=/var/tmp"
strategy=cookie,secret=opaque,location=/var/tmp"
strategy=local,accessKey=whdmstkddk,refreshKey=hsthvmxm,secret=static-token,token=kore3lab
strategy=local,accessKey=whdmstkddk,refreshKey=hsthvmxm,secret=static-user,username=admin,password=kore3lab
strategy=local,accessKey=whdmstkddk,refreshKey=hsthvmxm,secret=basic-user,location=/var/tmp"
strategy=local,accessKey=whdmstkddk,refreshKey=hsthvmxm,secret=opaque,location=/var/tmp"
*/
const (
	Realm                      = "Kore-Board"
	StrategyCookie    Strategy = "cookie"
	StrategyLocal     Strategy = "local"
	SecretStaticUser  Secret   = "static-user"
	SecretBasicUser   Secret   = "basic-user"
	SecretStaticToken Secret   = "static-token"
	SecretOpaque      Secret   = "opaque"
)

type Strategy string
type Secret string

type AuthenticatorOptions struct {
	Strategy   Strategy `json:"strategy"`
	Secret     Secret   `json:"secret"`
	AccessKey  string   `json:"accessKey"`
	RefreshKey string   `json:"refreshKey"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Token      string   `json:"token"`
	Location   string   `json:"location"`
}

// auth scheme (user, token)
func (me *AuthenticatorOptions) GetSchema() string {

	schema := "user"
	if me.Secret == "" {
		schema = ""
	} else if strings.Contains(string(me.Secret), "token") {
		schema = "token"
	}

	return schema

}

type Authenticator struct {
	Realm          string
	HandlerFunc    func() gin.HandlerFunc
	Validate       func(map[string]string) error
	LoginHandler   func(body map[string]string) (interface{}, error)
	RefreshHandler func(body map[string]string) (interface{}, error)
	LogoutHandler  func(body map[string]string)
	Options        *AuthenticatorOptions
}

type SecretProvider func(user, realm string) string
type ValidateFunc func(map[string]string) error
