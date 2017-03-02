package controllers

import "github.com/gin-gonic/gin"

func GetFile(c *gin.Context) {
	c.Redirect(302, c.Query("file"))
}
