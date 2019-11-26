package configformat

import (
	"fmt"
	"reflect"
	"strconv"
)

type IniFormat struct {
	YmlData
	currentProjectConfig string /*当itemformat是without_tree时，当前项目配置要放最前面，该字段用来临时保存用的*/
}

func (c *IniFormat) BuildString(data interface{}) (bool, string) {
	ftype := reflect.TypeOf(data)
	retVal := ""
	switch ftype.Kind() {
	case reflect.Map:
		fval := reflect.ValueOf(data)
		v, ok := fval.Interface().(map[string]string)
		if !ok {
			for _, v := range fval.MapKeys() {
				isString, val := c.BuildString(fval.MapIndex(v).Interface())
				if isString {
					c.currentProjectConfig += v.String() + " = " + val
					continue
				}
				retVal += "[" + v.String() + "]\n" + val
			}
			return false, retVal
		}
		for key, val := range v {
			_, v := c.BuildString(val)
			retVal += key + " = " + v
		}
	case reflect.String:
		fval := reflect.ValueOf(data)
		val, ok := fval.Interface().(string)
		if !ok {
			return true, "\n"
		}
		if _, err := strconv.ParseBool(val); err == nil {
			return true, val + "\n"
		} else if _, err := strconv.ParseInt(val, 10, 64); err == nil {
			return true, val + "\n"
		} else if _, err := strconv.ParseFloat(val, 64); err == nil {
			return true, val + "\n"
		} else {
			return true, fmt.Sprintf("%q\n", val)
		}
	default:
		fmt.Println("type not implement")
	}
	return false, retVal
}
func (c *IniFormat) Format(data *YmlData, projectName string) (string, error) {
	c.YmlData = *data
	c.ProjectName = projectName
	if err := c.FormatItem(); err != nil {
		return "", err
	}
	_, v := c.BuildString(c.FormatConfigs)
	if c.currentProjectConfig != "" {
		v = c.currentProjectConfig + v
	}
	return v, nil
}
