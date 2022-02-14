## http client包
***
基本使用方式

```go
u := "http://127.0.0.1:3001/api/v1/test"

res, err := gozzle.Post(u).
	//Timeout(5).
	//Debug(func(response *gozzle.Response) {}).
	//Trace(span).
	JSON(&params)

if err != nil {
	log.Println(err)
}else {
	res.DecodeJSON(&respon)
}

```

自定義
```go
t := http.DefaultTransport.(*http.Transport)
//t.MaxIdleConns = 100
t.MaxConnsPerHost = 100 //最大連線池數
//t.MaxIdleConnsPerHost = 1

res, err := gozzle.Post(u).
    Transport(t).
    Debug(func(response *gozzle.Response) {}).
    JSON(&account)
```