/*Package config 为保证配置加载在第一时间发生，故将加载配置独立到单独包中。
主程序使用时需进行导入包，已方便配置生效。
	import . "github.com/ysqi/goall-robot/config"
*/
package config

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// 运行模式(prod,test,dev)
var RunMode = "dev"

// 程序根目录，默认为运行时程序所在目录，
// 但在测试模式下将调整为源文件项目根目录
var AppDir string

func loadVar() {
	if flag.Lookup("test.v") != nil {
		RunMode = "test"
	}
	if RunMode == "test" {
		//因为此处在init中，故获得是当前config.go文件的存放目录，
		// 默认配置文件在的当前目录的上一级目录中
		_, file, _, _ := runtime.Caller(0)
		AppDir = filepath.Dir(filepath.Dir(file))
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			glog.Fatal(err)
		}
		AppDir = pwd
	}
	glog.Info("App Dir:", AppDir)
}
func loadConfig() {
	var cfgName string
	if f := flag.Lookup("config"); f != nil {
		cfgName = f.Value.String()
	} else {
		cfgName = "config"
	}
	viper.SetEnvPrefix("GOALL_")
	viper.SetConfigName(cfgName)
	viper.AutomaticEnv()

	//优先是用户根目录 $HOME/.goall/config.yaml
	u, err := user.Current()
	if err != nil {
		glog.Fatal(err)
	}
	viper.AddConfigPath(filepath.Join(u.HomeDir, ".goall"))
	//其次是程序目录
	viper.AddConfigPath(AppDir)

	err = viper.ReadInConfig()
	if err != nil {
		glog.Fatalf("Fatal error config file: %s \n", err)
	}
}

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}

	glog.Infoln("load config...")
	loadVar()
	loadConfig()
	glog.Infoln("load config end")
}
