package handlers

import (
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// SystemSettingsHandler 系统设置处理器
type SystemSettingsHandler struct {
	settingsService *services.SystemSettingsService
	config          *config.Config
}

// NewSystemSettingsHandler 创建系统设置处理器实例
func NewSystemSettingsHandler(settingsService *services.SystemSettingsService, config *config.Config) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		settingsService: settingsService,
		config:          config,
	}
}

// GetSettings 获取系统设置
// @Summary 获取系统设置
// @Description 获取完整的系统配置信息
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} models.SystemConfig
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/settings [get]
func (h *SystemSettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.GetSystemConfig()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取系统设置失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "获取系统设置成功", gin.H{
		"code": 0,
		"data": settings,
		"message": "获取系统设置成功",
	})
}

// UpdateSettings 更新系统设置
// @Summary 更新系统设置
// @Description 更新系统配置信息
// @Tags admin
// @Accept json
// @Produce json
// @Param settings body models.SystemConfig true "系统配置"
// @Success 200 {object} models.SystemConfig
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/settings [put]
func (h *SystemSettingsHandler) UpdateSettings(c *gin.Context) {
	var newConfig models.SystemConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		utils.BadRequestResponse(c, "请求参数无效", err)
		return
	}

	// 验证配置
	if err := h.validateConfig(&newConfig); err != nil {
		utils.BadRequestResponse(c, "配置验证失败", err)
		return
	}

	// 更新配置
	updatedConfig, err := h.settingsService.UpdateSystemConfig(&newConfig)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新系统设置失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "更新系统设置成功", gin.H{
		"code": 0,
		"data": updatedConfig,
		"message": "配置保存成功！",
	})
}

// ResetSettings 重置系统设置
// @Summary 重置系统设置
// @Description 将系统设置重置为默认值
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} models.SystemConfig
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/settings [post]
func (h *SystemSettingsHandler) ResetSettings(c *gin.Context) {
	defaultConfig, err := h.settingsService.ResetToDefaults()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "重置系统设置失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "重置系统设置成功", gin.H{
		"code": 0,
		"data": defaultConfig,
		"message": "配置已重置为默认值！",
	})
}

// TestEmailConfig 测试邮件配置
// @Summary 测试邮件配置
// @Description 发送测试邮件以验证SMTP配置
// @Tags admin
// @Accept json
// @Produce json
// @Param config body map[string]interface{} true "包含邮件配置和测试邮箱的请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/settings/test-email [post]
func (h *SystemSettingsHandler) TestEmailConfig(c *gin.Context) {
	var req struct {
		SMTPHost         string `json:"smtp_host"`
		SMTPPort         int    `json:"smtp_port"`
		SMTPUsername     string `json:"smtp_username"`
		SMTPPassword     string `json:"smtp_password"`
		SMTPEncryption   string `json:"smtp_encryption"`
		EmailFromName    string `json:"email_from_name"`
		EmailFromAddress string `json:"email_from_address"`
		TestEmail        string `json:"test_email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数无效", err)
		return
	}

	// 验证测试邮箱
	if req.TestEmail == "" {
		utils.BadRequestResponse(c, "测试邮箱地址不能为空", nil)
		return
	}

	// 创建临时配置用于测试
	testConfig := &models.SystemConfig{
		SMTPHost:         req.SMTPHost,
		SMTPPort:         req.SMTPPort,
		SMTPUsername:     req.SMTPUsername,
		SMTPPassword:     req.SMTPPassword,
		SMTPEncryption:   req.SMTPEncryption,
		EmailFromName:    req.EmailFromName,
		EmailFromAddress: req.EmailFromAddress,
	}

	// 验证邮件配置
	if err := h.settingsService.ValidateEmailConfig(testConfig); err != nil {
		utils.BadRequestResponse(c, "邮件配置验证失败", err)
		return
	}

	// 发送测试邮件
	if err := h.sendTestEmail(testConfig, req.TestEmail); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "测试邮件发送失败", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "测试邮件发送成功", gin.H{
		"code": 0,
		"message": "测试邮件发送成功！请检查邮箱: " + req.TestEmail,
	})
}

// validateConfig 验证系统配置
func (h *SystemSettingsHandler) validateConfig(config *models.SystemConfig) error {
	// 基本验证
	if config.SiteName == "" {
		return &utils.ValidationError{Field: "site_name", Message: "网站名称不能为空"}
	}

	// 数值范围验证
	if config.MaxLetterLength < 100 || config.MaxLetterLength > 100000 {
		return &utils.ValidationError{Field: "max_letter_length", Message: "信件长度限制必须在100-100000之间"}
	}

	if config.MaxFileSize < 1 || config.MaxFileSize > 100 {
		return &utils.ValidationError{Field: "max_file_size", Message: "文件大小限制必须在1-100MB之间"}
	}

	if config.PasswordMinLength < 6 || config.PasswordMinLength > 50 {
		return &utils.ValidationError{Field: "password_min_length", Message: "密码长度要求必须在6-50之间"}
	}

	if config.SessionTimeout < 300 || config.SessionTimeout > 86400 {
		return &utils.ValidationError{Field: "session_timeout", Message: "会话超时时间必须在5分钟到24小时之间"}
	}

	// 邮件配置验证（如果启用了邮件通知）
	if config.EmailNotifications {
		if err := h.settingsService.ValidateEmailConfig(config); err != nil {
			return err
		}
	}

	return nil
}

// sendTestEmail 发送测试邮件
func (h *SystemSettingsHandler) sendTestEmail(config *models.SystemConfig, toEmail string) error {
	// 构建测试邮件内容
	subject := "OpenPenPal 系统测试邮件"
	body := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>测试邮件</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #007bff;">OpenPenPal 系统测试邮件</h1>
        <p>恭喜！您的邮件配置已成功。</p>
        <p>这是一封来自 OpenPenPal 系统的测试邮件，用于验证您的SMTP设置是否正确。</p>
        <hr style="border: 1px solid #eee; margin: 20px 0;">
        <p><strong>配置信息：</strong></p>
        <ul>
            <li>SMTP服务器: ` + config.SMTPHost + `</li>
            <li>端口: ` + strconv.Itoa(config.SMTPPort) + `</li>
            <li>加密方式: ` + config.SMTPEncryption + `</li>
            <li>发件人: ` + config.EmailFromName + ` &lt;` + config.EmailFromAddress + `&gt;</li>
        </ul>
        <hr style="border: 1px solid #eee; margin: 20px 0;">
        <p style="color: #666; font-size: 14px;">此邮件由 OpenPenPal 系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`

	// 设置邮件头
	headers := make(map[string]string)
	headers["From"] = config.EmailFromName + " <" + config.EmailFromAddress + ">"
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += k + ": " + v + "\r\n"
	}
	message += "\r\n" + body

	// SMTP认证
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	err := smtp.SendMail(addr, auth, config.EmailFromAddress, []string{toEmail}, []byte(message))
	
	return err
}