package models

import "github.com/astaxie/beego/logs"

var T *logs.BeeLogger

func init() {
	T = logs.NewLogger(100)
	T.SetLogger("file", `{"filename":"logs/hiproxy.log","daily":true,"maxdays":7}`)
	T.SetLogger("console", `{"level":7}`)
	T.EnableFuncCallDepth(true)
	T.SetLogFuncCallDepth(2)
}
