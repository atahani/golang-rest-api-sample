## REST API Sample Write in Go

### project features

* using [echo](https://labstack.com/echo) as high performance framework
* implement custom Error Handling 
* implement OAuth Authentication mechanism
* using JWT as Token via [jwt-go package](github.com/dgrijalva/jwt-go)
* implement Role base authorization
* wtire unit test for API endpoint and middlewares
* using [glide](https://glide.sh) as package manager

### project dependencies
1. install Golang run and tested in Go 1.6 
2. install [Glide](https://github.com/Masterminds/glide) as package manager
3. have MongoDB instance in your localhost for store data


### how to use from this sample project
##### clone the repository
```
git clone https://github.com/atahani/golang-rest-api-sample.git
cd golang-rest-api-sample
```

##### install dependencies via glide
```
glide install
```

##### build, serve or test using make

```
make build
make serve
make test
```

