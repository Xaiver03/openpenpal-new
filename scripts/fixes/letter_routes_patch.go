// Add these routes to main.go in the letters group (around line 225)

// PROTECTED LETTER ROUTES - Add after existing letter routes:

// 草稿管理
letters.GET("/drafts", letterHandler.GetDrafts)                     // 获取草稿列表
letters.POST("/:id/publish", letterHandler.PublishLetter)           // 发布信件

// 互动功能
letters.POST("/:id/like", letterHandler.LikeLetter)                // 点赞信件
letters.POST("/:id/share", letterHandler.ShareLetter)               // 分享信件

// 模板功能
letters.GET("/templates", letterHandler.GetLetterTemplates)         // 获取模板列表
letters.GET("/templates/:id", letterHandler.GetLetterTemplateByID)  // 获取模板详情

// 搜索和发现
letters.POST("/search", letterHandler.SearchLetters)                // 搜索信件
letters.GET("/popular", letterHandler.GetPopularLetters)            // 获取热门信件
letters.GET("/recommended", letterHandler.GetRecommendedLetters)    // 获取推荐信件

// 批量操作和导出
letters.POST("/batch", letterHandler.BatchOperateLetters)           // 批量操作
letters.POST("/export", letterHandler.ExportLetters)                // 导出信件

// 写作辅助
letters.POST("/auto-save", letterHandler.AutoSaveDraft)             // 自动保存草稿
letters.POST("/writing-suggestions", letterHandler.GetWritingSuggestions) // 获取写作建议