package rebot

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreatePage(t *testing.T) {
	Convey("Hugo Create Page", t, func() {
		Convey("正常", func() {
			_, err := hugo.CreatePage("article", "001.md", []byte("body"))
			So(err, ShouldBeNil)
		})
		Convey("覆盖文件", func() {
			_, err := hugo.CreatePage("article", "001.md", []byte("body"))
			So(err, ShouldBeNil)
		})
	})
}

func TestCommit(t *testing.T) {
	Convey("git commit", t, func() {
		file, err := hugo.CreatePage("article", "001.md", []byte("# body"))
		So(err, ShouldBeNil)
		err = hugo.Commit(file, "test")
		So(err, ShouldBeNil)
	})
}
