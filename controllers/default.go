package controllers

import (
	"encoding/json"
	"net/http"
	"gitlab-config-server/config"
	"gitlab-config-server/helper"
	"gitlab-config-server/models"
	"strconv"

	"github.com/astaxie/beego"
	"bytes"
	"gitlab-config-server/services"
	"log"
	"io/ioutil"
	"time"
)

const (
	//公共响应码
	RESP_OK            = 10000
	RESP_ERR           = 10001
	RESP_PARAM_ERR     = 10002
	RESP_TOKEN_ERR     = 10003
	RESP_NO_ACCESS     = 10004
	RESP_APP_NOT_ON    = 10005
	RESP_USER_NOT_BIND = 10007
)

var (
	RESP_MSG = map[int]string{
		RESP_OK:         "成功",
		RESP_ERR:        "失败,未知错误",
		RESP_PARAM_ERR:  "参数错误",
		RESP_TOKEN_ERR:  "token错误",
		RESP_NO_ACCESS:  "没有访问权限",
		RESP_APP_NOT_ON: "暂时未提供服务",
	}
	configsChan  chan []MidifyConfigs
	GitLabDomain string
	GitLabToken  string
)

type MidifyConfigs struct {
	ConfigId int ``
	Branches []string
	HasTag   bool
}

func init() {
	configsChan = make(chan []MidifyConfigs, 40)
	GitLabDomain = config.GetString("GitLabDomain")
	GitLabToken = config.GetString("GitLabToken")
	go func() {
		for configs := range configsChan {
			updateDependentProject(configs)
		}
	}()

	gitLabSyncTTL := config.GetInt("GitLabSyncTTL")
	if gitLabSyncTTL == 0 {
		gitLabSyncTTL = 5
	}

	go func() {
		ticker := time.NewTicker(time.Duration(gitLabSyncTTL) * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			// 5分钟同步一次Gitlab项目信息
			log.Println("SyncGitlab ......")
			SyncGitlab(GitLabToken)
		}
	}()
}

/**
更新依赖配置的项目（这里是创建一个gitlab pileline）
*/
func updateDependentProject(configs []MidifyConfigs) {
	if len(configs) == 0 {
		return
	}
	projects := make(map[int][]MidifyConfigs)
	configIds := make([]int, 0)

	// 配置ID和已经修改配置MAP
	configMap := make(map[int]MidifyConfigs)
	for _, config := range configs {
		configIds = append(configIds, config.ConfigId)
		configMap[config.ConfigId] = config
	}
	var configProject *models.GitlabProject
	mconfigs := models.Config{}.FindByIds(configIds)
	for _, mconfig := range mconfigs {
		//配置所有项目信息只加载一次，因为这些配置都是在该项目下
		if configProject == nil {
			configProject = (&models.GitlabProject{
				Id: mconfig.ProjectId,
			}).Find()
		}
		for _, project := range mconfig.Dependent {
			if _, ok := projects[project]; !ok {
				projects[project] = make([]MidifyConfigs, 0)
			}
			projects[project] = append(projects[project], configMap[mconfig.Id])
		}
	}

	if configProject == nil {
		return
	}

	for projectId, configs := range projects {
		modifyBranches := make(map[string]bool)
		hasTag := false
		for _, config := range configs {
			for _, branch := range config.Branches {
				modifyBranches[branch] = true

			}
			if !hasTag {
				hasTag = config.HasTag
			}
		}
		for branch, _ := range modifyBranches {
			CreatePipeline(projectId, branch)
		}
		if !hasTag || len(configProject.Tags) == 0 {
			continue
		}
		/**
		如果tag更新了，默认更新依赖项目的最新tag，如果要优先更新同名tag，把下面这段注释取消了
		 */
		/*if CreatePipeline(projectId, configProject.Tags[0]) {
			// 如果tag值修改了，优先更新该配置所在项目的最新tag的同名tag，成功就返回
			continue
		}*/
		// 如果失败，则更新依赖项目的最新tag
		project := (&models.GitlabProject{
			Id: projectId,
		}).Find()
		if project == nil || len(project.Tags) == 0 {
			continue
		}
		CreatePipeline(projectId, project.Tags[0])
	}
}

type GitLabPipeline struct {
	Id     int    `json:id`
	Sha    string `json:sha`
	Ref    string `json:ref`
	Status string `json:status`
}

