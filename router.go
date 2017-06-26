package main

import (
	controllers "git.ishopex.cn/teegon/hiproxy/controllers"
	"git.ishopex.cn/teegon/hiproxy/midwares"
	"github.com/gin-gonic/gin"
)

func router(c *gin.Engine) {

	tb := c.Group("/service", midwares.CheckProxySign())
	{
		tb.POST("/addappinfo", hp.AddAppInfo)
		tb.POST("/addshopinfo", hp.ReloadShopInfo)
	}
	//

	na := c.Group("/node_authorization")
	{
		naRoutes := &controllers.HiProxy{}
		na.POST("add/", naRoutes.ReloadShopInfo)
	}

	// c.GET("/hiproxy", hp.ReverseFromT2P())
	c.POST("/hiproxy", hp.ReverseFromT2P())
	c.POST("/hiproxy/test", hp.TestProxy())
}
