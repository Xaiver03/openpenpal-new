// Add these routes to main.go in the appropriate sections

// PUBLIC MUSEUM ROUTES - Add after line 174 in the public museum group:
museum.GET("/popular", museumHandler.GetPopularMuseumEntries)        // 获取热门条目
museum.GET("/exhibitions/:id", museumHandler.GetMuseumExhibitionByID) // 获取展览详情
museum.GET("/tags", museumHandler.GetMuseumTags)                     // 获取标签列表

// PROTECTED MUSEUM ROUTES - Add after line 266 in the protected museum group:
museum.POST("/entries/:id/interact", museumHandler.InteractWithEntry)  // 记录互动（浏览、点赞等）
museum.POST("/entries/:id/react", museumHandler.ReactToEntry)         // 添加反应
museum.DELETE("/entries/:id/withdraw", museumHandler.WithdrawMuseumEntry) // 撤回条目
museum.GET("/my-submissions", museumHandler.GetMySubmissions)         // 获取我的提交记录

// ADMIN MUSEUM ROUTES - Add in the admin section (create new admin museum group):
// 博物馆管理相关
museumAdmin := admin.Group("/museum")
{
    museumAdmin.POST("/entries/:id/moderate", museumHandler.ModerateMuseumEntry)    // 审核条目
    museumAdmin.GET("/entries/pending", museumHandler.GetPendingMuseumEntries)      // 获取待审核条目
    museumAdmin.POST("/exhibitions", museumHandler.CreateMuseumExhibition)          // 创建展览
    museumAdmin.PUT("/exhibitions/:id", museumHandler.UpdateMuseumExhibition)       // 更新展览
    museumAdmin.DELETE("/exhibitions/:id", museumHandler.DeleteMuseumExhibition)    // 删除展览
    museumAdmin.POST("/refresh-stats", museumHandler.RefreshMuseumStats)            // 刷新统计数据
    museumAdmin.GET("/analytics", museumHandler.GetMuseumAnalytics)                 // 获取分析数据
}