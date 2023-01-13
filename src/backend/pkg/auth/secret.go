package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

/*
*
<SecretProvider>
	StaticUserSecretProvider	: static username, password
	StaticTokenSecretProvider	: static token
	BasicUserSecretProvider		: kubernetes basic-auth secret (username file + password file)
	OpaqueSecretProvider		: kubernetes service-account-token secret
*/

// StaticToken
func getStaticTokenSecretProvider(options *AuthenticatorOptions) (SecretProvider, error) {

	var mu sync.RWMutex

	return func(user, realm string) string {
		mu.RLock()
		defer mu.RUnlock()
		return options.Token
	}, nil
}

// StaticUser
func getStaticUserSecretProvider(options *AuthenticatorOptions) (SecretProvider, error) {

	var mu sync.RWMutex

	return func(user, realm string) string {
		mu.RLock()
		defer mu.RUnlock()
		exists := (options.Username == user)
		if !exists {
			return ""
		}
		return options.Password
	}, nil
}

// BasicUserFile
func getBasicUserSecretProvider(options *AuthenticatorOptions) (SecretProvider, error) {

	if _, err := os.Stat(options.Location); os.IsNotExist(err) {
		return nil, err
	}

	pathUsername := filepath.Join(options.Location, "username")
	pathPassword := filepath.Join(options.Location, "password")
	fileUsername, err := os.Stat(pathUsername)
	if err != nil {
		return nil, err
	}
	filePassword, err := os.Stat(pathPassword)
	if err != nil {
		return nil, err
	}

	var mu sync.RWMutex
	var Username string
	var Password string

	fnLoad := func() error {
		mu.Lock()
		if d, err := os.ReadFile(pathUsername); err == nil {
			Username = string(d)
		} else {
			Username = ""
		}
		if d, err := os.ReadFile(pathPassword); err == nil {
			Password = string(d)
		} else {
			Password = ""
		}
		mu.Unlock()
		if Username == "" || Password == "" {
			return errors.New("username or password is empty")
		} else {
			return nil
		}
	}
	if err := fnLoad(); err != nil {
		return nil, err
	}

	return func(user, realm string) string {

		// ReloadIfNeeded
		mu.Lock()
		reload := false
		info, err := os.Stat(pathPassword)
		if err != nil {
			log.Warnf("Fail to authenticate (secret=basic-user, cause=%v)", err)
		} else if fileUsername.ModTime() != info.ModTime() {
			fileUsername = info
			reload = true
		}

		if !reload {
			info, err = os.Stat(pathPassword)
			if err != nil {
				//return err
			} else if filePassword.ModTime() != info.ModTime() {
				filePassword = info
				reload = true
			}
		}
		mu.Unlock()

		if reload {
			if err = fnLoad(); err != nil {
				log.Warnf("Fail to authenticate (secret=basic-user, cause=%v)", err)
			}
		}

		mu.RLock()
		defer mu.RUnlock()
		exists := (Username == user)
		if !exists {
			return ""
		}
		return Password
	}, nil

}

// OpaqueSecre
func getOpaqueSecretProvider(options *AuthenticatorOptions) (SecretProvider, error) {

	if _, err := os.Stat(options.Location); os.IsNotExist(err) {
		return nil, err
	}

	var mu sync.RWMutex
	var modifed map[string]struct {
		modTime time.Time
		value   string
	} = make(map[string]struct {
		modTime time.Time
		value   string
	})

	fnLoad := func(key string) error {
		obj := struct {
			modTime time.Time
			value   string
		}{}
		mu.Lock()
		file := filepath.Join(options.Location, key)
		if f, err := os.Stat(file); err == nil {
			obj.modTime = f.ModTime()
			if d, err := os.ReadFile(file); err == nil {
				b, _ := base64.StdEncoding.DecodeString(string(d))
				obj.value = string(b)
			} else {
				obj.value = ""
			}
			modifed[key] = obj
		}
		mu.Unlock()
		if obj.value == "" {
			return errors.New(fmt.Sprintf("value is empty (key=%s)", key))
		} else {
			return nil
		}
	}

	return func(key, realm string) string {

		file := filepath.Join(options.Location, key)

		// ReloadIfNeeded
		mu.Lock()
		var modTime time.Time = time.Time{}
		if obj, ok := modifed[key]; ok {
			modTime = obj.modTime
		}

		reload := false
		info, err := os.Stat(file)
		if err != nil {
			log.Warnf("Fail to authenticate (secret=opaque, cause=%v)", err)
		} else if info.ModTime() != modTime {
			modTime = info.ModTime()
			reload = true
		}
		mu.Unlock()

		if reload {
			if err = fnLoad(key); err != nil {
				log.Warnf("Fail to authenticate (secret=opaque, cause=%v)", err)
			}
		}

		mu.RLock()
		defer mu.RUnlock()
		obj, exists := modifed[key]
		if !exists {
			return ""
		}
		return obj.value
	}, nil

}

// ServiceAccountToken (not used/ to be use)
func getServiceAccountTokenSecretProvider(options *AuthenticatorOptions) (SecretProvider, error) {

	return func(token, realm string) string {

		claims, err := getTokenClaims(token)
		if err != nil {
			log.Errorln(err.Error())
			return ""
		}
		ns := claims["kubernetes.io/serviceaccount/namespace"].(string)
		nm := claims["kubernetes.io/serviceaccount/secret.name"].(string)

		// secret 을 읽어올 cluster 선정

		if conf, err := rest.InClusterConfig(); err != nil {
			log.Warnf("cannot create a kubernetes api-client (cause=%s)", err)
			return ""
		} else if apiClient, err := kubernetes.NewForConfig(conf); err != nil {
			log.Warnf("cannot create a kubernetes api-client (cause=%s)", err)
			return ""
		} else {
			se, err := apiClient.CoreV1().Secrets(ns).Get(context.TODO(), nm, v1.GetOptions{})
			if err == nil {
				return string(se.Data["token"])
			} else {
				log.Warnf("cannot load token from service-account (namespace=%s,service-account=%s, cause=%s)", ns, nm, err)
				return ""
			}
		}
	}, nil
}
