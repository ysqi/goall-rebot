package git

import "testing"
import . "github.com/smartystreets/goconvey/convey"

func init() {

}

func TestGitPull(t *testing.T) {
	Convey("git pull", t, func() {
		Convey("", func() {
			msg, err := CommitFile("o1.md", "test")
			t.Log("Output Info:", msg)
			So(err, ShouldBeNil)
		})
	})
}
