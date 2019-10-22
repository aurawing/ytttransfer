# ytttransfer
eos-go中：
responses.go 第195行改为More string
api.go 第561行增加以下方法：
```
func (api *API) Call(baseAPI string, endpoint string, body interface{}, out interface{}) error {
	return api.call(baseAPI, endpoint, body, out)
}
```