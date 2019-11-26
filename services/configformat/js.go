package configformat

import (
	"encoding/json"
)

const (
	ES6_PREFIX = "export default "
)

type JsFormat struct {
	YmlData
}

func (c *JsFormat) Format(data *YmlData, projectName string) (string, error) {
	c.YmlData = *data
	c.ProjectName = projectName
	if err := c.FormatItem(); err != nil {
		return "", err
	}
	bt, err := json.Marshal(c.FormatConfigs)
	if err != nil {
		return "", err
	}
	return ES6_PREFIX + string(bt), nil
}
