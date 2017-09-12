package rebot

import (
	"net/http"

	"github.com/golang/glog"

	"github.com/spf13/viper"
)

func startHTTP() {
	http.HandleFunc("/githubhook", githubMonitor.GithubHookHandler)
	go func() {
		baseURL := viper.GetString("HTTPBaseURL")
		if baseURL == "" {
			baseURL = ":2017"
		}
		glog.Infof("Running HTTP on %s\n", baseURL)
		err := http.ListenAndServe(baseURL, nil)
		if err != nil {
			glog.Fatalln(err)
		}
	}()
}

func defaultHandler(w http.ResponseWriter, h *http.Request) {
	w.Write([]byte("hello"))
}

// outputError 输出错误信息到HTTP Response
func outputError(w http.ResponseWriter, err string) {
	w.Write([]byte(err))
	w.WriteHeader(500)
}
