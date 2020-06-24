package controllers

import (
	"encoding/json"
	"fmt"
	"gitlab-config-server/config"
	"gitlab-config-server/helper"
	"gitlab-config-server/models"
	"gitlab-config-server/services"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"
)

type GitLabController struct {
	MainController
}

/**
获取项目信息
*/
func (c *GitLabController) Projects() {
	if c.User.IsAdmin {
		c.Json(map[string]interface{}{
			"statusCode": RESP_OK,
			"data":       models.GitlabProject{}.FindAll(),
		})
		return
	}
	c.Json(map[string]interface{}{
		"statusCode": RESP_OK,
		"data": models.UserProject{
			UserId: c.User.Id,
		}.UserProjects(),
	})

}

func (c *GitLabController) Login() {
	log.Println(GitLabDomain +
		fmt.Sprintf("/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=api",
			config.GetString("GitLabClientId"),
			url.QueryEscape(config.GetString("domain")+"/gitlab/callback")))

	c.Redirect(GitLabDomain+
		fmt.Sprintf("/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=api",
			config.GetString("GitLabClientId"),
			url.QueryEscape(config.GetString("domain")+"/gitlab/callback")), 302)
}

func (c *GitLabController) Callback() {
	code := c.GetString("code")
	path := "/oauth/token"
	client := helper.Client(path)
	params := url.Values{}
	params.Add("client_id", config.GetString("GitLabClientId"))
	params.Add("code", code)
	params.Add("client_secret", config.GetString("GitLabClientSecret"))
	params.Add("grant_type", "authorization_code")
	params.Add("redirect_uri", config.GetString("domain")+"/gitlab/callback")
	resp, err := client.Post(GitLabDomain+path, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		c.Redirect(config.GetString("gitlab-config-web.domain")+"?msg=get gitlab user info error", 302)
		log.Println("err: ", err)
		return
	}
	defer resp.Body.Close()
	bt, _ := ioutil.ReadAll(resp.Body)
	var oauth OAuthResp
	if err := json.Unmarshal(bt, &oauth); err != nil {
		log.Println("decode fail: ", err)
		c.Redirect(config.GetString("gitlab-config-web.domain")+"?msg=get gitlab user info error", 302)
		return
	}
	if oauth.AccessToken == "" {
		c.Redirect(config.GetString("gitlab-config-web.domain")+"?msg=get gitlab user info error", 302)
		return
	}
	token := helper.GetGuid()
	c.AccessToken = token
	c.GitLabAccessToken = oauth.AccessToken
	user := c.GetUserInfo()
	if user == nil {
		log.Println("get gitlab user info error")
		c.Redirect(config.GetString("gitlab-config-web.domain")+"?msg=get gitlab user info error", 302)
		return
	}

	// 登录成功后同步一次gitlab信息
	go SyncGitlab(c.GitLabAccessToken)

	user.ClientIp = c.Ctx.Input.IP()
	userbt, _ := json.Marshal(user)
	services.Redis.Set("sid:"+token, string(userbt), time.Second*86400)
	services.Redis.Set("sid:"+token+":gitlab", string(bt), time.Second*86400)
	c.Redirect(config.GetString("gitlab-config-web.domain")+"?token="+token, 302)
}
