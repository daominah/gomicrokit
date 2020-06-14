# Gomicrokit
Often used packages for developing microservices

## Packages

### `auth`
* __genrsa__: generate a RSA key pair as PEM files (idRsa, idRsaPub)
* __jwt__: easy to use JSON web token,
  depend on [dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go)  
  `CreateAuthToken(authInfo interface{}) (jwtToken string)`  
  `CheckAuthToken(jwtToken string, outPointer interface{}) error`  
* __password__:  
`HashPassword(plain string) (hashed string)`  
`CheckHashPassword(hashed string, plain string) bool`

### `gofast`
Often used functions. Ex: cron job, find index in slice, UUID, ..

### `httpsvr`
Http server supports http method, url params, logging, metric.  
API is similar to standard http ServeMux HandleFunc.  
Depend on [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)

### `kafka`
An easy to use, pure go [Kafka](https://kafka.apache.org/) client.  
Depend on [Shopify/sarama](https://github.com/Shopify/sarama)

### `log`
A leveled, rotated (by time and file size) logger.  
Depend on [go.uber.org/zap](https://github.com/uber-go/zap)
and [natefinch/lumberjack](https://github.com/natefinch/lumberjack)

### `metric`
Package metric is used for observing request count and duration.  
It use an order statistic tree to store durations, so it can calculate 
percentiles very fast.

### `textproc`
Extracting information from text and html

### `websocket`
An easy-to-use websocket client and server.  
Depend on [gorilla/websocket](https://github.com/gorilla/websocket)

## Example usages
Directory `a_examples` contains executables as example usage of 
other packages in this project
