package models

import (
	"time"

	"gorm.io/gorm"
)

// AlertRule 存储告警规则的数据模型
type AlertRule struct {
	gorm.Model
	Name        string    `gorm:"type:text" json:"name"`
	Alert       string    `gorm:"uniqueIndex;not null" json:"alert"`
	Expr        string    `gorm:"not null" json:"expr"`
	For         string    `gorm:"not null" json:"for"`
	Labels      string    `gorm:"type:text" json:"labels"`      // 存储JSON格式的标签
	Annotations string    `gorm:"type:text" json:"annotations"` // 存储JSON格式的注释
	GroupName   string    `gorm:"index;not null" json:"group_name"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AlertRule) TableName() string {
	return "alert_rules"
}
