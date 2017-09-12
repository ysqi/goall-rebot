package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ysqi/goall-robot/config"

	_ "github.com/ysqi/goall-robot/config"

	"github.com/spf13/viper"
)

func execShell(shellPath string, env map[string]string) (string, error) {
	if _, err := os.Stat(shellPath); os.IsNotExist(err) {
		return "", err
	}
	cmd := exec.Command("/bin/sh", shellPath)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}
	output, err := cmd.Output()
	return string(output), err
}

// CommitFile 提交文件，在提交的同时将会同步Push到远程端
// commit执行push.sh脚本时，会将配置文件中GitPush下所有配置载入
func CommitFile(file string, commintMsg string) (string, error) {
	shellPath := viper.GetString("GitPush.Shell")
	if shellPath == "" {
		return "", fmt.Errorf("配置文件中不存在配置项GitPush.Shell")
	} else if !filepath.IsAbs(shellPath) {
		shellPath = filepath.Join(config.AppDir, shellPath)
	}
	env := viper.GetStringMapString("GitPush")
	env["COMMIT_FILE"] = file
	env["COMMIT_MSG"] = commintMsg
	return execShell(shellPath, env)
}
