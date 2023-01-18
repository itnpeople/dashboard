package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

/*
*
<Authenticator>

	createCookieAuthenticator
	LocalAuthenticator
	BasicAuthAuthenticator (미사용)
*/

func NewAuthenticator(opts string) (*Authenticator, error) {

	var options *AuthenticatorOptions
	var err error

	//unmarshal options
	if options, err = getAuthenticatorOptions(opts); err != nil {
		return nil, err
	}

	// carete a authentocator
	validatorFn, err := getValidateFunc(options)
	if err != nil {
		return nil, err
	}

	authenticator := &Authenticator{AuthenticatorOptions: *options}

	if options.Strategy == StrategyCookie {

		// create a cookie-authenticator
		authenticator = &Authenticator{
			Validate: validatorFn,
			HandlerFunc: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Next()
				}
			},
		}
	} else if options.Strategy == StrategyLocal {

		// create a local-authenticator
		authenticator = &Authenticator{
			Validate: validatorFn,
			HandlerFunc: func() gin.HandlerFunc {

				return func(c *gin.Context) {

					postfix, _ := c.Cookie("auth.strategy")
					accessToken, err := c.Cookie(fmt.Sprintf("auth._token.%s", postfix))
					if err != nil {
						log.Warnf("prasing token (%s) failed  (cause=%s)", accessToken, err.Error())
						c.AbortWithStatus(http.StatusUnauthorized)
						return
					}
					if expired, err := validateSessionToken(options.AccessKey, accessToken); err != nil {
						log.Warnf("validate token (%s) failed  (cause=%s, expired=%v)", accessToken, err.Error(), expired)
						c.AbortWithStatus(http.StatusUnauthorized)
						return
					} else {
						if expired {
							log.Warnf("expired=%s", err.Error())
							c.AbortWithStatus(http.StatusUnauthorized)
							return
						}
					}
					c.Next()
				}
			},
			//login, refresh, logout callback
			LoginHandler: func(params map[string]string) (interface{}, error) {
				return newJWTToken(options.AccessKey, options.RefreshKey)
			},
			RefreshHandler: func(params map[string]string) (interface{}, error) {
				// validating refresh-token
				if expired, err := validateSessionToken(options.RefreshKey, params["refreshToken"]); err != nil {
					return nil, fmt.Errorf("invalid refresh token (cause=%s)", err.Error())
				} else if expired {
					return nil, errors.New("refresh token expired")
				} else {
					// new access, refresh token
					return newJWTToken(options.RefreshKey, options.RefreshKey)
				}
			},
		}

	} else {
		if options.Strategy != "" {
			return nil, fmt.Errorf("not supported '%s' strategy yet", options.Strategy)
		} else {
			// create a dummy-authenticator
			authenticator = &Authenticator{
				Validate: func(map[string]string) error {
					return nil
				},
				HandlerFunc: func() gin.HandlerFunc {
					return func(c *gin.Context) {
						c.Next()
					}
				},
			}
		}

	}

	return authenticator, nil

}

func getAuthenticatorOptions(opts string) (*AuthenticatorOptions, error) {

	options := &AuthenticatorOptions{}

	//parsing
	for _, e := range strings.Split(opts, ",") {
		parts := strings.Split(e, "=")
		if parts[0] == "strategy" {
			options.Strategy = Strategy(parts[1])
		} else if parts[0] == "secret" {
			options.Secret = Secret(parts[1])
		} else if parts[0] == "access-key" {
			options.AccessKey = parts[1]
		} else if parts[0] == "refresh-key" {
			options.RefreshKey = parts[1]
		} else if parts[0] == "username" {
			options.Username = parts[1]
		} else if parts[0] == "password" {
			options.Password = parts[1]
		} else if parts[0] == "token" {
			options.Token = parts[1]
		} else if parts[0] == "location" {
			options.Location = parts[1]
		}
	}

	if options.Strategy == StrategyLocal {
		if options.AccessKey == "" {
			return options, fmt.Errorf("access-key is mandatory on '%s' strategy", options.Strategy)
		}
		if options.RefreshKey == "" {
			return options, fmt.Errorf("refresh-key is mandatory on '%s' strategy", options.Strategy)
		}
	}

	return options, nil
}

func newJWTToken(accessSecret string, refreshSecret string) (map[string]string, error) {

	token, err := generateSessionToken(accessSecret, 60*15)
	if err != nil {
		return nil, errors.New("can't generated a access-token")
	}
	refreshToken, err := generateSessionToken(refreshSecret, 60*60*24*7)
	if err != nil {
		return nil, errors.New("can't genrated a refresh-token")
	}
	return map[string]string{
		"token":        token,
		"refreshToken": refreshToken,
	}, nil

}

func createBasicAuthAuthenticator(filename string, validateFunc ValidateFunc) *Authenticator {

	h := &Authenticator{}
	h.Validate = validateFunc
	h.HandlerFunc = func() gin.HandlerFunc {

		return func(c *gin.Context) {
			// Get the Basic Authentication credentials
			user, password, ok := c.Request.BasicAuth()
			if ok {
				err := validateFunc(map[string]string{"username": user, "password": password})
				ok = (err == nil)
			}
			if !ok {
				c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Next()
		}
	}

	return h

}

func getValidateFunc(options *AuthenticatorOptions) (ValidateFunc, error) {

	var secret SecretProvider
	var err error

	// choice provider
	ty := options.Secret
	if ty == SecretBasicUser {
		if secret, err = getBasicUserSecretProvider(options); err != nil {
			return nil, err
		}
	} else if ty == SecretStaticUser {
		if secret, err = getStaticUserSecretProvider(options); err != nil {
			return nil, err
		}
	} else if ty == SecretStaticToken {
		if secret, err = getStaticTokenSecretProvider(options); err != nil {
			return nil, err
		}
	} else if ty == SecretOpaque {
		if secret, err = getOpaqueSecretProvider(options); err != nil {
			return nil, err
		}
	} else if ty == "" {
	} else {
		return nil, fmt.Errorf("cannot found '%s' secret provider", ty)
	}

	schema := options.GetSchema()
	if schema == "user" {
		//username, password
		return func(params map[string]string) error {
			if params["username"] == "" {
				return errors.New("username is empty")
			}
			if secret(params["username"], Realm) != params["password"] {
				return errors.New("invalid password")
			} else {
				return nil
			}
		}, nil

	} else if schema == "token" {
		//token
		return func(params map[string]string) error {
			if params["token"] == "" {
				return errors.New("token is empty")
			}
			if secret(params["token"], Realm) != params["token"] {
				return errors.New("invalid token")
			} else {
				return nil
			}
		}, nil

	} else {
		return func(params map[string]string) error {
			return nil
		}, nil

	}

}
