package httpx

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func Post(url string, body interface{}, out interface{}) error {
	jsonStr, err := json.Marshal(body)
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: 3 * time.Second, // todo 把不同接口的限制超时时间都定义在一块
	}
	contentType := "application/json"
	resp, err := client.Post(url, contentType, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(respStr, out)
}

func PostForm(url string, value url.Values, out interface{}) error {
	client := http.Client{
		Timeout: 5 * time.Second, // todo 把不同接口的限制超时时间都定义在一块
	}

	resp, err := client.PostForm(url, value)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respData, out)
}
