package auth

import (
	"strings"
)

/*
*
auth-info USE-CASES

	{"strategy":"cookie"}
	{"strategy":"cookie",	"secret": {"type": "static-user",			"username": "admin", "password": "kore3lab"} }
	{"strategy":"cookie",	"secret": {"type": "static-token",			"token": "kore3lab"} }
	{"strategy":"cookie",	"secret": {"type": "basic-auth",			"dir": "/var/tmp"} }
	{"strategy":"cookie",	"secret": {"type": "service-account-token"} }
	{"strategy":"local",	"key": {"access": "whdmstkddk", "refresh":"hsthvmxm"} }
	{"strategy":"local",	"key": {"access": "whdmstkddk", "refresh":"hsthvmxm"},	"secret": {"type": "static-user",			"username": "admin", "password": "kore3lab"} }
	{"strategy":"local",	"key": {"access": "whdmstkddk", "refresh":"hsthvmxm"},	"secret": {"type": "static-token"	,		"token": "kore3lab"} }
	{"strategy":"local",	"key": {"access": "whdmstkddk", "refresh":"hsthvmxm"},	"secret": {"type": "basic-auth",			"dir": "/var/tmp"} }
	{"strategy":"local",	"key": {"access": "whdmstkddk", "refresh":"hsthvmxm"},	"secret": {"type": "service-account-token"} }
*/
const (
	Realm                     = "Kore-Board"
	StrategyCookie            = "cookie"
	StrategyLocal             = "local"
	SecretStaticUser          = "static-user"
	SecretBasicAuth           = "basic-auth"
	SecretStaticToken         = "static-token"
	SecretServiceAccountToken = "service-account-token"
)

type AuthenticatorOptions struct {
	Strategy   string            `json:"strategy"` // (file,configmap)
	Secret     string            `json:"secret"`
	AccessKey  string            `json:"accessKey"`
	RefreshKey string            `json:"refreshKey"`
	Data       map[string]string `json:"data"`
}

// auth scheme (user, token)
func (me *AuthenticatorOptions) GetSchema() string {

	schema := "user"
	if me.Secret == "" {
		schema = ""
	} else if strings.Contains(me.Secret, "token") {
		schema = "token"
	}

	return schema

}

type SecretProvider func(user, realm string) string
type ValidateFunc func(map[string]string) error
