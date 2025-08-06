package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProxyToCourierService 创建一个代理中间件，将请求转发到courier-service
func ProxyToCourierService(targetPath string) gin.HandlerFunc {
	courierServiceURL := "http://localhost:8002" // courier-service地址
	
	return func(c *gin.Context) {
		// 构建目标URL
		targetURL := courierServiceURL + targetPath
		
		// 处理路径参数
		for _, param := range c.Params {
			targetURL = strings.Replace(targetURL, ":"+param.Key, param.Value, 1)
		}
		
		// 添加查询参数
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}
		
		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		
		// 创建新请求
		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "代理请求创建失败",
				"error":   err.Error(),
			})
			return
		}
		
		// 复制请求头
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		
		// 从gin上下文中获取认证信息并转发
		if token := c.GetHeader("Authorization"); token != "" {
			req.Header.Set("Authorization", token)
		}
		
		// 设置额外的代理相关头部
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		
		// 执行请求
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			// 如果courier-service不可用，返回友好的错误信息
			if strings.Contains(err.Error(), "connection refused") {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code":    503,
					"message": "信使服务暂时不可用",
					"error":   "courier-service未启动或无法连接",
				})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    502,
				"message": "代理请求失败",
				"error":   err.Error(),
			})
			return
		}
		defer resp.Body.Close()
		
		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "读取响应失败",
				"error":   err.Error(),
			})
			return
		}
		
		// 复制响应头
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}
		
		// 设置状态码并返回响应
		c.Status(resp.StatusCode)
		c.Writer.Write(respBody)
	}
}

// ProxyToService 通用的服务代理中间件
func ProxyToService(serviceURL string, pathPrefix string) gin.HandlerFunc {
	parsedURL, _ := url.Parse(serviceURL)
	
	return func(c *gin.Context) {
		// 获取相对路径
		relativePath := strings.TrimPrefix(c.Request.URL.Path, pathPrefix)
		
		// 构建目标URL
		targetURL := serviceURL + relativePath
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}
		
		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		
		// 创建代理请求
		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建代理请求失败",
				"error":   err.Error(),
			})
			return
		}
		
		// 复制请求头
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		
		// 设置代理头部
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Host = parsedURL.Host
		
		// 执行请求
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    502,
				"message": fmt.Sprintf("无法连接到%s服务", parsedURL.Host),
				"error":   err.Error(),
			})
			return
		}
		defer resp.Body.Close()
		
		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "读取服务响应失败",
				"error":   err.Error(),
			})
			return
		}
		
		// 处理Content-Type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			c.Header("Content-Type", contentType)
		}
		
		// 返回响应
		c.Status(resp.StatusCode)
		
		// 如果是JSON响应，尝试美化输出
		if strings.Contains(contentType, "application/json") {
			var jsonData interface{}
			if err := json.Unmarshal(respBody, &jsonData); err == nil {
				c.JSON(resp.StatusCode, jsonData)
				return
			}
		}
		
		// 否则直接输出原始响应
		c.Writer.Write(respBody)
	}
}