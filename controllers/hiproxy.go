package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"gopkg.in/mgo.v2/bson"

	"database/sql"

	"container/list"

	"time"

	"git.ishopex.cn/teegon/hiproxy/lib"
	"git.ishopex.cn/teegon/hiproxy/midwares"
	. "git.ishopex.cn/teegon/hiproxy/models"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type HiProxy struct {
	TeegonSecret string
	BackendURL   string
	AppInfo      map[string]*ApiStat
	ShopInfo     map[string]interface{}
	rwmutex      sync.RWMutex //锁
	db           *sql.DB
}

type ApiStat struct {
	Appkey string
	NodeID string
	Apis   []string
}

//初始化代理
func (h *HiProxy) Init(backendurl, dns string) error {
	//1、初始化mysql
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return err
	}
	//2、设置最大连接和空闲连接
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	if err = db.Ping(); err != nil {
		return err
	}

	h.db = db
	h.BackendURL = backendurl
	h.ShopInfo = make(map[string]interface{})
	h.AppInfo = make(map[string]*ApiStat)

	if err = h.LoadAppInfo(); err != nil {
		return err
	}

	if err = h.LoadShopInfo(); err != nil {
		return err
	}
	return nil
}

//load appkey购买的api信息及店铺node
func (h *HiProxy) LoadAppInfo() error {
	l, err := h.QueryAppInfo("")
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			T.Error("LoadAppInfo failed,error :%s", err)
		}
	}()

	for e := l.Front(); e != nil; e = e.Next() {
		h.rwmutex.Lock()
		as := e.Value.(ApiStat)
		h.AppInfo[as.Appkey] = &as
		h.rwmutex.Unlock()
	}
	return nil
}

//加载店铺信息
func (h *HiProxy) LoadShopInfo() error {
	l, err := h.QueryShopexInfo("")
	if err != nil {
		return err
	}

	for e := l.Front(); e != nil; e = e.Next() {
		var t map[string]interface{}
		err := json.Unmarshal([]byte(e.Value.(string)), &t)
		if err != nil {
			T.Warn("load ShopInfo failed,err:%s,appinfo :%s", err, e.Value)
			continue
		}

		h.rwmutex.Lock()
		h.ShopInfo[t["from_node_id"].(string)] = &t
		h.rwmutex.Unlock()
	}
	return nil
}

//添加店铺信息
func (h *HiProxy) ReloadShopInfo(c *gin.Context) {
	node_id := c.Query("node_id")
	shopinfostr := c.Query("shop_info")
	var m map[string]interface{}

	err := json.Unmarshal([]byte(shopinfostr), &m)
	if err != nil {
		c.Abort()
		return
	}

	h.rwmutex.Lock()
	h.ShopInfo[node_id] = m
	h.rwmutex.Unlock()
}

func (h *HiProxy) ReloadAppInfo(c *gin.Context) {
	var t ApiStat
	param := url.Values{}
	appkey := c.Query("appkey")
	param.Add("appkey", appkey)

	auth, err := lib.Request(h.BackendURL, "GET", param.Encode())
	if err != nil {
		c.Abort()
		T.Error("query appinfo failed,error:%s,param:%s", err, param.Encode())
		return
	}

	err = json.Unmarshal(auth, &t)
	if err != nil {
		T.Error("query appinfo failed,error:%s,param:%s", err, param.Encode())
		c.Abort()
		return
	}

	h.rwmutex.Lock()
	h.AppInfo[appkey] = &t
	h.rwmutex.Unlock()
}

