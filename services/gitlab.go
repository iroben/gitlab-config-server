package services

import (
	"encoding/json"
	"gitlab-config-server/config"
	"gitlab-config-server/helper"
	"gitlab-config-server/models"
	"io/ioutil"
	"log"
	"strconv"
)

type GitLab struct {
	Domain string
	Token  string
}

var (
	client = helper.Client("/api/v4/projects")
)

type ProjectBranchTag struct {
	Name string `json:"name"`
}

func (g GitLab) Branches(projectId string) []string {

	branchesClient := helper.Client("/api/v4/projects/branches")
	branchResp, err := branchesClient.Get(g.Domain + "/api/v4/projects/" + projectId + "/repository/branches?access_token=" + g.Token)
	if err != nil {
		return nil
	}
	defer branchResp.Body.Close()
	bt, _ := ioutil.ReadAll(branchResp.Body)
	var branches []ProjectBranchTag
	if err := json.Unmarshal(bt, &branches); err != nil {
		return nil
	}
	retVal := make([]string, len(branches))

	for i, branch := range branches {
		retVal[i] = branch.Name
	}
	return retVal

}

func (g GitLab) Tags(projectId string) []string {
	tagsClient := helper.Client("/api/v4/projects/tags")
	tagResp, err := tagsClient.Get(g.Domain + "/api/v4/projects/" + projectId + "/repository/tags?access_token=" + g.Token)
	if err != nil {
		return nil
	}
	defer tagResp.Body.Close()
	bt, _ := ioutil.ReadAll(tagResp.Body)
	var tags []ProjectBranchTag
	if err := json.Unmarshal(bt, &tags); err != nil {
		return nil
	}
	retVal := make([]string, len(tags))

	for i, tag := range tags {
		retVal[i] = tag.Name
	}
	return retVal
}
func (g GitLab) Project(projectId int) *models.GitlabProject {
	resp, err := client.Get(g.Domain + "/api/v4/projects/" + strconv.Itoa(projectId) + "?access_token=" + g.Token)

	if err != nil {
		log.Println("请求项目数据失败:", err)
		return nil
	}
	defer resp.Body.Close()
	bt, _ := ioutil.ReadAll(resp.Body)
	var project models.GitlabProject
	if err := json.Unmarshal(bt, &project); err != nil {
		log.Print("项目信息解析失败: ", err)
		return nil
	}
	log.Printf("%+v\n", project)
	return &project
}

func (g GitLab) Projects() {

	page := 1
	projects := make([]*models.GitlabProject, 0)
	for {

		log.Println("请求第" + strconv.Itoa(page) + "页项目信息")
		resp, err := client.Get(g.Domain + "/api/v4/projects?page=" +
			strconv.Itoa(page) + "&access_token=" + g.Token)

		if err != nil {
			log.Println("请求项目数据失败:", err)
			return
		}

		bt, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		// log.Println(string(bt))
		var gitlabProjects []*models.GitlabProject
		if err := json.Unmarshal(bt, &gitlabProjects); err != nil {
			log.Println("项目信息解析失败: ", err)
			return
		}
		// 没有项目信息了
		if len(gitlabProjects) == 0 {
			log.Println("没有项目信息了: ", page)
			break
		}
		go func(_gitlabProjects []*models.GitlabProject) {
			for _, project := range _gitlabProjects {
				projectId := strconv.Itoa(project.Id)
				project.Branches = g.Branches(projectId)
				project.Tags = g.Tags(projectId)
				project.Group = project.NameSpace.Name
				project.Save()
			}
		}(gitlabProjects)
		projects = append(projects, gitlabProjects...)
		page += 1
	}

	user := (&models.User{}).GetUserInfo(g.Token)
	if user == nil {
		log.Println("ERROR GetUserInfo fail")
		return
	}

	models.UserProject{
		UserId: user.Id,
	}.SaveProjects(projects, config.GetString("GitLabToken") == g.Token)

}
