package rebot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/ysqi/goall-robot/git"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"github.com/ysqi/com"
)

var hugo Hugo

// Hugo 处理Hugo结构相关
type Hugo struct {
	dir      string
	sections []string
}

func init() {
	dir := viper.GetString("GitPush.GIT_BASEDIR")
	if dir == "" {
		glog.Fatalf("缺失配置项GitPush.GIT_BASEDIR")
	} else if !com.IsExist(dir) {
		glog.Fatalf("GitPush.GIT_BASEDIR=%q路径不存在", dir)
	}
	hugo = Hugo{
		dir:      dir,
		sections: []string{"article", "book", "job", "news", "pkg"},
	}
}

const articleMDTemplate = `
+++
title= "{{.P.Title}}"
date= {{.Issue.GetCreatedAt}}

link= "{{.P.URL}}"
tags=[{{.P.Tags}}]

# 提交信息
[submit]
    url = "{{.Issue.GetHTMLURL}}"
    [submit.by]
        name = "{{.Issue.User.GetLogin}}"
        url = "{{.Issue.User.GetBlog}}"
+++

{{.P.Content}}
`

// GenPageContent 将Issue转换为指定的文章内容
func (h Hugo) GenPageContent(category string, page PageInfo, issue *github.Issue) ([]byte, error) {
	data := map[string]interface{}{
		"P":     page,
		"Issue": issue,
	}

	t, err := template.New("").Parse(articleMDTemplate)
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer([]byte{})
	err = t.Execute(r, data)
	if err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

// CreatePage 在本地创建Page文件，并返回文件路径
// 按照所指定的分类，在本地对应目录下创建md文件
// 		1.article : root/content/articles/YYYYMM
//		2.news: 	root/content/news/YYYY
//		3.book:	root/content/books
//		4.job:		root/content/jobs/YYYY
// 		5.pkg:		root/content/packages/
func (h Hugo) CreatePage(section, filename string, content []byte) (file string, err error) {
	if !com.IsSliceContainsStr(h.sections, section) {
		return "", fmt.Errorf("非法类型%q无法处理，合理范围为：%s", section, h.sections)
	}
	dir := filepath.Join(h.dir, "content")
	switch section {
	case "article":
		dir = filepath.Join(dir, "articles", time.Now().Format("200601"))
	case "news":
		dir = filepath.Join(dir, "news", time.Now().Format("2006"))
	case "job":
		dir = filepath.Join(dir, "jobs", time.Now().Format("2006"))
	case "book":
		dir = filepath.Join(dir, "books")
	case "pkg":
		dir = filepath.Join(dir, "packages")
	default:
		return "", fmt.Errorf("该类型%q尚未处理", section)
	}
	//如果文件夹不存在，则创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	//先在临时目录中
	tmpfile, err := ioutil.TempFile("", "goall")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name()) //最终删除
	if _, err = tmpfile.Write(content); err != nil {
		return "", err
	}
	if err = tmpfile.Close(); err != nil {
		return "", err
	}
	//复制文件到项目指定目录
	filename = filepath.Join(dir, filename)
	if err = os.Rename(tmpfile.Name(), filename); err != nil {
		return "", err
	}
	return filepath.Rel(h.dir, filename)
}

// Commit 提交Hugo项目文件到Git
func (h Hugo) Commit(file, message string) error {
	result, err := git.CommitFile(file, "Rebot:"+message)
	if err != nil {
		glog.Warning("git push:", result)
		return err
	}
	return nil
}