// ReverseFromT2P 从后台代理api到第三方平台
func (h *HiProxy) ReverseFromT2P() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			auth_message interface{}
			ok           bool
			err          error
			appkey       string
			apistat      *ApiStat
		)
		appkey = c.PostForm("app_key")
		method := c.PostForm("api_method")
		node_id := c.PostForm("node_id") //店铺
		//TODO from_node_id h
		//判断是否有调用权限
		if apistat, ok = h.AppInfo[appkey]; !ok {
			T.Error("appinfo :%v", h.AppInfo)
			T.Debug("apistat :%s", appkey)
			c.JSON(200, lib.Errors["002"])
			return
		}

		for _, v := range apistat.Apis {
			if v == method {
				ok = true
			}
		}

		//读取授权信息
		if ok {
			var b []byte
			if auth_message, ok = h.ShopInfo[apistat.NodeID]; !ok {
				T.Debug("can't find shopinfo ,nodeid :%s", apistat.NodeID)
			}

			//如果没有，则主动查找授权信息
			if auth_message == nil {
				auth_message, err = h.QueryNodeAuthMessage(apistat.NodeID, c.PostForm("type"), node_id)
				if err != nil {
					c.JSON(200, lib.Errors["101"])
					return
				}
				if auth_message == nil {
					T.Debug("load shop info failed,node_id:%s,type:%s", apistat.NodeID, c.Request.Form.Get("type"))
					c.JSON(200, lib.Errors["101"])
					return
				}
			}
			//验签，代理

			u := c.Request.Form
			platform_type := auth_message.(map[string]interface{})["from_type"].(string)

			//添加系统参数

			u.Add("key", auth_message.(map[string]interface{})["from_api_key"].(string))
			auth_secret := auth_message.(map[string]interface{})["from_api_secret"].(string)
			u.Add("secret", auth_secret)
			//u.Add("token", auth_message.(map[string]interface{})["from_token"].(string))
			pu := auth_message.(map[string]interface{})["to_api_url"].(string)
			// puu, err := url.Parse(pu)
			// if err != nil {
			// 	c.Writer.Write([]byte(err.Error()))
			// 	return
			// }
			h.AddPlatformParams(&u, platform_type, method, appkey)

			u.Add("sign", midwares.CreateSign(&u, platform_type, auth_secret))

			//puu.RawQuery = u.Encode()
			// c.Request.Header.Set("Content-Length", strconv.Itoa(len(u.Encode())))
			// c.Request.Body = ioutil.NopCloser(strings.NewReader(u.Encode()))
			b, err = lib.Request(pu, "POST", u.Encode())
			//	b, err = newReverseProxy(puu).ServeHTTP(c.Writer, c.Request)
			if err != nil {
				T.Error("proxy failed,error:%s", err)
				c.JSON(200, lib.Errors.Get("500", err))
			} else {
				var res map[string]interface{}
				json.Unmarshal(b, &res)
				if _, ok := res["error_response"]; ok {
					T.Error("platfrom return error resopnse ,error :%s", string(b))
					h.QueryNodeAuthMessage(apistat.NodeID, platform_type, node_id)
				} else {
					T.Info("proxy success，result:%s", string(b))
				}
			}
		} else {
			c.JSON(400, lib.Errors["100"])
		}

	}
}

func (h *HiProxy) TestProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, _ := url.Parse("http://127.0.0.1:8080/test")
		b, err := newReverseProxy(u).ServeHTTP(c.Writer, c.Request)
		if err != nil {
			T.Error("proxy failed,error:%s", err)
			return
		}
		T.Error("responese :%s", string(b))
	}
}

//查找AppInfo
func (h *HiProxy) QueryAppInfo(appkey string) (*list.List, error) {
	var rows *sql.Rows
	sql := ""
	var err error
	if len(appkey) <= 0 {
		sql = "select fd_app_key,fd_node_id,fd_api_info from t_app_proxy where fd_status=1"
	} else {
		sql = fmt.Sprintf("select fd_app_key,fd_node_id,fd_api_info from t_app_proxy where fd_app_key=\"%s\" and fd_status=1", appkey)
	}

	rows, err = h.db.Query(sql)
	if err != nil {
		T.Warn("query all AppInfo, query failed, sql:%s, error:", sql, err.Error())
		return nil, err
	}

	defer rows.Close()

	T.Info("query all AppInfo, query success, sql:%s", sql)

	all_info := list.New()
	for rows.Next() {
		var s ApiStat
		var ss string
		err := rows.Scan(&s.Appkey, &s.NodeID, &ss)
		if err != nil {
			T.Error("scan appinfo failed,error:%s,sql:%s", err, sql)
			return nil, err
		}
		err = json.Unmarshal([]byte(ss), &s.Apis)
		if err != nil {
			T.Error("full in result failed,error:%s，result,%s", err, ss)
			return nil, err
		}
		all_info.PushBack(s)
	}
	T.Info("query all AppInfo, end, node_id:%s, result num:%d", appkey, all_info.Len())
	return all_info, nil
}

