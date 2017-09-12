package rebot

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewPageInfo(t *testing.T) {
	Convey("New Page Info", t, func() {

		Convey("正常", func() {
			body := `
			title= Go Test
			URL = http://example.com
			tags= Go，Test,Golang
			`
			p, err := NewPageInfo(body)
			So(err, ShouldBeNil)
			So(p, ShouldResemble, PageInfo{Title: "Go Test", Tags: []string{"Go", "Test", "Golang"}, URL: "http://example.com"})
		})
		Convey("必填项Title", func() {
			body := ` 
			url= http://example.com
			`
			_, err := NewPageInfo(body)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "title")
		})
		Convey("必填项URL", func() {
			body := ` 
			title= test
			`
			_, err := NewPageInfo(body)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "url")
		})

		Convey("Page INI", func() {
			body := `
			title=test
			url=http://example.com 
			tags=Go,GoTest，Golang
			`
			p, err := NewPageInfo(body)
			So(err, ShouldBeNil)
			result := p.ToIni()

			So(result, ShouldEqual,
				`title = test
url = http://example.com
tags = Go,GoTest,Golang
`)

			body = `
title=test
url=http://example.com 
tags=
other= test
`
			p, err = NewPageInfo(body)
			So(err, ShouldBeNil)
			result = p.ToIni()

			So(result, ShouldEqual,
				`title = test
url = http://example.com
`)

		})

	})
}
