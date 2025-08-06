package handler

import (
	"net/http"
	"os"
	
	"github.com/gin-gonic/gin"
)

var (
	app *gin.Engine
	initialized = false
)

func init() {
	if !initialized {
		// 设置 Gin 模式
		gin.SetMode(gin.ReleaseMode)
		
		// 创建 Gin 实例
		app = gin.New()
		app.Use(gin.Recovery())
		
		// 添加基础路由
		app.GET("/api/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"service": "openpenpal-backend",
				"serverless": true,
			})
		})
		
		// TODO: 这里需要导入你的路由配置
		// setupRoutes(app)
		
		initialized = true
	}
}

// Handler 是 Vercel Serverless Function 的入口
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}