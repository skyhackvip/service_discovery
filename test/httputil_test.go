package test

import (
	"encoding/json"
	"github.com/skyhackvip/service_discovery/model"
	"github.com/skyhackvip/service_discovery/pkg/httputil"
	"log"
	"testing"
)

func TestPost(t *testing.T) {
	resp, err := httputil.HttpPost("http://localhost:6666/api/fetchall", nil)

	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
	var res struct {
		Code    int                          `json:"code"`
		Message string                       `json:"message"`
		Data    map[string][]*model.Instance `json:"data"`
	}
	err = json.Unmarshal([]byte(resp), &res)
	if err != nil {
		t.Error(err)
	}
	for k, v := range res.Data {
		t.Log(k)
		for _, i := range v {
			t.Log(i.Hostname, i.Addrs)
		}
	}
}

func TestPost2(t *testing.T) {
	/*params := url.Values{}
	params.Set("env", "testenv")
	params.Set("appid", "testappid")
	params.Set("hostname", "testhostname")
	params.Set("addrs", strings.Join([]string{"testaddr1", "testaddr2"}, ","))
	params.Set("status", "1")
	*/

	params := make(map[string]interface{})
	params["env"] = "testenv"
	params["appid"] = "testappid"
	params["hostname"] = "testhostname"
	params["addrs"] = []string{"testaddr1", "testaddr2"}
	params["status"] = 1

	resp, err := httputil.HttpPost("http://localhost:6666/api/register", params)
	t.Log(resp, err)
	log.Println(resp)
	var res struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	err = json.Unmarshal([]byte(resp), &res)
	t.Log(res.Code, res.Data)

}
