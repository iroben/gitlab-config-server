package configformat

import (
	"errors"
	"gitlab-config-server/models"
)

/*
`ignore` 忽略项目名
`prefix` 项目名+配置名
`dot` 项目名+ `.` +连接配置
`tree`  保持层级结构
`project_without_prefix` 当前项目不要加前缀
`project_without_dot` 当前项目不要加`.
`project_without_tree` 当前项目不要层级结构
`js`(es6)
`json`
`ini`
`yml`
*/
const (
	ITEM_IGNORE         = "ignore"
	ITEM_PREFIX         = "prefix"
	ITME_DOT            = "dot"
	ITEM_TREE           = "tree"
	ITEM_WITHOUT_PREFIX = "project_without_prefix"
	ITEM_WITHOUT_DOT    = "project_without_dot"
	ITEM_WITHOUT_TREE   = "project_without_tree"
	FORMAT_JS           = "js"
	FORMAT_JSON         = "json"
	FORMAT_INI          = "ini"
	FORMAT_YML          = "yml"
)

type ConfigFormat interface {
	Format(*YmlData, string) (string, error)
}

type YmlData struct {
	ProjectName string
	Format      string
	Env         string
	ItemFormat  string
	// yml文件中需要的配置信息key名称
	ConfigNames map[string][]string
	// 项目下的所有配置信息
	Configs map[string][]*models.Config
	// 格式化后的结果
	FormatConfigs interface{}
	// project+key的映射，用来检查key是否存在用
	namesMap map[string]bool
}

func (c *YmlData) FormatItem() error {
	c.namesMap = make(map[string]bool)
	for projectName, v := range c.ConfigNames {
		for _, v1 := range v {
			c.namesMap[projectName+v1] = true
		}
	}
	switch c.ItemFormat {
	case ITEM_IGNORE:
		fallthrough
	case ITEM_WITHOUT_PREFIX:
		fallthrough
	case ITEM_PREFIX:
		fallthrough
	case ITME_DOT:
		fallthrough
	case ITEM_WITHOUT_DOT:
		c.Flat()
	case ITEM_WITHOUT_TREE:
		fallthrough
	case ITEM_TREE:
		c.Tree()
	default:
		return errors.New("ItemFormat not implement")
	}
	return nil
}
func (c *YmlData) GetValueByEnv(config *models.Config) string {
	if v, ok := config.Val[c.Env]; ok {
		return v
	}
	return ""
}
func (c *YmlData) Flat() {
	formatData := make(map[string]string, len(c.namesMap))
	for projectName, configs := range c.Configs {
		for _, config := range configs {
			if _, ok := c.namesMap[projectName+config.Key]; !ok {
				continue
			}
			switch c.ItemFormat {
			case ITEM_IGNORE:
				formatData[config.Key] = c.GetValueByEnv(config)
			case ITEM_PREFIX:
				formatData[projectName+config.Key] = c.GetValueByEnv(config)
			case ITEM_WITHOUT_PREFIX:
				if c.ProjectName == projectName {
					formatData[config.Key] = c.GetValueByEnv(config)
					continue
				}
				formatData[projectName+config.Key] = c.GetValueByEnv(config)
			case ITEM_WITHOUT_DOT:
				if c.ProjectName == projectName {
					formatData[config.Key] = c.GetValueByEnv(config)
					continue
				}
				formatData[projectName+"."+config.Key] = c.GetValueByEnv(config)
			case ITME_DOT:
				formatData[projectName+"."+config.Key] = c.GetValueByEnv(config)
			}
		}
	}
	c.FormatConfigs = formatData
}
func (c *YmlData) Tree() {
	formatData := make(map[string]interface{}, len(c.ConfigNames))
	for projectName, configs := range c.Configs {
		projectConfig := make(map[string]string, len(c.ConfigNames[projectName]))
		for _, config := range configs {
			if _, ok := c.namesMap[projectName+config.Key]; !ok {
				continue
			}
			projectConfig[config.Key] = c.GetValueByEnv(config)
		}
		if c.ItemFormat == ITEM_WITHOUT_TREE && c.ProjectName == projectName {
			for k, v := range projectConfig {
				formatData[k] = v
			}
			continue
		}
		formatData[projectName] = projectConfig
	}
	c.FormatConfigs = formatData
}

/**
获取格式化数据
*/
func Format(data *YmlData, projectName string) (string, error) {
	var convert ConfigFormat
	switch data.Format {
	case FORMAT_JS:
		convert = &JsFormat{}
	case FORMAT_JSON:
		convert = &JsonFormat{}
	case FORMAT_INI:
		convert = &IniFormat{}
	case FORMAT_YML:
		convert = &YmlFormat{}
	default:
		return "", errors.New("Format not implement")
	}
	return convert.Format(data, projectName)
}
