package main

import (
	"time"

	"os"

	"fmt"

	"git.ishopex.cn/teegon/hiproxy/controllers"
	. "git.ishopex.cn/teegon/hiproxy/models"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
)

var hp *controllers.HiProxy

func CustomLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		params := c.Request.Form.Encode()
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.String()

		T.Info("[HiProxy] %v | %3d | %s | %13v | %s | %-7s %s\n%s",
			end.Format("2006/01/02 - 15:04:05"),
			statusCode,
			params,
			latency,
			clientIP,
			method,
			path,
			comment,
		)
	}
}

func main() {
	hp = &controllers.HiProxy{}
	T.Debug("Starting. Date:%s", time.Now().Local().String())
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	cfg, err := ini.LooseLoad("conf/app.conf", "")
	if err != nil {
		fmt.Println("hiproxy conf init failed:%s", err)
		os.Exit(-1)
	}

	backendurl := cfg.Section("app").Key("backendurl").String()
	dns := cfg.Section("app").Key("dns").String()
	err = hp.Init(backendurl, dns)
	if err != nil {
		fmt.Printf("hiprpxy init :%s \n", err)
		os.Exit(-1)
		return
	}
	r.Use(CustomLog(), gin.Recovery())
	router(r)
	r.Run(":3000")
}
