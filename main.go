package main

import (
	_ "mouban/agent"
	_ "mouban/common"
	"mouban/consts"
	"mouban/controller"
	"mouban/util"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	serviceOpsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mouban_service_ops_duration",
		Help:    "Histogram of the duration of HTTP service requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "referer"})
)

func main() {

	router := gin.New()

	router.Use(handle)
	router.Use(cors)
	router.Use(logger)
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		c.Request.Referer()
		duration := time.Since(start).Seconds()
		serviceOpsHistogram.WithLabelValues(c.Request.Method, c.Request.URL.Path, c.Request.Referer()).Observe(duration)
	})

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "")
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	adminGroup := router.Group("/admin")
	{
		adminGroup.GET("/load_data", controller.LoadData)
		adminGroup.GET("/refresh_item", controller.RefreshItem)
		adminGroup.GET("/refresh_user", controller.RefreshUser)
	}

	queryGroup := router.Group("/guest")
	{
		queryGroup.GET("/check_user", controller.CheckUser)
		queryGroup.GET("/user_book", func(ctx *gin.Context) {
			controller.ListUserItem(ctx, consts.TypeBook.Code)
		})
		queryGroup.GET("/user_game", func(ctx *gin.Context) {
			controller.ListUserItem(ctx, consts.TypeGame.Code)
		})
		queryGroup.GET("/user_movie", func(ctx *gin.Context) {
			controller.ListUserItem(ctx, consts.TypeMovie.Code)
		})
		queryGroup.GET("/user_song", func(ctx *gin.Context) {
			controller.ListUserItem(ctx, consts.TypeSong.Code)
		})
	}

	panic(router.Run(":" + viper.GetString("server.port")))
}

func handle(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"msg":     "服务内部错误，请联系开发者处理",
			})
			logrus.Errorln("handle panic", r, "=>", util.GetCurrentGoroutineStack())
		}
	}()
	ctx.Next()
}

func cors(c *gin.Context) {
	cors := viper.GetString("server.cors")
	method := c.Request.Method
	origin := c.Request.Header.Get("Origin")
	if cors == "*" || strings.Contains(cors, origin) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "*")
	}

	if method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
	}

	c.Next()
}

func logger(c *gin.Context) {

	// 开始时间
	startTime := time.Now()

	// 处理请求
	c.Next()

	// 结束时间
	endTime := time.Now()

	// 执行时间
	latencyTime := endTime.Sub(startTime)

	// 请求路由
	reqUri := c.Request.RequestURI

	// 状态码
	statusCode := c.Writer.Status()

	// 请求IP
	clientIP := c.ClientIP()

	// 日志格式
	logrus.Infoln("uri", reqUri, "status_code", statusCode, "cost", latencyTime, "client_ip", clientIP)
}
