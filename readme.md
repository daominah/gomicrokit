# Gomicrokit
Usually used package for microservices.   
Get dependencies:
```
export GO111MODULE=on
go get ./...
```

# Packages

### `log`
A leveled, rotated by time or file size logger.  
Base on [go.uber.org/zap]() and [github.com/natefinch/lumberjack]()

### `maths`
Often used math functions

### `kafka`
An easy-to-use, pure go kafka client base on [github.com/Shopify/sarama]()

### `websocket`
An easy-to-use websocket client and server base on [github.com/gorilla/websocket]()

### `socketio`
Modified from [github.com/graarh/golang-socketio]()

# Docker
```docker build --tag=gomicrokit --file=./Dockerfile .```  
```docker run -dit --name=gomicrokit --restart=no --network=host gomicrokit```  
