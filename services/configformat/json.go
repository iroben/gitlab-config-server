package configformat

import "encoding/json"

type JsonFormat struct {
	YmlData
}

func (c *JsonFormat) Format(data *YmlData, projectName string) (string, error) {
	c.YmlData = *data
	c.ProjectName = projectName
	if err := c.FormatItem(); err != nil {
		return "", err
	}
	bt, err := json.Marshal(c.FormatConfigs)
	if err != nil {
		return "", err
	}
	return string(bt), nil
}
