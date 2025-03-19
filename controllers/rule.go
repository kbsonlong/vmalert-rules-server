package controllers

import (
	"encoding/json"
	"net/http"
	"vmalert-rules/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RuleController struct {
	DB *gorm.DB
}

// CreateRule 创建新的告警规则
func (rc *RuleController) CreateRule(c *gin.Context) {
	var requestData struct {
		Name        string            `json:"name"`
		Alert       string            `json:"alert"`
		Expr        string            `json:"expr"`
		For         string            `json:"for"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		GroupName   string            `json:"group_name"`
		Enabled     bool              `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将labels和annotations转换为JSON字符串
	labelsBytes, err := json.Marshal(requestData.Labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid labels format"})
		return
	}

	annotationsBytes, err := json.Marshal(requestData.Annotations)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid annotations format"})
		return
	}

	rule := models.AlertRule{
		Name:        requestData.Name,
		Alert:       requestData.Alert,
		Expr:        requestData.Expr,
		For:         requestData.For,
		Labels:      string(labelsBytes),
		Annotations: string(annotationsBytes),
		GroupName:   requestData.GroupName,
		Enabled:     requestData.Enabled,
	}

	if err := rc.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetRule 获取单个告警规则
func (rc *RuleController) GetRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.AlertRule

	if err := rc.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// ListRules 获取所有告警规则
func (rc *RuleController) ListRules(c *gin.Context) {
	var rules []models.AlertRule
	groupName := c.Query("group_name")

	db := rc.DB
	if groupName != "" {
		db = db.Where("group_name = ?", groupName)
	}

	if err := db.Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// UpdateRule 更新告警规则
func (rc *RuleController) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.AlertRule

	if err := rc.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	var requestData struct {
		Name        string            `json:"name"`
		Alert       string            `json:"alert"`
		Expr        string            `json:"expr"`
		For         string            `json:"for"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		GroupName   string            `json:"group_name"`
		Enabled     bool              `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将labels和annotations转换为JSON字符串
	labelsBytes, err := json.Marshal(requestData.Labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid labels format"})
		return
	}

	annotationsBytes, err := json.Marshal(requestData.Annotations)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid annotations format"})
		return
	}

	rule.Name = requestData.Name
	rule.Alert = requestData.Alert
	rule.Expr = requestData.Expr
	rule.For = requestData.For
	rule.Labels = string(labelsBytes)
	rule.Annotations = string(annotationsBytes)
	rule.GroupName = requestData.GroupName
	rule.Enabled = requestData.Enabled

	if err := rc.DB.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteRule 删除告警规则
func (rc *RuleController) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.AlertRule

	if err := rc.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	if err := rc.DB.Delete(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}
