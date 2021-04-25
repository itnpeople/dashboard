# :whale2: Kore Board
> Multi-cluster kubernetes dashboard

![Version](https://img.shields.io/github/release/acornsoftlab/dashboard) ![License](https://img.shields.io/github/license/acornsoftlab/dashboard) ![issues](https://img.shields.io/github/issues/acornsoftlab/dashboard) ![issues](https://img.shields.io/github/issues-closed/acornsoftlab/dashboard) ![pull requests](https://img.shields.io/github/issues-pr-closed/acornsoftlab/dashboard) ![workflow](https://img.shields.io/github/workflow/status/acornsoftlab/dashboard/Docker) ![code size](https://img.shields.io/github/languages/code-size/acornsoftlab/dashboard) 

## Screen shoot

![](https://raw.githubusercontent.com/itnpeople/dashboard/feature/auth/docs/resources/screenshoot-210421.png?token=AIRLNCQ4A6B27R6BUSWIFKTAP7ISK)

## Preparation

아래 소프트웨어 설치 및 $PATH 변수에 추가 필요

* Curl 7+
* Git 2.13.2+
* Docker 1.13.1+ ([installation manual](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/))
* Golang 1.13.9+ ([installation manual](https://golang.org/dl/))
* Node.js 12+ and npm 6+ ([installation with nvm](https://github.com/creationix/nvm#usage))
* Gulp.js 4+ ([installation manual](https://github.com/gulpjs/gulp/blob/master/docs/getting-started/1-quick-start.md))

## Getting started

* clone

```
$ git clone https://github.com/acornsoftlab/dashboard.git
$ cd dashboard
```

* `subtree`로 구성된 소스 중 사용되지 않는 파일을 제외하도록 지정하고 소스 트리를 업데이트합니다.

```
# sparse checkout 옵션 지정
$ git config core.sparsecheckout true

# 대상 파일 지정
$ cat <<EOF> .git/info/sparse-checkout
/*
!/src/app/metrics-scraper
/src/app/metrics-scraper/hack/build.sh
/src/app/metrics-scraper/pkg
/src/app/metrics-scraper/vendor
/src/app/metrics-scraper/Dockerfile
/src/app/metrics-scraper/go.*
/src/app/metrics-scraper/*.go
!/src/app/metrics-scraper/**/*_test.go
EOF

# 트리 업데이트
$ git read-tree HEAD -m -u
```


* Installation dependencies (frontend,graph)

```
# graph
$ cd ../src/app/graph
$ npm i

$ cd ../../..
$ npm i
```

* Run

```
$ npm run start
```

* Verified running

```
# frontend
$ curl http://localhost:3000/

# backend
$ curl http://localhost:3001/healthy

# metrics scraper
$ curl http://localhost:8000/api/v1
```

## Installation guide

[Installation guide](./scripts/install/README.md)


## Login token

### 조회

* Kubernetes : `kore-board` serviceaccount 의 secret

```
$ SECRET="$(kubectl get sa -n kore -l app=kore-board -o jsonpath='{.items[0].secrets[0].name}')"
$ echo "$(kubectl get secret kore-board -n kore -o jsonpath='{.data.token}' | base64 --decode)"
```

* backend 로그에서 조회

```
$ docker logs backend | grep TOKEN

$ kubectl logs $(kubectl get po -l kore.board=backend -o jsonpath='{..items[0]..metadata.name}' -n kore) -n kore | grep TOKEN
```

## Development

### npm

* Develop
```
$ npm run start                 # 실행 (backend, frontend, metrics-scraper)
$ npm run start:frontend        # frontend 실행
$ npm run start:backend         # backend 실행
$ npm run start:metrics-scraper # metrics scraper 실행
$ npm run start:graph           # graph 실행
$ npm run start:graph:backend   # 그래프 개발 실행 (backend + graph)
```

* Build
```
$ npm run build:frontend      # frontend 빌드 (using on docker build)
$ npm run build:graph         # 그래프 빌드 frontend 에 변경된 최신 그래프 적용시 사용
$ npm run run                 # frontend container 에서 nuxt 실행 (docker image entrypoint) 
```

* Containerization
  * default tag : "latest" (.npmrc 파일 참조)
  * `--kore-board:docker_image_tag=<value>` 옵션으로 latest 대신 tag 지정 가능

```
# docker build
$ npm run docker:build:frontend
$ npm run docker:build:backend
$ npm run docker:build:metrics-scraper

# docker push
$ npm run docker:push:frontend    
$ npm run docker:push:backend
$ npm run docker:push:metrics-scraper

# docker build & push
$ npm run docker:build:push:frontend    
$ npm run docker:build:push:backend
$ npm run docker:build:push:metrics-scraper

# all (frontend, backend)
$ npm run docker:build        # build
$ npm run docker:build:push   # build & push

# tag 를 옵션으로 지정하는 예
$ npm run docker:build:backend --kore-board:docker_image_tag=v0.2.0
```

### Using ports
* 3000 : front-end
* 3001 : backend (restful-api)
* 3002 : graph 개발
* 8000 : metrics scraper

## Debugging for "in-cluster" clusater enviroment 

```
# delete token, ca fils
$ sudo rm /var/run/secrets/kubernetes.io/serviceaccount/token /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
$ sudo mkdir -p /var/run/secrets/kubernetes.io/serviceaccount/

# set namespace & serviceaccount name
$ export NAMESPACE="kore"
$ export SERVICE_ACCOUNT="kore-board"

# create token, ca files
$ echo -n "$(kubectl get secret -n ${NAMESPACE} $(kubectl get sa ${SERVICE_ACCOUNT} -n ${NAMESPACE} -o jsonpath={.secrets..name} | cut -f1 -d ' ') -o jsonpath='{$.data.token}' | base64 --decode)" | sudo tee /var/run/secrets/kubernetes.io/serviceaccount/token
$ echo -n "$(kubectl config view --raw=true -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 --decode)" | sudo tee /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

# set api-server endpoint
$ export KUBERNETES_SERVICE_HOST=$(kubectl config view --raw=true -o jsonpath='{.clusters[0].cluster.server}' |  awk -F/ '{print $3}' |  awk -F: '{print $1}')
$ export KUBERNETES_SERVICE_PORT=$(kubectl config view --raw=true -o jsonpath='{.clusters[0].cluster.server}' |  awk -F/ '{print $3}' |  awk -F: '{print $2}')

# create a empty kubeconfig file
$ rm "$(pwd)/.tmp/config"
$ export KUBECONFIG="$(pwd)/.tmp/config"
```

## Modules

* [backend](./src/app/backend/README.md)
* [frontend](./src/app/frontend/README.md)
* [metrics-scraper](./src/app/metrics-scraper/README.md)
* [graph](./src/app/graph/README.md)


## Link
* https://github.com/acornsoftlab/dashboard
* https://github.com/kubernetes/client-go
* https://bootstrap-vue.org/docs/components
* https://github.com/gin-gonic/gin
