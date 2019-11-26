package controllers

import (
	"encoding/json"
	"gitlab-config-server/models"
	"sort"
	"strconv"

	"github.com/go-yaml/yaml"
)

type Config struct {
	MainController
}

func (c *Config) Get() {
	pidStr := c.Input().Get("pid")
	pid, _ := strconv.Atoi(pidStr)
	model := models.Config{
		ProjectId: pid,
	}
	configs := model.GetConfigByProjectId()
	configSlice := ConfigSlice(configs)
	sort.Sort(configSlice)
	hasPermission := false
	if len(configs) > 0 {
		hasPermission = c.CheckProjectPermission(configs[0].ProjectId)
	}

	c.Json(map[string]interface{}{
		"statusCode":    RESP_OK,
		"data":          configSlice,
		"hasPermission": hasPermission,
	})
}

type PostKey struct {
	Key         string            `json:"key"`
	Val         map[string]string `json:"val"`
	Description string            `json:"description"`
}

type PostConfig struct {
	Id        int       `json:"id"`
	ProjectId int       `json:"project_id"`
	Configs   []PostKey `json:"configs"`
}

func (c *Config) Post() {
	var params []PostConfig
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &params); err != nil {
		c.Error("配置数据解析失败：", err.Error())
		c.Json(map[string]interface{}{
			"statusCode": RESP_PARAM_ERR,
		})
		return
	}
	msg := "添加成功"
	for _, project := range params {
		for _, cf := range project.Configs {
			model := models.Config{
				ProjectId:   project.ProjectId,
				Key:         cf.Key,
				Description: cf.Description,
				Val:         cf.Val,
			}
			if !model.Add(false) {
				msg = "部分数据添加失败"
			}
		}
	}
	c.Json(map[string]interface{}{
		"statusCode": RESP_OK,
		"msg":        msg,
	})
}

func (c *Config) Put() {
	updateDependent := c.Input().Get("dependent") == "1"
	var params map[int]PostKey
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &params); err != nil {
		c.Error("配置数据解析失败：", err.Error())
		c.Json(map[string]interface{}{
			"statusCode": RESP_OK,
		})
		return
	}
	msg := "更新成功"

	// 保存更新后配置信息，用于创建gitlab pipeline用的
	configs := make([]MidifyConfigs, 0)

	for cid, cf := range params {
		model := models.Config{
			Id:          cid,
			Description: cf.Description,
			Val:         cf.Val,
		}
		config := model.Find()
		if config == nil {
			c.Json(map[string]interface{}{
				"statusCode": RESP_ERR,
			})
			return
		}
		if !c.CheckProjectPermission(config.ProjectId) {
			c.Json(map[string]interface{}{
				"statusCode": RESP_NO_ACCESS,
			})
			return
		}
		if updateDependent {
			branches := make([]string, 0)
			hasTag := false
			for branch, _ := range cf.Val {
				if branch != "tag" {
					branches = append(branches, branch)
				} else {
					hasTag = true
				}
			}
			configs = append(configs, MidifyConfigs{
				ConfigId: cid,
				Branches: branches,
				HasTag:   hasTag,
			})
		}
		if !model.Update(c.User) {
			msg = "部分数据更新失败"
			continue
		}
	}
	/**
	如果更新依赖启动了，把相关配置更新的信息发送到configsChan通道，
	单独一个携程处理pipeline更新请求，不阻塞用户请求
	*/
	if updateDependent {
		configsChan <- configs
	}
	c.Json(map[string]interface{}{
		"statusCode": RESP_OK,
		"msg":        msg,
	})
}

func (c *Config) Delete() {
	cid, err := strconv.Atoi(c.Input().Get("id"))
	if err != nil {
		c.Error("删除配置参数解析失败：", err.Error())
		c.Json(map[string]interface{}{
			"statusCode": RESP_PARAM_ERR,
		})
		return
	}
	model := models.Config{
		Id: cid,
	}
	config := model.Find()
	if config == nil {
		c.Json(map[string]interface{}{
			"statusCode": RESP_ERR,
		})
		return
	}
	if !c.CheckProjectPermission(config.ProjectId) {
		c.Json(map[string]interface{}{
			"statusCode": RESP_NO_ACCESS,
		})
		return
	}
	if model.Delete(c.User) {
		c.Json(map[string]interface{}{
			"statusCode": RESP_OK,
		})
		return
	}
	c.Json(map[string]interface{}{
		"statusCode": RESP_ERR,
	})
}

type PostYmlConfig struct {
	ProjectId int    `form:"project_id"`
	Yml       string `form:"yml"`
}

/**
从yml文件中导入
*/
func (c *Config) Yml() {
	var params PostYmlConfig
	if err := c.ParseForm(&params); err != nil {
		c.Error("yml解析失败", err.Error())
		c.Json(map[string]interface{}{
			"statusCode": RESP_PARAM_ERR,
		})
		return
	}
	if params.ProjectId == 0 || params.Yml == "" {
		c.Json(map[string]interface{}{
			"statusCode": RESP_PARAM_ERR,
		})
		return
	}
	var yamlData map[string]map[string]string
	err := yaml.Unmarshal([]byte(params.Yml), &yamlData)
	if err != nil {
		c.Error("yml参数解析失败", err.Error())
		c.Json(map[string]interface{}{
			"statusCode": RESP_PARAM_ERR,
			"msg":        "yml参数解析失败",
		})
		return
	}
	envConfigs := make(map[string]map[string]string)
	for env, configs := range yamlData {
		for k, v := range configs {
			if _, ok := envConfigs[k]; !ok {
				envConfigs[k] = make(map[string]string)
			}
			envConfigs[k][env] = v
		}
	}
	msg := "添加成功"
	for k, configs := range envConfigs {
		model := models.Config{
			ProjectId: params.ProjectId,
			Key:       k,
			Val:       configs,
		}
		if !model.Add(true) {
			msg = "部分数据添加失败"
		}
	}
	c.Json(map[string]interface{}{
		"statusCode": RESP_OK,
		"msg":        msg,
	})
}
