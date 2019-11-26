package configformat

import "github.com/go-yaml/yaml"

type YmlFormat struct {
	YmlData
}

func (c *YmlFormat) Format(data *YmlData, projectName string) (string, error) {
	c.YmlData = *data
	c.ProjectName = projectName
	if err := c.FormatItem(); err != nil {
		return "", err
	}
	bt, err := yaml.Marshal(c.FormatConfigs)
	if err != nil {
		return "", err
	}
	return string(bt), nil
}
