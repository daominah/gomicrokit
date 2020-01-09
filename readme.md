# Gomicrokit
Often used package for microservices


### Packages

##### `_examples`
Executables as example usage of packages.

##### `auth`
Packages: genrsa, jwt, password  
Depend on [dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go)

##### `gofast`, `maths`
Often used functions. Ex: cron job, copy similar struct,
gen uuid, ..

##### `httpsvr`
Http server supports http method, url variables and
 logs all pairs of request/response.  
 Depend on [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)

##### `kafka`
An easy-to-use, pure go kafka client.  
Depend on [Shopify/sarama](https://github.com/Shopify/sarama)

##### `log`
A leveled, rotated by time or file size logger.  
Depend on [go.uber.org/zap](https://github.com/uber-go/zap)
and [natefinch/lumberjack](https://github.com/natefinch/lumberjack)

##### `textproc`:
Extracting information from text and html

##### `socketio`
It is recommend to use websocket (standardized by IETF) instead.  
Modified from [graarh/golang-socketio](https://github.com/graarh/golang-socketio)

##### `websocket`
An easy-to-use websocket client and server.  
Depend on [gorilla/websocket](https://github.com/gorilla/websocket)


### Docker
```docker build --tag=gomicrokit --file=./Dockerfile .```  
```docker run -dit --name=gomicrokit --restart=no --network=host gomicrokit```  
