# Backend

## Introduction

* Backend rest-api for Kore-board
* Written in Golang (version 1.15)
* Frameworks
  * [gin-gonic](https://github.com/gin-gonic/gin)

* References
  * https://github.com/kubernetes/client-go
  * https://github.com/kubernetes/api/blob/master/core/v1/types.go
  * https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/meta.go
  * https://github.com/kubernetes/client-go/blob/master/listers/core/v1 



## Tsting

### Environment (in-cluster)


```
# token과 certificate 파일을 만들기 위해 serviceaccount(incluster-sa) 과 secret(incluster-sa-token) 생성 - 1회 실행

$ kubectl create serviceaccount incluster-sa -n kube-public

$ kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: incluster-sa-token
  namespace: kube-public
  annotations:
    kubernetes.io/service-account.name: incluster-sa
type: kubernetes.io/service-account-token
EOF

$ kubectl create clusterrolebinding incluster-binding --clusterrole=cluster-admin --serviceaccount=kube-public:incluster-sa


# 테스트용 configmap 생성 - 1회 실행

$ kubectl create configmap kore-board-kubeconfig -n default --from-file=config=${HOME}/.kube/config


# 리부팅되어 `/var/run/secrets/kubernetes.io/serviceaccount` 디렉토리가 없는 경우 실행

$ sudo mkdir -p /var/run/secrets/kubernetes.io/serviceaccount
$ kubectl get secret incluster-sa-token -n kube-public -o jsonpath='{.data.ca\.crt}' | base64 --decode | sudo tee /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
$ kubectl get secret incluster-sa-token -n kube-public -o jsonpath='{.data.token}' | base64 --decode | sudo tee /var/run/secrets/kubernetes.io/serviceaccount/token

# 환경변수 `KUBERNETES_SERVICE_HOST`, `KUBERNETES_SERVICE_PORT` 없는 경우 실행

$ CLUSTER="$(kubectl config view --raw=true -o jsonpath="{.contexts[?(@.name==\"$(kubectl config current-context)\")].context.cluster}")"
$ export KUBERNETES_SERVICE_HOST=$(kubectl config view --raw=true -o jsonpath="{.clusters[?(@.name==\"$CLUSTER\")].cluster.server}" |  awk -F/ '{print $3}' |  awk -F: '{print $1}')
$ export KUBERNETES_SERVICE_PORT=$(kubectl config view --raw=true -o jsonpath="{.clusters[?(@.name==\"$CLUSTER\")].cluster.server}" |  awk -F/ '{print $3}' |  awk -F: '{print $2}')
```

### Testing

* kubeconfig

```
# all
$ go test github.com/kore3lab/dashboard/backend/pkg/kubeconfig -v

# kubeconfig (file and configmap)
$ go test github.com/kore3lab/dashboard/backend/pkg/kubeconfig -run TestKubeConfig  -v

# provider of file-strategy
$ go test github.com/kore3lab/dashboard/backend/pkg/kubeconfig -run TestFileKubeConfigProvider  -v

# provider of configmap-strategy
$ go test github.com/kore3lab/dashboard/backend/pkg/kubeconfig -run TestConfigMapKubeConfigProvider  -v

# cluster add/remove
$ go test github.com/kore3lab/dashboard/backend/pkg/kubeconfig -run TestKubeClusterAddRemove  -v


```



* authenticator

```
$ go test github.com/kore3lab/dashboard/backend/pkg/auth -v

$ go test github.com/kore3lab/dashboard/backend/pkg/auth -run TestSecretProvider -v
$ go test github.com/kore3lab/dashboard/backend/pkg/auth -run TestNewAuthenticator -v
$ go test github.com/kore3lab/dashboard/backend/pkg/auth -run TestTokenGenerateValidate -v  # generate token & validate token
$ go test github.com/kore3lab/dashboard/backend/pkg/auth -run TestTokenExpired -v           # validate token expired
```


* config
```
$ go test github.com/kore3lab/dashboard/backend/pkg/config -v
```
