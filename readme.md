# Gomicrokit
Usually used package for microservices.   
Get dependencies:
```
export GO111MODULE=on
go get ./...
```

# Packages

### `log`
A leveled, rotated, fast, structured logger.  
Base on `go.uber.org/zap` and `github.com/natefinch/lumberjack`.
### `maths`
### `kafka`
### `websocket`

# Docker
```docker build --tag=gomicrokit --file=./Dockerfile .```  
```docker run -dit --name=gomicrokit --restart=no --network=host gomicrokit```  
