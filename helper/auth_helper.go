package helper

import (
	"github.com/gin-gonic/gin"
)

// CheckUserRoleAndID memeriksa apakah current_id dan role ada di context
// dan apakah role sesuai dengan salah satu yang diharapkan.
func CheckUserRoleAndID(c *gin.Context, expectedRoles ...string) (uint, string, bool) {
	currentUserID, currentUserExists := c.Get("current_id")
	role, roleExists := c.Get("role")

	if !currentUserExists || !roleExists {
		c.JSON(403, gin.H{"error": "Unauthorized"})
		return 0, "", false
	}

	for _, expectedRole := range expectedRoles {
		if role == expectedRole {
			return currentUserID.(uint), role.(string), true
		}
	}

	c.JSON(403, gin.H{"error": "Forbidden: Invalid role"})
	return 0, "", false
}
