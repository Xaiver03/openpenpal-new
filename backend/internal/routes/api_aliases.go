package routes

import (
	"strings"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/utils"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SetupAPIAliases creates route aliases to fix frontend-backend API mismatches
// This is a SOTA solution that maintains backward compatibility
func SetupAPIAliases(router *gin.Engine) {
	v1 := router.Group("/api")
	
	// Authentication route aliases
	authAlias := v1.Group("/auth")
	{
		// Map frontend expected routes to actual backend handlers
		authAlias.POST("/login", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/auth/login"
			router.HandleContext(c)
		})
		
		authAlias.POST("/register", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/auth/register"
			router.HandleContext(c)
		})
		
		authAlias.POST("/logout", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/auth/logout"
			router.HandleContext(c)
		})
		
		authAlias.GET("/me", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/users/me"
			router.HandleContext(c)
		})
		
		authAlias.POST("/refresh", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/auth/refresh-token"
			router.HandleContext(c)
		})
		
		// CSRF endpoint - using real CSRF protection
		authAlias.GET("/csrf", middleware.GetCSRFTokenHandler)
		
		// Validation endpoints
		authAlias.POST("/check-username", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequestResponse(c, "Invalid request", err)
				return
			}
			
			// TODO: Check against actual database
			utils.SuccessResponse(c, 200, "Username availability checked", gin.H{
				"available": true, // Mock response
				"username": req.Username,
			})
		})
		
		authAlias.POST("/check-email", func(c *gin.Context) {
			var req struct {
				Email string `json:"email" binding:"required,email"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequestResponse(c, "Invalid request", err)
				return
			}
			
			// TODO: Check against actual database
			utils.SuccessResponse(c, 200, "Email availability checked", gin.H{
				"available": true, // Mock response
				"email": req.Email,
			})
		})
	}
	
	// School endpoints
	v1.GET("/schools", func(c *gin.Context) {
		// Return list of schools
		schools := []gin.H{
			{"code": "PKU001", "name": "北京大学", "city": "北京"},
			{"code": "THU001", "name": "清华大学", "city": "北京"},
			{"code": "RUC001", "name": "中国人民大学", "city": "北京"},
			{"code": "BNU001", "name": "北京师范大学", "city": "北京"},
			{"code": "BUAA001", "name": "北京航空航天大学", "city": "北京"},
		}
		
		// Handle search query
		if search := c.Query("search"); search != "" {
			filtered := []gin.H{}
			for _, school := range schools {
				if contains(school["name"].(string), search) || contains(school["code"].(string), search) {
					filtered = append(filtered, school)
				}
			}
			schools = filtered
		}
		
		utils.SuccessResponse(c, 200, "Schools retrieved", gin.H{
			"schools": schools,
			"total": len(schools),
		})
	})
	
	v1.POST("/schools/validate", func(c *gin.Context) {
		var req struct {
			Code string `json:"code" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequestResponse(c, "Invalid request", err)
			return
		}
		
		validSchools := map[string]bool{
			"PKU001": true,
			"THU001": true,
			"RUC001": true,
			"BNU001": true,
			"BUAA001": true,
		}
		
		utils.SuccessResponse(c, 200, "School code validated", gin.H{
			"valid": validSchools[req.Code],
			"code": req.Code,
		})
	})
	
	// Postcode endpoints
	v1.GET("/postcode/:code", func(c *gin.Context) {
		code := c.Param("code")
		
		// Mock postcode data - TODO: Integrate with real postcode service
		postcodes := map[string]gin.H{
			"100080": {
				"postcode": "100080",
				"province": "北京市",
				"city": "北京市",
				"district": "海淀区",
				"address": "中关村",
			},
			"100084": {
				"postcode": "100084",
				"province": "北京市",
				"city": "北京市",
				"district": "海淀区",
				"address": "清华大学",
			},
		}
		
		if data, ok := postcodes[code]; ok {
			utils.SuccessResponse(c, 200, "Postcode found", data)
		} else {
			utils.NotFoundResponse(c, "Postcode not found")
		}
	})
	
	// Address search
	v1.GET("/address/search", func(c *gin.Context) {
		query := c.Query("q")
		limit := c.DefaultQuery("limit", "10")
		
		// Mock address search - TODO: Integrate with real address service
		results := []gin.H{
			{
				"address": query + " - 北京市海淀区中关村南大街1号",
				"postcode": "100080",
				"lat": 39.95933,
				"lng": 116.31785,
			},
			{
				"address": query + " - 北京市海淀区学院路30号",
				"postcode": "100083",
				"lat": 39.99140,
				"lng": 116.35215,
			},
		}
		
		utils.SuccessResponse(c, 200, "Address search results", gin.H{
			"results": results,
			"query": query,
			"limit": limit,
		})
	})
	
	// Admin permission endpoints (temporary implementation)
	adminGroup := v1.Group("/admin/permissions")
	{
		adminGroup.GET("", func(c *gin.Context) {
			permType := c.Query("type")
			
			var response interface{}
			switch permType {
			case "overview":
				response = gin.H{
					"total_roles": 5,
					"total_permissions": 20,
					"total_users": 100,
				}
			case "roles":
				response = []gin.H{
					{"id": "1", "name": "admin", "permissions": 20},
					{"id": "2", "name": "user", "permissions": 5},
				}
			case "courier-levels":
				response = []gin.H{
					{"level": 1, "name": "楼栋信使", "permissions": 3},
					{"level": 2, "name": "片区信使", "permissions": 5},
					{"level": 3, "name": "校级信使", "permissions": 8},
					{"level": 4, "name": "城市总代", "permissions": 12},
				}
			default:
				response = gin.H{"permissions": []string{}}
			}
			
			utils.SuccessResponse(c, 200, "Permissions retrieved", response)
		})
		
		adminGroup.POST("", func(c *gin.Context) {
			// Mock permission update
			utils.SuccessResponse(c, 200, "Permissions updated", gin.H{
				"updated": true,
			})
		})
		
		adminGroup.GET("/audit", func(c *gin.Context) {
			// Mock audit log
			utils.SuccessResponse(c, 200, "Audit log retrieved", gin.H{
				"logs": []gin.H{
					{
						"id": "1",
						"action": "permission_updated",
						"user": "admin",
						"timestamp": "2025-01-26T10:00:00Z",
					},
				},
			})
		})
	}
	
	// Letter route aliases
	lettersAlias := v1.Group("/letters")
	{
		lettersAlias.GET("/public", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/letters/public"
			router.HandleContext(c)
		})
		
		lettersAlias.GET("/popular", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/letters/popular"
			router.HandleContext(c)
		})
		
		lettersAlias.GET("/recommended", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/letters/recommended"
			router.HandleContext(c)
		})
		
		lettersAlias.GET("/templates", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/letters/templates"
			router.HandleContext(c)
		})
	}
	
	// AI route aliases - SOTA fix for frontend compatibility
	aiAlias := v1.Group("/ai")
	{
		aiAlias.GET("/daily-inspiration", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/daily-inspiration"
			router.HandleContext(c)
		})
		
		aiAlias.GET("/stats", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/stats"
			router.HandleContext(c)
		})
		
		aiAlias.POST("/inspiration", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/inspiration"
			router.HandleContext(c)
		})
		
		aiAlias.POST("/match", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/match"
			router.HandleContext(c)
		})
		
		aiAlias.POST("/reply", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/reply"
			router.HandleContext(c)
		})
		
		aiAlias.POST("/reply-advice", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/reply-advice"
			router.HandleContext(c)
		})
		
		aiAlias.POST("/curate", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/curate"
			router.HandleContext(c)
		})
		
		aiAlias.GET("/personas", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/ai/personas"
			router.HandleContext(c)
		})
	}
	
	// Shop route aliases
	shopAlias := v1.Group("/shop")
	{
		shopAlias.GET("/products", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/shop/products"
			router.HandleContext(c)
		})
		
		shopAlias.GET("/products/:id", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/shop/products/" + c.Param("id")
			router.HandleContext(c)
		})
		
		shopAlias.GET("/products/:id/reviews", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/shop/products/" + c.Param("id") + "/reviews"
			router.HandleContext(c)
		})
	}
	
	// Museum route aliases
	museumAlias := v1.Group("/museum")
	{
		museumAlias.GET("/entries", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/entries"
			router.HandleContext(c)
		})
		
		museumAlias.GET("/entries/:id", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/entries/" + c.Param("id")
			router.HandleContext(c)
		})
		
		museumAlias.GET("/stats", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/stats"
			router.HandleContext(c)
		})
		
		museumAlias.GET("/exhibitions", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/exhibitions"
			router.HandleContext(c)
		})
		
		museumAlias.GET("/exhibitions/:id", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/exhibitions/" + c.Param("id")
			router.HandleContext(c)
		})
		
		museumAlias.GET("/exhibitions/:id/items", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/exhibitions/" + c.Param("id") + "/items"
			router.HandleContext(c)
		})
		
		museumAlias.GET("/popular", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/popular"
			router.HandleContext(c)
		})
		
		museumAlias.GET("/tags", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/tags"
			router.HandleContext(c)
		})
		
		museumAlias.POST("/submit", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/submit"
			router.HandleContext(c)
		})
		
		museumAlias.GET("/my-submissions", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/my-submissions"
			router.HandleContext(c)
		})
		
		museumAlias.POST("/search", func(c *gin.Context) {
			c.Request.URL.Path = "/api/v1/museum/search"
			router.HandleContext(c)
		})
	}
	
	// User profile endpoints
	usersAlias := v1.Group("/users")
	{
		// Get user profile by username
		usersAlias.GET("/:username/profile", func(c *gin.Context) {
			username := c.Param("username")
			if username == "" {
				utils.BadRequestResponse(c, "用户名不能为空", nil)
				return
			}
			
			// Mock user profile - TODO: Implement real user profile service
			profiles := map[string]gin.H{
				"alice": {
					"id":       1,
					"username": "alice",
					"nickname": "Alice Smith",
					"email":    "alice@example.com",
					"role":     "student",
					"bio":      "爱好写信的学生，希望通过文字传递温暖",
					"school":   "北京大学",
					"created_at": "2024-01-15T08:00:00Z",
					"op_code":  "PK5F3D",
					"writing_level": 3,
					"courier_level": 0,
					"stats": gin.H{
						"letters_sent":        15,
						"letters_received":    12,
						"museum_contributions": 3,
						"total_points":        450,
						"writing_points":      320,
						"courier_points":      0,
						"current_streak":      7,
						"achievements":        []string{"first_letter", "active_writer", "museum_contributor"},
					},
					"privacy": gin.H{
						"show_email":    false,
						"show_op_code":  true,
						"show_stats":    true,
						"op_code_privacy": "partial", // full, partial, hidden
					},
				},
				"admin": {
					"id":       2,
					"username": "admin",
					"nickname": "系统管理员",
					"role":     "super_admin",
					"bio":      "OpenPenPal系统管理员，维护平台运行",
					"school":   "北京大学",
					"created_at": "2024-01-01T00:00:00Z",
					"op_code":  "PK1L01",
					"writing_level": 5,
					"courier_level": 4,
					"stats": gin.H{
						"letters_sent":        5,
						"letters_received":    8,
						"museum_contributions": 10,
						"total_points":        1000,
						"writing_points":      600,
						"courier_points":      400,
						"current_streak":      30,
						"achievements":        []string{"system_admin", "master_writer", "city_coordinator", "museum_curator"},
					},
					"privacy": gin.H{
						"show_email":    false,
						"show_op_code":  true,
						"show_stats":    true,
						"op_code_privacy": "full",
					},
				},
			}
			
			if profile, exists := profiles[username]; exists {
				utils.SuccessResponse(c, 200, "获取用户资料成功", profile)
			} else {
				utils.NotFoundResponse(c, "用户不存在")
			}
		})
		
		// Get user letters by username
		usersAlias.GET("/:username/letters", func(c *gin.Context) {
			username := c.Param("username")
			publicOnly := c.Query("public") == "true"
			
			if username == "" {
				utils.BadRequestResponse(c, "用户名不能为空", nil)
				return
			}
			
			// Mock user letters - TODO: Implement real letter service integration
			userLetters := map[string][]gin.H{
				"alice": {
					{
						"id":              1,
						"title":           "给远方朋友的问候",
						"content_preview": "亲爱的朋友，最近过得怎么样？我想和你分享一些生活中的小确幸...",
						"created_at":      "2024-01-20T10:30:00Z",
						"status":          "delivered",
						"recipient_username": "bob",
						"sender_username":   "alice",
						"is_public":        true,
					},
					{
						"id":              2,
						"title":           "冬日暖阳",
						"content_preview": "今天的阳光特别温暖，让我想起了童年的冬天...",
						"created_at":      "2024-01-18T15:20:00Z",
						"status":          "delivered",
						"recipient_username": "carol",
						"sender_username":   "alice",
						"is_public":        true,
					},
				},
				"admin": {
					{
						"id":              3,
						"title":           "系统维护通知",
						"content_preview": "亲爱的用户们，系统将于本周末进行例行维护...",
						"created_at":      "2024-01-15T09:00:00Z",
						"status":          "delivered",
						"recipient_username": "all_users",
						"sender_username":   "admin",
						"is_public":        true,
					},
				},
			}
			
			letters, exists := userLetters[username]
			if !exists {
				letters = []gin.H{}
			}
			
			// Filter public letters if requested
			if publicOnly {
				publicLetters := []gin.H{}
				for _, letter := range letters {
					if isPublic, ok := letter["is_public"].(bool); ok && isPublic {
						publicLetters = append(publicLetters, letter)
					}
				}
				letters = publicLetters
			}
			
			utils.SuccessResponse(c, 200, "获取用户信件成功", gin.H{
				"letters": letters,
				"count":   len(letters),
			})
		})
	}

	// Error reporting endpoint
	v1.POST("/errors/report", func(c *gin.Context) {
		var req struct {
			Error   string `json:"error"`
			Stack   string `json:"stack"`
			Context gin.H  `json:"context"`
		}
		c.ShouldBindJSON(&req)
		
		// Log error (in production, this would go to error tracking service)
		utils.SuccessResponse(c, 200, "Error reported", gin.H{
			"reported": true,
			"id": "err_" + uuid.New().String(),
		})
	})
}

// contains is a simple string contains helper
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}