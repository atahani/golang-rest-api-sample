## REST API Sample Write in Go

### Project Features

* Using [echo](https://labstack.com/echo) as high performance framework
* Implement custom Error Handling
* Implement OAuth Authentication mechanism
* Using JWT as Token via [jwt-go package](https://github.com/dgrijalva/jwt-go)
* Implement Role base authorization
* Write unit test for API endpoint and middlewares
* Using [glide](https://glide.sh) as package manager

### Project Dependencies
1. Install Golang (tested with Go 1.6)
2. Install [Glide](https://github.com/Masterminds/glide) as package manager
3. Install and run MongoDB service on your localhost for storing data


### How to use from this sample project
##### Clone the repository
```
git clone https://github.com/atahani/golang-rest-api-sample.git
cd golang-rest-api-sample
```

##### Install dependencies via glide
```
glide install
```

##### Build, serve or test using make

```
make build
make serve
make test
```

