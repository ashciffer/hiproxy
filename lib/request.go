package lib

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
)

//请求
func Request(u string, method string, params string) ([]byte, error) {
	client := &http.Client{
		Transport: http.DefaultTransport,
	}
	var req *http.Request
	if method == "POST" || method == "PUT" {
		url_param_reader := bytes.NewReader([]byte(params))
		req, _ = http.NewRequest(
			method,
			u,
			url_param_reader,
		)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		temp_url, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		temp_url.RawQuery = temp_url.RawQuery + "&" + params
		req, _ = http.NewRequest(
			method,
			temp_url.String(),
			nil,
		)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
