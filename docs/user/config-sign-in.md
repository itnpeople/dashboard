# Sign-in Configuration

## Introduction

* strategy
  * cookie : managed login true/false cookie
  * local : using JWT mechanism, access-token and refresh-token issuing local,  access-key & refresh-key properties required

* supported secret
  * static-user : compare static username/password
  * static-token : compare static token string
  * basic-auth : compare kubernetes's basic-auth secret ( username, password )
  * opaque : compare kubernetes's opaque secret (key, value)

* login schema
  * user : username, password
  * token : string

## How to apply
  * apply the feature as a startup parameter.

```
auth=strategy=<cookie/local>,secret=<supported secret>,<prop1=value1>,<prop2=value2>
```

* example

```
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: kore-board
    kore.board: backend
  name: backend
  namespace: kore
spec:
...
    spec:
      containers:
        - name: backend
          image: ghcr.io/kore3lab/kore-board.backend:latest
          args:
            - --metrics-scraper-url=http://metrics-scraper:8000
            - --log-level=info
            - --auth=strategy=cookie,secret=static-token,token=kore3lab
```

### static-token

* define static token string

```
spec:
  containers:
    - name: backend
      image: ghcr.io/kore3lab/kore-board.backend:latest
      args:
        - --auth=strategy=<cookie/local>,secret=static-token,token=<token-string>
```


### static-user

* define static username, password

```
spec:
  containers:
    - name: backend
      image: ghcr.io/kore3lab/kore-board.backend:latest
      args:
        - --auth=strategy=<cookie/local>,secret=static-user,username=<username>,password=<password>
```

### basic-auth

* create a 'basic-auth' secret
```
# basic-auth

$ kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: secret-basic-auth
type: kubernetes.io/basic-auth
data:
  username: admin
  password: t0p-Secret
EOF
```

* using volumn mount

```
spec:
  containers:
    - name: backend
      image: ghcr.io/kore3lab/kore-board.backend:latest
      args:
        - --auth=strategy=<cookie/local>,secret=basic-auth,location=/var/user
        ...
      volumeMounts:
      - name: user-vol
        mountPath: "/var/user"
    volumes:
    - name: user-vol
      secret:
        secretName: secret-basic-auth
```

###  opaque

```
$ kubectl create secret generic secret-auth --from-literal=admin=t0p-secret
```

* using volumn mount

```
spec:
  containers:
    - name: backend
      image: ghcr.io/kore3lab/kore-board.backend:latest
      args:
        - --auth=strategy=<cookie/local>,secret=opaque,location=/var/user
        ...
      volumeMounts:
      - name: user-vol
        mountPath: "/var/user"
    volumes:
    - name: user-vol
      secret:
        secretName: secret-auth
```

* input key and value string in your browser login page


