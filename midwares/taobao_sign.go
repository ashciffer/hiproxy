package midwares

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"git.ishopex.cn/matrix/kaola/lib"

	"github.com/gin-gonic/gin"
)

func CheckProxySign() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CheckSign(c.Writer, c.Request) {
			c.Abort()
			return
		}
	}
}

func CheckSign(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	token := r.PostFormValue("sign")

	if !(token == TaoBaoSign(&r.PostForm, "")) {
		w.Write([]byte(lib.Errors.Get("001", "sign error").String()))
		return false
	}
	return true
}

func TaoBaoSign(form *url.Values, secret string) string {
	signstr := secret + Sortedstr(form, "", "", "sign")
	h := md5.New()
	io.WriteString(h, signstr)
	return fmt.Sprintf("%032X", h.Sum(nil))
}

func AddTaobaoSystemParams(from *url.Values, method, app_key string) {
	from.Add("method", method)
	from.Add("app_key", app_key)
	from.Add("sgin_method", "md5")

	from.Add("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	from.Add("format", "json")
	from.Add("v", "2.0")
}
