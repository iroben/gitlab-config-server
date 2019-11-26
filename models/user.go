package models

import (
	"gitlab-config-server/helper"
	"gitlab-config-server/config"
	"encoding/json"
	"log"
)

type User struct {
	Id        int    `json:"id"`
	IsAdmin   bool   `json:"is_admin"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	ClientIp  string `json:"client_ip"`
}

func (m *User) GetUserInfo(accessToken string) *User {

	path := "/api/v4/user"
	client := helper.Client(path)
	resp, err := client.Get(config.GetString("GitLabDomain") + path + "?access_token=" + accessToken)
	if err != nil {
		log.Println("ERROR: 获取用户出错：", err.Error())
		return nil
	}
	defer resp.Body.Close()
	decode := json.NewDecoder(resp.Body)
	var user User
	if err := decode.Decode(&user); err != nil {
		log.Println("ERROR: 解析用户数据失败：", err.Error())
		return nil
	}
	return &user
}
