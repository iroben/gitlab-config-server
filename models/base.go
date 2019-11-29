package models

import (
	"log"
	"gitlab-config-server/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"sync"
)

type BaseModel struct {
	//gorm.Model
}

/**
操作权限
 */
type Operation struct {
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

var (
	DB             *gorm.DB
	GitLabProjects map[int]Project
	lock           sync.Mutex
)

func init() {
	dbConfig := config.GetDbConfig()
	var err error
	DB, err = gorm.Open("mysql",
		dbConfig.User+":"+dbConfig.Passwd+"@tcp("+dbConfig.Host+")/"+dbConfig.Name+"?multiStatements=true&charset=utf8mb4&loc=Asia%2FShanghai")
	if err != nil {
		log.Fatalln("数据库连接创建失败", err.Error())
	}
	if config.ENV != "PROD" {
		DB.LogMode(true)
	}
	GitLabProjects = make(map[int]Project)
	var project GitlabProject
	project.CacheProjects(project.FindAll())
}
func (m BaseModel) CacheProjects(projects []*GitlabProject) {
	// 如果是GitLab管理员创建的token，则缓存一份项目信息，用于配置模型用
	lock.Lock()
	defer lock.Unlock()
	for _, project := range projects {
		GitLabProjects[project.Id] = Project{
			Id:          project.Id,
			Name:        project.Name,
			Description: project.Description,
		}
	}
}
func (m *BaseModel) Error(e error) {
	log.Println("DB ERROR:", e.Error())
}
func (m *BaseModel) Info(msg string) {
	log.Println("DB INFO:", msg)
}
