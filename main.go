package main

import (
	"flag"

	"github.com/ysqi/goall-robot/rebot"

	_ "github.com/ysqi/goall-robot/config"
)

func main() {
	flag.Parse()
	rebot.Run()
}
