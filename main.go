package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"vmalert-rules/controllers"
	"vmalert-rules/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Rule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

type Group struct {
	Name        string `yaml:"name"`
	Concurrency int    `yaml:"concurrency"`
	Interval    int    `yaml:"interval"`
	Rules       []Rule `yaml:"rules"`
}

type Config struct {
	Groups []Group `yaml:"groups"`
}

var (
	groupCount      int
	ruleCount       int
	config          Config
	mutex           sync.RWMutex
	enableAutoRules bool
	db              *gorm.DB
)

func init() {
	flag.IntVar(&groupCount, "groups", 1, "number of groups to generate")
	flag.IntVar(&ruleCount, "rules", 1, "number of rules per group")
	flag.BoolVar(&enableAutoRules, "enable-auto-rules", true, "enable auto-generated rules")
}

func loadTemplate() error {
	data, err := ioutil.ReadFile("template.yaml")
	if err != nil {
		return err
	}

	var templateConfig Config
	if err := yaml.Unmarshal(data, &templateConfig); err != nil {
		return err
	}

	if len(templateConfig.Groups) == 0 || len(templateConfig.Groups[0].Rules) == 0 {
		return fmt.Errorf("template must contain at least one group with one rule")
	}

	mutex.Lock()
	defer mutex.Unlock()

	// 获取所有可用的规则
	allRules := templateConfig.Groups[0].Rules
	totalRules := len(allRules)
	if totalRules == 0 {
		return fmt.Errorf("no rules found in template")
	}

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	config.Groups = make([]Group, groupCount)
	for i := 0; i < groupCount; i++ {
		config.Groups[i] = Group{
			Name:        fmt.Sprintf("group_%d", i+1),
			Concurrency: templateConfig.Groups[0].Concurrency,
			Interval:    templateConfig.Groups[0].Interval,
			Rules:       make([]Rule, ruleCount),
		}

		for j := 0; j < ruleCount; j++ {
			// 随机选择一个规则
			randomIndex := rand.Intn(totalRules)
			templateRule := allRules[randomIndex]
			config.Groups[i].Rules[j] = Rule{
				Alert:       fmt.Sprintf("group_%d_%s_%d", i+1, templateRule.Alert, j+1),
				Expr:        templateRule.Expr,
				For:         templateRule.For,
				Labels:      templateRule.Labels,
				Annotations: templateRule.Annotations,
			}
		}
	}

	return nil
}

func getRules(c *gin.Context) {
	var dbRules []models.AlertRule
	if err := db.Find(&dbRules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换数据库规则为YAML格式
	groups := make(map[string]*Group)
	for _, rule := range dbRules {
		if _, exists := groups[rule.GroupName]; !exists {
			groups[rule.GroupName] = &Group{
				Name:     rule.GroupName,
				Interval: 30, // 默认值
				Rules:    []Rule{},
			}
		}

		// 解析Labels和Annotations
		var labels map[string]string
		var annotations map[string]string
		yaml.Unmarshal([]byte(rule.Labels), &labels)
		yaml.Unmarshal([]byte(rule.Annotations), &annotations)

		groups[rule.GroupName].Rules = append(groups[rule.GroupName].Rules, Rule{
			Alert:       rule.Alert,
			Expr:        rule.Expr,
			For:         rule.For,
			Labels:      labels,
			Annotations: annotations,
		})
	}

	// 合并自动生成的规则
	if enableAutoRules {
		mutex.RLock()
		for _, group := range config.Groups {
			if _, exists := groups[group.Name]; !exists {
				groups[group.Name] = &Group{
					Name:        group.Name,
					Concurrency: group.Concurrency,
					Interval:    group.Interval,
					Rules:       group.Rules,
				}
			} else {
				groups[group.Name].Rules = append(groups[group.Name].Rules, group.Rules...)
			}
		}
		mutex.RUnlock()
	}

	// 转换map为slice
	result := Config{Groups: make([]Group, 0, len(groups))}
	for _, group := range groups {
		result.Groups = append(result.Groups, *group)
	}

	format := c.DefaultQuery("format", "yaml")
	if format == "yaml" {
		c.Header("Content-Type", "application/x-yaml")
		data, err := yaml.Marshal(result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, string(data))
	} else {
		c.JSON(http.StatusOK, result)
	}
}

func main() {
	flag.Parse()

	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("rules.db"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	// 自动迁移数据库结构
	db.AutoMigrate(&models.AlertRule{})

	// 加载规则模板
	if err := loadTemplate(); err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		return
	}

	// 初始化控制器
	ruleController := &controllers.RuleController{DB: db}

	// 设置路由
	r := gin.Default()

	// 规则管理API
	r.POST("/api/rules", ruleController.CreateRule)
	r.GET("/api/rules/:id", ruleController.GetRule)
	r.GET("/api/rules", ruleController.ListRules)
	r.PUT("/api/rules/:id", ruleController.UpdateRule)
	r.DELETE("/api/rules/:id", ruleController.DeleteRule)

	// 保留原有的自动生成规则API
	r.GET("/rules", getRules)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
