package rebot

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	_ "github.com/ysqi/goall-robot/config"
)

var (
	githubMonitor *GithubWorker
	endRunning    = make(chan bool, 1)
)

// Run 启动
func Run() {
	var err error
	githubMonitor, err = NewGithubWorker()
	if err != nil {
		glog.Fatalln(err)
	}
	startHTTP()

	githubMonitor.StartWorker()

	go runSignalMonitor()

	glog.Infoln("running...")
	<-endRunning
}

// Exit 指示退出程序
func Exit() {
	glog.Infoln("program exiting")
	glog.Flush()
	githubMonitor.exitChan <- true
	endRunning <- true
}

func runSignalMonitor() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	Exit()
}
