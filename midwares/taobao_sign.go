package midwares

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"

	"time"

	"net/url"

	"git.ishopex.cn/teegon/hiproxy/lib"
	"github.com/gin-gonic/gin"
)

func TaobaoSign() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CheckSign(c.Writer, c.Request) {
			c.Abort()
			return
		}
	}
}

func CheckSign(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	secret := r.PostFormValue("secret")
	token := r.PostFormValue("token")
	timestamp := r.PostFormValue("timestamp")
	ts, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		w.Write([]byte(lib.Errors["001"].String()))
		return false
	}
	if time.Now().Sub(ts) > time.Duration(60)*time.Second {
		w.Write([]byte(lib.Errors.Get("001", "timeout").String()))
		return false
	}

	u := r.PostForm
	u.Del("token")
	return token == Sign(&u, secret)
}

func Sign(form *url.Values, secret string) string {
	signstr := secret + Sortedstr(form, "", "", "sign") + secret
	h := md5.New()
	io.WriteString(h, signstr)
	return fmt.Sprintf("%032X", h.Sum(nil))
}
