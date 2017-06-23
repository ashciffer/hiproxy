package main

import (
	"encoding/json"
	"net/url"
	"testing"

	"git.ishopex.cn/teegon/hiproxy/lib"
	"gopkg.in/mgo.v2/bson"
)

func TestGetAuth(t *testing.T) {
	m := bson.M{"to_node_id": "1559155735", "status": "true", "from_node_id": "130732343"}
	mb, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	param := url.Values{}
	param.Add("method", "matrix.get.pollinfo2")
	param.Add("params", string(mb))
	t.Logf("params :%s \n", string(mb))
	b, err := lib.Request("http://127.0.0.1", "POST", param.Encode())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
	var shops []interface{}
	err = json.Unmarshal(b, &shops)
	if err != nil {
		t.Log(shops)
	}

}
