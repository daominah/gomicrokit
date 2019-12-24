export LOG_FILE_PATH=/tmp/httpsvr.log
go run httpsvr.go 2>>$LOG_FILE_PATH &