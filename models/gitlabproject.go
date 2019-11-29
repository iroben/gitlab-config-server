package models

import (
	"github.com/jinzhu/gorm"
	"encoding/json"
)

const (
	GuestAccess      = 10
	ReporterAccess   = 20
	DeveloperAccess  = 30
	MaintainerAccess = 40
	OwnerAccess      = 50
)

type GitlabProjectAccess struct {
	AccessLevel       int `json:"access_level"`
	NotificationLevel int `json:"notification_level"`
}
type GitlabPermissions struct {
	GroupAccess   GitlabProjectAccess `json:"group_access"`
	ProjectAccess GitlabProjectAccess `json:"project_access"`
}

func (g GitlabPermissions) CheckAccess() bool {
	return g.ProjectAccess.AccessLevel >= MaintainerAccess ||
		g.GroupAccess.AccessLevel >= MaintainerAccess
}

type ProjectNameSpace struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type GitlabProject struct {
	BaseModel
	Id          int               `gorm:"column:id"`
	Name        string            `gorm:"column:name"`
	Description string            `gorm:"column:description"`
	Group       string            `gorm:"column:group"`
	BranchesStr string            `gorm:"column:branches" json:"-"`
	TagsStr     string            `gorm:"column:tags" json:"-"`
	Branches    []string          `gorm:"-"`
	Tags        []string          `gorm:"-"`
	Permissions GitlabPermissions `gorm:"-" json:"permissions"`
	NameSpace   ProjectNameSpace  `gorm:"-" json:"namespace"`
}

func (m *GitlabProject) TableName() string {
	return "gitlab_project"
}

func (m GitlabProject) FindAll() []*GitlabProject {
	var retVal []*GitlabProject
	if err := DB.Find(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	for _, val := range retVal {
		val.UnMarshal()
	}
	return retVal
}

func (m GitlabProject) FindByIds(projectIds []int) []*GitlabProject {
	var retVal []*GitlabProject
	if err := DB.Where("id IN (?)", projectIds).Find(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	for _, val := range retVal {
		val.UnMarshal()
	}
	return retVal
}

func (m *GitlabProject) Find() *GitlabProject {
	var retVal GitlabProject
	if err := DB.Where("id=?", m.Id).First(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	retVal.UnMarshal()
	return &retVal
}
func (m *GitlabProject) Update() bool {
	if err := DB.Model(m).Omit("id").UpdateColumns(m).Error; err != nil {
		m.Error(err)
		return false
	}
	return true
}
func (m *GitlabProject) Marshal() {
	if m.Branches != nil && len(m.Branches) != 0 {
		bt, _ := json.Marshal(m.Branches)
		m.BranchesStr = string(bt)
	}
	if m.Tags != nil && len(m.Tags) != 0 {
		bt, _ := json.Marshal(m.Tags)
		m.TagsStr = string(bt)
	}

}
func (m *GitlabProject) UnMarshal() {
	if m.BranchesStr != "" {
		var val []string
		json.Unmarshal([]byte(m.BranchesStr), &val)
		m.Branches = val

	}
	if m.TagsStr != "" {
		var val []string
		json.Unmarshal([]byte(m.TagsStr), &val)
		m.Tags = val
	}
}
func (m *GitlabProject) Save() (*GitlabProject, bool) {
	if m.Name == "" || m.Id == 0 {
		return nil, false
	}
	var retVal GitlabProject
	if err := DB.Where("id=?", m.Id).First(&retVal).Error; err != nil &&
		err != gorm.ErrRecordNotFound {
		m.Error(err)
		return nil, false
	}
	if retVal.Id != 0 {
		retVal.UnMarshal()
		return &retVal, true
	}
	m.Marshal()
	if err := DB.Create(m).Error; err != nil {
		m.Error(err)
		return nil, false
	}
	return m, true
}

func checkTag(m *GitlabProject, retVal GitlabProject) {
	newTag := make([]string, 0)
	for _, tag := range m.Tags {
		exists := false
		for _, oldTag := range retVal.Tags {
			if oldTag == tag {
				exists = true
				break
			}
		}
		if !exists {
			newTag = append(newTag, tag)
		}
	}
	if len(newTag) >= 0 {
		m.Tags = append(m.Tags, newTag...)
	}
}
func checkBranch(m *GitlabProject, retVal GitlabProject) []string {
	newBranch := make([]string, 0)
	for _, branch := range m.Branches {
		exists := false
		for _, oldBranch := range retVal.Branches {
			if oldBranch == branch {
				exists = true
				break
			}
		}
		if !exists {
			newBranch = append(newBranch, branch)
		}
	}
	if len(newBranch) >= 0 {
		m.Branches = append(m.Branches, newBranch...)
	}
	return newBranch
}

func (m GitlabProject) GetProjectByName() *GitlabProject {
	var retVal GitlabProject
	if err := DB.Where("name=?", m.Name).First(&retVal).Error; err != nil {
		m.Error(err)
		return nil
	}
	retVal.UnMarshal()
	return &retVal
}

/**
根据.config.sh传过来的branch寻找对应的分支，主要是处理tag的，如果branch是tag，则返回tag字符串
 */
func (m GitlabProject) GetBranch(branch string) string {
	for _, _branch := range m.Branches {
		if _branch == branch {
			return branch
		}
	}
	for _, _tag := range m.Tags {
		if _tag == branch {
			return "tag"
		}
	}
	return ""
}
