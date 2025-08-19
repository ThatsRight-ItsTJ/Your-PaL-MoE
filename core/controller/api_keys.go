package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy/core/model"
	"github.com/labring/aiproxy/core/pkg"
	log "github.com/sirupsen/logrus"
)

// POST /api/v1/keys - Create new API key
func CreateAPIKey(c *gin.Context) {
	var params pkg.CreateKeyParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if params.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if params.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Create the API key
	key, err := pkg.CreateAPIKey(c.Request.Context(), params)
	if err != nil {
		log.Errorf("Failed to create API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    key,
	})
}

// GET /api/v1/keys - List user's API keys
func ListAPIKeys(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Query("user_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	order := c.DefaultQuery("order", "id-desc")

	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Convert user ID to group ID for compatibility
	groupID := "user_" + strconv.Itoa(userID)
	
	tokens, total, err := model.GetTokens(groupID, page, perPage, order, status)
	if err != nil {
		log.Errorf("Failed to get tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve API keys",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tokens": tokens,
			"total":  total,
			"page":   page,
			"per_page": perPage,
		},
	})
}

// GET /api/v1/keys/:keyId - Get specific key details
func GetAPIKey(c *gin.Context) {
	keyID, err := strconv.Atoi(c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	token, err := model.GetTokenByID(keyID)
	if err != nil {
		if model.HandleNotFound(err, model.ErrTokenNotFound) != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		log.Errorf("Failed to get token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token,
	})
}

// PUT /api/v1/keys/:keyId - Update key settings
func UpdateAPIKey(c *gin.Context) {
	keyID, err := strconv.Atoi(c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	var updateReq model.UpdateTokenRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	token, err := model.UpdateToken(keyID, updateReq)
	if err != nil {
		if model.HandleNotFound(err, model.ErrTokenNotFound) != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		log.Errorf("Failed to update token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token,
	})
}

// DELETE /api/v1/keys/:keyId - Delete/disable key
func DeleteAPIKey(c *gin.Context) {
	keyID, err := strconv.Atoi(c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	err = model.DeleteTokenByID(keyID)
	if err != nil {
		if model.HandleNotFound(err, model.ErrTokenNotFound) != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		log.Errorf("Failed to delete token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key deleted successfully",
	})
}

// POST /api/v1/keys/:keyId/rotate - Rotate key
func RotateAPIKey(c *gin.Context) {
	keyID, err := strconv.Atoi(c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	rotatedKey, err := pkg.RotateAPIKey(c.Request.Context(), keyID)
	if err != nil {
		log.Errorf("Failed to rotate API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to rotate API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rotatedKey,
	})
}

// GET /api/v1/keys/:keyId/usage - Get usage analytics
func GetAPIKeyUsage(c *gin.Context) {
	keyID, err := strconv.Atoi(c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	token, err := model.GetTokenByID(keyID)
	if err != nil {
		if model.HandleNotFound(err, model.ErrTokenNotFound) != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		log.Errorf("Failed to get token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve API key",
			"details": err.Error(),
		})
		return
	}

	usage := gin.H{
		"key_id":        token.ID,
		"request_count": token.RequestCount,
		"used_amount":   token.UsedAmount,
		"quota":         token.Quota,
		"created_at":    token.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    usage,
	})
}

// GET /api/v1/keys/:keyId/audit - Get audit log for key
func GetAPIKeyAudit(c *gin.Context) {
	keyID, err := strconv.Atoi(c.Param("keyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	action := c.Query("action")

	logs, total, err := model.GetAuditLogs(nil, &keyID, action, "", page, perPage)
	if err != nil {
		log.Errorf("Failed to get audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve audit logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":     logs,
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	})
}

// POST /api/v1/keys/validate - Validate API key
func ValidateAPIKeyEndpoint(c *gin.Context) {
	var request struct {
		Key string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := pkg.ValidateAPIKey(c.Request.Context(), request.Key)
	if err != nil {
		log.Errorf("Failed to validate API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to validate API key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}