package models

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"time"
)

/**
精简版Gitlab项目信息
 */
type Project struct {
	Id          int
	Name        string
	Description string
}

type Config struct {
	BaseModel
	Id                int               `gorm:"column:id;primary_key"`
	ProjectId         int               `gorm:"column:project_id"`
	Key               string            `gorm:"column:key"`
	Valstr            string            `gorm:"column:val" json:"-"`
	Val               map[string]string `gorm:"-"`
	Description       string            `gorm:"column:description"`
	DependentStr      string            `gorm:"column:dependent" json:"-"`
	Dependent         []int             `gorm:"-" json:"-"`
	DependentProjects []Project
}

func (m *Config) TableName() string {
	return "config"
}
func (m *Config) Marshal() {
	if m.Val != nil {
		if bt, err := json.Marshal(m.Val); err == nil {
			m.Valstr = string(bt)
		}
	}
	if m.Dependent != nil {
		if bt, err := json.Marshal(m.Dependent); err == nil {
			m.DependentStr = string(bt)
		}
	}
}
func (m *Config) UnMarshal() {
	if m.Valstr != "" {
		var val map[string]string
		if err := json.Unmarshal([]byte(m.Valstr), &val); err == nil {
			m.Val = val
		}
	}
	if m.DependentStr != "" {
		var val []int
		if err := json.Unmarshal([]byte(m.DependentStr), &val); err == nil {
			m.Dependent = val
		}
	}
}
func (m *Config) Add(isSave bool) bool {
	val := Config{}
	err := DB.Where("project_id=? and `key`=?", m.ProjectId, m.Key).First(&val).Error
	if err == nil && val.Id != 0 {
		m.Info("配置已经存在：" + m.Key)
		if isSave {
			return val.UpdateConfigEnv(m)
		}
		return false
	}
	if err != gorm.ErrRecordNotFound {
		m.Error(err)
		return false
	}
	m.Marshal()
	err = DB.Create(m).Error
	if err != nil {
		m.Error(err)
		return false
	}
	return true
}

/**
更新某个环境的配置
*/
func (m Config) UpdateConfigEnv(config *Config) bool {
	m.UnMarshal()
	for k, v := range config.Val {
		m.Val[k] = v
	}
	config.Val = m.Val
	m.Marshal()
	db := DB.Model(m).
		Where("id=?", m.Id)
	columns := make(map[string]string)
	columns["val"] = m.Valstr
	if config.Description != "" {
		columns["description"] = config.Description
	}
	err := db.UpdateColumns(columns).Error
	if err != nil {
		m.Error(err)
		return false
	}
	return true
}

func (m Config) FindByIds(ids []int) []*Config {
	var retVal []*Config
	if err := DB.Where("id IN (?)", ids).Find(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	for _, val := range retVal {
		val.UnMarshal()
	}
	return retVal
}

/**
通过主键获取配置信息
*/
func (m *Config) Find() *Config {
	var retVal Config
	if err := DB.Where("id=?", m.Id).First(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	retVal.UnMarshal()
	return &retVal
}
func (m Config) FindAll() []*Config {
	var retVal []*Config
	if err := DB.Preload("Project").Find(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	for _, v := range retVal {
		v.UnMarshal()
	}
	return retVal
}

/**
更新配置信息，并返回更新后的配置结果
*/
func (m *Config) UpdateAndReturn(user *User) *Config {
	if !m.Update(user) {
		return nil
	}
	return m.Find()
}

// 记录日志
func (old *Config) Log(new Config, user *User, activeType string) {

	if user.Id == 0 {
		return
	}
	data := make(map[string]interface{})
	data["before"] = old
	data["after"] = new
	(&ActiveLog{
		Who:    user.Name,
		Uid:    user.Id,
		Data:   data,
		Time:   int(time.Now().Unix()),
		Action: activeType,
		Ip:     user.ClientIp,
	}).Add()
}

func (m *Config) Update(user *User) bool {
	oldConfig := m.Find()

	if oldConfig == nil {
		return false
	}
	result := oldConfig.UpdateConfigEnv(m)
	if result {
		oldConfig.Log(*m, user, ACTIVE_UPDATE)
	}
	return result
}
func (m *Config) Delete(user *User) bool {
	old := m.Find()
	old.Log(Config{}, user, ACTIVE_DELETE)

	err := DB.Where("id=?", m.Id).Delete(Config{}).Error
	if err != nil {
		m.Error(err)
		return false
	}
	return true
}
func formatConfig(configs []*Config) {
	for _, v := range configs {
		v.UnMarshal()
		for _, pid := range v.Dependent {
			if project, ok := GitLabProjects[pid]; ok {
				v.DependentProjects = append(v.DependentProjects, project)
			}

		}
	}
}
func (m *Config) GetConfigByProjectId() []*Config {
	var retVal []*Config
	if m.ProjectId == 0 {
		return nil
	}
	err := DB.Where("project_id=?", m.ProjectId).Find(&retVal).Error
	if err != nil {
		m.Error(err)
		return nil
	}
	formatConfig(retVal)
	return retVal
}

/**
将slice转换成map
*/
func buildSliceToMap(items []string) map[string]bool {
	retVal := make(map[string]bool)
	for _, v := range items {
		retVal[v] = true
	}
	return retVal
}

func (m *Config) GetProjectConfig(
	configNames []string,
	gitlabProject *GitlabProject,
	dependentProjectName string) []*Config {

	project := (GitlabProject{
		Name: dependentProjectName,
	}).GetProjectByName()
	if project == nil {
		return nil
	}

	m.ProjectId = project.Id
	configs := m.GetConfigByProjectId()
	if configs == nil {
		return nil
	}
	configNameMap := buildSliceToMap(configNames)
	var retVal []*Config
	for _, config := range configs {
		if _, ok := configNameMap[config.Key]; !ok {
			continue
		}
		config.AddDependent(gitlabProject)
		retVal = append(retVal, config)
	}
	return retVal
}

func (m *Config) UpdateDependent() bool {
	m.Marshal()
	err := DB.Model(m).
		Where("id=?", m.Id).
		UpdateColumn("dependent", m.DependentStr).Error
	if err != nil {
		m.Error(err)
		return false
	}
	return true
}

/**
添加配置依赖项目
*/
func (m *Config) AddDependent(gitlabProject *GitlabProject) {
	for _, projectId := range m.Dependent {
		if projectId == gitlabProject.Id {
			m.Info("dependent exists")
			return
		}
	}
	m.Dependent = append(m.Dependent, gitlabProject.Id)
	if m.UpdateDependent() {
		m.Info("update dependent success")
		return
	}
	m.Info("update dependent fail")
}
