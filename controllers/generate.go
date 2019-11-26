package controllers

import (
	"gitlab-config-server/models"
	"gitlab-config-server/services/configformat"
	"sort"

	"github.com/go-yaml/yaml"
	"gitlab-config-server/config"
)

type GenerateConfig struct {
	MainController
}

func ParseYmlData(data *YmlConfig) (*configformat.YmlData, error) {
	retVal := configformat.YmlData{
		Format:      data.Format,
		Env:         data.Env,
		ItemFormat:  data.ItemFormat,
		ConfigNames: map[string][]string{},
	}
	for key, value := range data.Configs {

		if _, ok := retVal.ConfigNames[key]; !ok {
			retVal.ConfigNames[key] = make([]string, 0)
		}
		for _, configName := range value {
			retVal.ConfigNames[key] = append(retVal.ConfigNames[key], configName)
		}
	}
	return &retVal, nil
}

type YmlProject struct {
	Id          int    `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}
type YmlConfig struct {
	Format     string              `yaml:"format"`
	ItemFormat string              `yaml:"itemFormat"`
	Configs    map[string][]string `yaml:"configs"`
	Project    YmlProject          `yaml:"project"`
	Env        string              `yaml:"branch"`
}

/**
根据yml文件生成配置文件
*/
func (c *GenerateConfig) Generate() {
	if c.GetString("token") != config.GetString("apiToken") {
		c.Error("token验证失败", "")
		c.Ctx.ResponseWriter.WriteHeader(500)
		return
	}
	var val YmlConfig
	c.Info("ymldata ", string(c.Ctx.Input.RequestBody))
	if err := yaml.Unmarshal(c.Ctx.Input.RequestBody, &val); err != nil {
		c.Error("yml数据解析失败：", err.Error())
		c.Ctx.ResponseWriter.WriteHeader(500)
		return
	}
	gitlabModel := models.GitlabProject{
		Id:          val.Project.Id,
		Name:        val.Project.Name,
		Description: val.Project.Description,
	}
	gitlab, result := gitlabModel.Save()
	if ! result {
		c.Error("gitlab信息添加失败", "")
		c.Ctx.ResponseWriter.WriteHeader(500)
		return
	}
	branch := gitlab.GetBranch(val.Env)
	if branch == "" {
		c.Error("分支未找到：", val.Env)
		c.Ctx.ResponseWriter.WriteHeader(500)
		return
	}
	val.Env = branch
	ymlData, err := ParseYmlData(&val)
	if err != nil {
		c.Error("yml数据解析失败：", err.Error())
		c.Ctx.ResponseWriter.WriteHeader(500)
		return
	}
	var model models.Config
	configs := make(map[string][]*models.Config, 0)
	for project, cf := range ymlData.ConfigNames {
		/**
		输出格式和环境不解析
		*/
		_configs := model.GetProjectConfig(cf, &gitlabModel, project)
		_configSlice := ConfigSlice(_configs)
		sort.Sort(_configSlice)
		configs[project] = _configSlice
	}
	ymlData.Configs = configs
	data, err := configformat.Format(ymlData, val.Project.Name)
	if err != nil {
		c.Error("yml数据解析失败：", err.Error())
		c.Ctx.ResponseWriter.WriteHeader(500)
		return
	}
	c.Ctx.WriteString(data)
	return
}