// 调用gitlab create pipeline接口
func CreatePipeline(projectId int, ref string) bool {
	if projectId == 0 || ref == "" {
		return false
	}
	client := helper.Client("pipeline")
	params := make(map[string]interface{})
	params["ref"] = ref
	params["variables"] = []map[string]string{{"key": "config_trigger", "value": "gitlab_config_server"}}

	bt, _ := json.Marshal(params)

	log.Println("pipeline: ", GitLabDomain+"/api/v4/projects/"+strconv.Itoa(projectId)+"/pipeline?ref="+ref)
	req, err := http.NewRequest("POST", GitLabDomain+"/api/v4/projects/"+strconv.Itoa(projectId)+"/pipeline", bytes.NewReader(bt))
	if err != nil {
		log.Println("ERROR: gitlab连接创建失败", err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Private-Token", GitLabToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR: gitlab连接执行失败：", err.Error())
		return false
	}
	defer resp.Body.Close()
	var gitlabPipeline GitLabPipeline
	bt, _ = ioutil.ReadAll(resp.Body)
	log.Println("pipeline resp: ", string(bt))
	if err := json.Unmarshal(bt, &gitlabPipeline); err != nil {
		log.Println("ERROR: gitlab pipeline失败：", err.Error())
		return false
	}
	log.Println("gitlab pipeline：", gitlabPipeline)
	if gitlabPipeline.Id != 0 && gitlabPipeline.Status == "pending" {
		log.Println("gitlab pipeline create success")
		return true
	}
	log.Println("ERROR: gitlab pipeline create fail：", gitlabPipeline)
	return false
}

type OAuthResp struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreateAt     int    `json:"created_at"`
}

type MainController struct {
	beego.Controller
	AccessToken       string
	GitLabAccessToken string
	GitlabOAuth       *OAuthResp
	User              *models.User
}

//获取用户信息

func (c *MainController) UserInfo() {
	c.Json(map[string]interface{}{
		"statusCode": RESP_OK,
		"data":       c.User,
	})
}
func (c *MainController) GetUserInfo() *models.User {
	if c.User != nil && c.User.Id > 0 {
		return c.User
	}
	c.User = c.User.GetUserInfo(c.GitLabAccessToken)
	return c.User
}

func (c *MainController) Options() {
	c.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	c.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS,PUT,DELETE")
	c.Ctx.WriteString("")
}

func (c *MainController) Prepare() {
	token := c.GetString("token")
	if token != "" {
		c.AccessToken = token
		userStr, _ := services.Redis.Get("sid:" + token).Result()
		gitlabStr, _ := services.Redis.Get("sid:" + token + ":gitlab").Result()
		if userStr != "" {
			var user models.User
			if err := json.Unmarshal([]byte(userStr), &user); err == nil {
				c.User = &user

			}
		}
		if gitlabStr != "" {
			var auth OAuthResp
			if err := json.Unmarshal([]byte(gitlabStr), &auth); err == nil {
				c.GitlabOAuth = &auth
				c.GitLabAccessToken = auth.AccessToken
			}
		}
	}

	if c.User == nil {
		c.User = &models.User{}
	}

}

func (c *MainController) Json(data map[string]interface{}) {
	if _, ok := data["msg"]; !ok {
		v, _ok := data["statusCode"].(int)
		if _ok {
			data["msg"] = RESP_MSG[v]
		}
	}
	c.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	if data["statusCode"].(int) == RESP_PARAM_ERR { //参数错误
		c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
	}
	if data["statusCode"].(int) == RESP_TOKEN_ERR { //token(access token , refresh access token) 错误
		c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
	}

	c.Data["json"] = &data
	c.ServeJSON()
}
func (c *MainController) Info(key, value string) {
	log.Println("INFO: ", key, "\t", value)
}
func (c *MainController) Error(key, value string) {
	log.Println("ERROR: ", key, "\t", value)
}

type ConfigSlice []*models.Config

func (c ConfigSlice) Len() int {
	return len(c)
}
func (c ConfigSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c ConfigSlice) Less(i, j int) bool {
	return c[i].Key < c[j].Key
}

func SyncGitlab(token string) {
	services.GitLab{
		Domain: GitLabDomain,
		Token:  token,
	}.Projects()

}
func (c *MainController) CheckProjectPermission(projectId int) bool {
	hasPermission := false
	project := services.GitLab{
		Token:  c.GitLabAccessToken,
		Domain: GitLabDomain,
	}.Project(projectId)
	if project != nil {
		hasPermission = project.Permissions.CheckAccess()
	}
	return hasPermission
}
