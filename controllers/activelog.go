package controllers

import (
	"strconv"
	"gitlab-config-server/models"
)

type ActiveLog struct {
	MainController
}

func (c ActiveLog) Get() {
	pageStr := c.GetString("page", "1")
	page, _ := strconv.Atoi(pageStr)
	data, total := (models.ActiveLog{}).Find(page)
	c.Json(map[string]interface{}{
		"statusCode": RESP_OK,
		"total":      total,
		"data":       data,
	})
}
