package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
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
	groupCount int
	ruleCount  int
	config     Config
	mutex      sync.RWMutex
)

func init() {
	flag.IntVar(&groupCount, "groups", 1, "number of groups to generate")
	flag.IntVar(&ruleCount, "rules", 1, "number of rules per group")
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

	config.Groups = make([]Group, groupCount)
	for i := 0; i < groupCount; i++ {
		config.Groups[i] = Group{
			Name:        fmt.Sprintf("group_%d", i+1),
			Concurrency: templateConfig.Groups[0].Concurrency,
			Interval:    templateConfig.Groups[0].Interval,
			Rules:       make([]Rule, ruleCount),
		}

		for j := 0; j < ruleCount; j++ {
			templateRule := templateConfig.Groups[0].Rules[j%len(templateConfig.Groups[0].Rules)]
			config.Groups[i].Rules[j] = Rule{
				Alert:       fmt.Sprintf("%s_%d", templateRule.Alert, j+1),
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
	mutex.RLock()
	defer mutex.RUnlock()

	format := c.DefaultQuery("format", "yaml")
	if format == "yaml" {
		c.Header("Content-Type", "application/x-yaml")
		data, err := yaml.Marshal(config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, string(data))
	} else {
		c.JSON(http.StatusOK, config)
	}
}

func main() {
	flag.Parse()

	if err := loadTemplate(); err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		return
	}

	r := gin.Default()
	r.GET("/rules", getRules)

	fmt.Printf("Starting server with %d groups and %d rules per group\n", groupCount, ruleCount)
	r.Run(":8080")
}
