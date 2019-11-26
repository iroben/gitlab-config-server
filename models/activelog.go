package models

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
)

const (
	ACTIVE_UPDATE = "update"
	ACTIVE_DELETE = "delete"
)

type ActiveLog struct {
	BaseModel
	Id      int                    `gorm:"column:id;primary_key"`
	Time    int                    `gorm:"column:time"`
	Uid     int                    `gorm:"column:uid"`
	Ip      string                 `gorm:"column:ip"`
	Who     string                 `gorm:"column:who"`
	Action  string                 `gorm:"column:action"`
	DataStr string                 `gorm:"column:data" json:"Data"`
	Data    map[string]interface{} `gorm:"-" json:"-"`
}

func (m *ActiveLog) TableName() string {
	return "action_log"
}

func (m *ActiveLog) Marshal() {
	if m.Data != nil {
		if bt, err := json.Marshal(m.Data); err == nil {
			m.DataStr = string(bt)
		}
	}
}

func (m *ActiveLog) UnMarshal() {
	if m.DataStr != "" {
		var val map[string]interface{}
		if err := json.Unmarshal([]byte(m.DataStr), &val); err == nil {
			m.Data = val
		}
	}
}

func (m *ActiveLog) Add() bool {
	m.Marshal()
	err := DB.Create(&m).Error
	if err != nil {
		m.Error(err)
		return false
	}
	return true
}

func (m ActiveLog) Find(page int) ([]*ActiveLog, int) {
	var retVal []*ActiveLog
	pageSize := 20
	err := DB.Order("time DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&retVal).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.Error(err)
		return nil, 0
	}
	for _, record := range retVal {
		record.UnMarshal()
	}
	total := 0
	DB.Model(ActiveLog{}).Count(&total)

	return retVal, total
}