//查找店铺信息
func (h *HiProxy) QueryShopexInfo(nodeid string) (*list.List, error) {

	var rows *sql.Rows
	sql := ""
	var err error
	if len(nodeid) <= 0 {
		sql = "select fd_shop_info from t_app_shop "
	} else {
		sql = fmt.Sprintf("select fd_shop_info from t_app_shop where fd_node_id=\"%s\" ", nodeid)
	}

	rows, err = h.db.Query(sql)
	if err != nil {
		T.Warn("query all shopInfo, query failed, sql:%s, error:", sql, err.Error())
		return nil, err
	}

	defer rows.Close()

	T.Info("query all shopInfo, query success, sql:%s", sql)

	all_info := list.New()
	for rows.Next() {
		var a map[string]interface{}
		var ss string
		err := rows.Scan(&ss)
		if err != nil {
			T.Error("scan shopinfo failed,error:%s,sql:%s", err, sql)
			return nil, err
		}
		err = json.Unmarshal([]byte(ss), &a)
		if err != nil {
			T.Error("full in result failed,error:%s,", err)
			return nil, err
		}
		all_info.PushBack(a)
	}
	T.Info("query all AppInfo, end, node_id:%s, result num:%d", nodeid, all_info.Len())
	return all_info, nil
}

//从服务后台获取店铺授权信息
func (h *HiProxy) QueryNodeAuthMessage(node_id, _type string, from_node_id string) (auth interface{}, err error) {
	m := bson.M{"to_node_id": node_id, "status": "true", "from_node_id": from_node_id}
	mb, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	param := url.Values{}
	param.Add("method", "matrix.get.pollinfo2")
	param.Add("params", string(mb))
	T.Debug("params :%s", string(mb))
	b, err := lib.Request(h.BackendURL, "POST", param.Encode())
	if err != nil {
		return nil, err
	}

	var shops []interface{}
	err = json.Unmarshal(b, &shops)
	if err != nil {
		return nil, err
	}

	if len(shops) < 1 {
		T.Error("shopinfo not right,len :%d,result:%s", len(shops), string(b))
		return
	}

	a := shops[0].(map[string]interface{})

	h.rwmutex.Lock()
	h.ShopInfo[node_id] = a
	h.rwmutex.Unlock()

	str, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("insert into t_app_shop(fd_node_id,fd_shop_info,fd_create_time) values('%s','%s','%s') on duplicate key update fd_shop_info='%s'", node_id, str, time.Now().Format("2006-01-02 15:04:05"), str)
	_, err = h.db.Exec(sql)
	if err != nil {
		return nil, err
	}
	return a, nil
}

//添加调用信息
func (h *HiProxy) AddAppInfo(c *gin.Context) {
	appkey := c.Query("appkey")
	node_id := c.Query("node_id")
	t := c.Query("type")
	status := c.Query("status")
	info := c.Query("apiinfo")

	sql := fmt.Sprintf("insert into t_app_proxy(fd_app_key,fd_node_id,fd_api_type,fd_status,fd_api_info) values('%s','%s','%s','%s','%s')", appkey, node_id, t, status, info)

	_, err := h.db.Exec(sql)
	if err != nil {
		h.rwmutex.Lock()
		if api, ok := h.AppInfo[appkey]; ok {
			api.Apis = append(api.Apis, info)
		} else {
			h.AppInfo[appkey] = &ApiStat{
				Appkey: appkey,
				NodeID: node_id,
				Apis:   []string{info},
			}
		}
		h.rwmutex.Unlock()
		c.Writer.Write([]byte(err.Error()))
	} else {
		c.Writer.Write([]byte("success"))
	}
}

func (h *HiProxy) AddPlatformParams(u *url.Values, platform_type, method, appkey string) {
	switch platform_type {
	case "taobao":
		midwares.AddTaobaoSystemParams(u, method, appkey)
	}

}

func toTarget(target *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("User-Agent", "HiProxy/2.2")
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		targetQuery := target.RawQuery

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
}

func newReverseProxy(target *url.URL) *lib.ReverseProxy {
	return &lib.ReverseProxy{
		Director: toTarget(target),
	}
}
