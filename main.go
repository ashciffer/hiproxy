package main

import (
	"time"

	"os"

	"fmt"

	"git.ishopex.cn/teegon/hiproxy/controllers"
	. "git.ishopex.cn/teegon/hiproxy/models"
	"github.com/astaxie/beego"
	"github.com/gin-gonic/gin"
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
	backendurl := beego.AppConfig.String("app::backendurl")
	hp.TeegonSecret = beego.AppConfig.String("app::secret")
	dns := beego.AppConfig.String("app::dns")
	err := hp.Init(backendurl, dns)
	if err != nil {
		fmt.Printf("hp init :%s", err)
		os.Exit(-1)
		return
	}
	r.Use(CustomLog(), gin.Recovery())
	router(r)
	r.Run(":3000")
}
