package rebot

import (
	"errors"
	"strings"

	"github.com/go-ini/ini"
)

// PageInfo 获取的信息集合
// type PageInfo map[string]interface{}

func analysisNewIssue(body string) error {
	f, err := ini.Load(body)
	if err != nil {
		return err
	} else if section, err := f.GetSection(""); err != nil {
		return err
	} else {
		rawurl := section.Key("url").String()
		if rawurl == "" {
			return errors.New("缺失url")
		}
		info := make(map[string]interface{}, 10)
		info["url"] = rawurl

		// 标签过滤处理
		tags := section.Key("tags").String()
		if tags != "" {
			tags = strings.Replace(tags, "，", ",", -1)
			arr := strings.Split(tags, ",")
			for i, t := range arr {
				arr[i] = strings.TrimSpace(t)
			}
			info["tags"] = arr
		}
		// u, err := url.Parse(rawurl)
		// if err != nil {
		// 	return err
		// }

	}
	return nil
}
