package middleware

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"openpenpal-backend/internal/models"
)

// RoleCompatibilityMiddleware ensures role compatibility between frontend and backend
// This middleware runs after AuthMiddleware and maps roles appropriately
func RoleCompatibilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from context (set by AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			// No user, nothing to map
			c.Next()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.Next()
			return
		}

		// Create a compatible user object with both role representations
		compatibleUser := &CompatibleUser{
			User:             user,
			FrontendRole:     models.GetFrontendRole(user.Role),
			CourierLevelInfo: models.GetCourierLevelInfo(user.Role),
		}

		// Store both the original user and compatible user
		c.Set("user", user)                      // Keep original for backend logic
		c.Set("compatible_user", compatibleUser) // For API responses

		// Also set frontend-compatible role for middleware checks
		c.Set("frontend_role", compatibleUser.FrontendRole)

		// If it's a courier role, also set the courier level
		if info := models.GetCourierLevelInfo(user.Role); info != nil {
			c.Set("courier_level", info.Level)
		}

		c.Next()
	}
}

// CompatibleUser wraps user with frontend-compatible fields
type CompatibleUser struct {
	*models.User
	FrontendRole     string                   `json:"role_frontend"`
	CourierLevelInfo *models.CourierLevelInfo `json:"courier_info,omitempty"`
}

// MarshalJSON customizes JSON output to include both role formats
func (cu *CompatibleUser) MarshalJSON() ([]byte, error) {
	// Create a map with all user fields
	userMap := map[string]interface{}{
		"id":           cu.ID,
		"username":     cu.Username,
		"email":        cu.Email,
		"nickname":     cu.Nickname,
		"avatar":       cu.Avatar,
		"school_code":  cu.SchoolCode,
		"zone_code":    "", // Not implemented yet
		"address_code": "", // Not implemented yet
		"is_active":    cu.IsActive,
		"created_at":   cu.CreatedAt,
		"updated_at":   cu.UpdatedAt,

		// Role compatibility fields
		"role":         cu.FrontendRole,      // Frontend expects this
		"backend_role": string(cu.User.Role), // Backend role for debugging
		"role_display": cu.GetRoleDisplayName(),

		// Privacy settings (defaults for now)
		"is_receiver":     true,  // Default to true
		"is_code_public":  false, // Default to private
		"allow_ai_penpal": true,  // Default to allow

		// Stats (defaults for now)
		"points": 0, // Default points
		"level":  1, // Default level
	}

	// Add courier info if applicable
	if cu.CourierLevelInfo != nil {
		userMap["courier_level"] = cu.CourierLevelInfo.Level
		userMap["courier_info"] = cu.CourierLevelInfo
	}

	return json.Marshal(userMap)
}

// GetCompatibleUser retrieves the compatible user from context
func GetCompatibleUser(c *gin.Context) (*CompatibleUser, bool) {
	if user, exists := c.Get("compatible_user"); exists {
		if compatibleUser, ok := user.(*CompatibleUser); ok {
			return compatibleUser, true
		}
	}
	return nil, false
}

// TransformUserResponse transforms a user model to frontend-compatible format
func TransformUserResponse(user *models.User) map[string]interface{} {
	response := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"nickname":    user.Nickname,
		"avatar":      user.Avatar,
		"school_code": user.SchoolCode,
		"is_active":   user.IsActive,
		"created_at":  user.CreatedAt,
		"updated_at":  user.UpdatedAt,

		// Role transformation
		"role": models.GetFrontendRole(user.Role),

		// Additional role info
		"role_display": user.GetRoleDisplayName(),
	}

	// Add courier level info if applicable
	if info := models.GetCourierLevelInfo(user.Role); info != nil {
		response["courier_level"] = info.Level
		response["is_courier"] = true
	} else {
		response["is_courier"] = false
	}

	return response
}
