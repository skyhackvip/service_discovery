package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HttpPost(uri string, data interface{}) (string, error) {
	client := &http.Client{}
	js, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", uri, bytes.NewReader(js))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}
