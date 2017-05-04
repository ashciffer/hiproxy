package midwares

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	TAOBAO = "taobao"
	ALIBABA = "alibaba"
)

func Sortedstr(sets *url.Values, sep1 string, sep2 string, skip string) string {
	mk := make([]string, len(*sets))
	i := 0
	for k, _ := range *sets {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	s := []string{}

	for _, k := range mk {
		if k != skip {
			for _, v := range (*sets)[k] {
				s = append(s, k + sep2 + v)
			}
		}
	}
	return strings.Join(s, sep1)
}

func query(r *http.Request, key string) string {
	if values, ok := r.URL.Query()[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

//生成Sign
func CreateSign(p *url.Values, _type, secret string, auth_message interface{}) string {
	switch _type {
	case TAOBAO:
		return TaoBaoSign(p, secret)
	case ALIBABA:
		return AlibabaSign(p, secret, auth_message)
	}
	return ""
}
