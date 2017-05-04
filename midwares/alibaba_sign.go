package midwares

import (
	"crypto/sha1"
	"crypto/hmac"
	"fmt"
	"strings"
	"net/url"
)

func AlibabaSign(form *url.Values, secret string, auth_message interface{}) string {
	uri := form.Get("method") + "/" + auth_message.(map[string]interface{})["from_api_key"].(string)
	form.Del("sign_time")
	form.Del("method")
	form.Del("node_id")
	form.Del("secret")
	form.Del("sign_method")
	form.Del("secret")
	form.Del("session")

	signstr := uri + Sortedstr(form, "", "", "sign")
	fmt.Println(signstr)

	//hmac ,use sha1
	key := []byte(secret)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(signstr))
	return strings.ToUpper(fmt.Sprintf("%x", mac.Sum(nil)))
}


func AddAlibabaSystemParams(from *url.Values, method string, auth_message interface{}) {
	from.Add("method", method)
	from.Add("sellerMemberId", auth_message.(map[string]interface{})["from_auth_unikey"].(string))
	from.Add("access_token", auth_message.(map[string]interface{})["from_auth_code"].(string))
}