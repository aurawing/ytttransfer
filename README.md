# ytttransfer
eos-go中：
responses.go 第195行改为More string
api.go 第561行增加以下方法：
```
func (api *API) Call(baseAPI string, endpoint string, body interface{}, out interface{}) error {
	return api.call(baseAPI, endpoint, body, out)
}
```

注：使用go-mod时需要修改%GOPATH%\pkg\mod\github.com\eoscanada\eos-go@v0.8.11下的源文件