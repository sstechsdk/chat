package main

import (
	"flag"
	"github.com/OpenIMSDK/chat/pkg/common/chatrpcstart"
	"github.com/OpenIMSDK/chat/tools/component"
	"github.com/OpenIMSDK/tools/log"
	"math/rand"
	"time"

	"github.com/OpenIMSDK/chat/internal/rpc/chat"
	"github.com/OpenIMSDK/chat/pkg/common/config"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var configFile string
	flag.StringVar(&configFile, "config_folder_path", "../config/config.yaml", "Config full path")

	var rpcPort int

	flag.IntVar(&rpcPort, "port", 30400, "get rpc ServerPort from cmd")

	var hide bool
	flag.BoolVar(&hide, "hide", true, "hide the ComponentCheck result")

	flag.Parse()
	err := component.ComponentCheck(configFile, hide)
	if err != nil {
		return
	}
	if err := config.InitConfig(configFile); err != nil {
		panic(err)
	}
	if err := log.InitFromConfig("oauth.log", "oauth-rpc", *config.Config.Log.RemainLogLevel, *config.Config.Log.IsStdout, *config.Config.Log.IsJson, *config.Config.Log.StorageLocation, *config.Config.Log.RemainRotationCount, *config.Config.Log.RotationTime); err != nil {
		panic(err)
	}
	err = chatrpcstart.Start(rpcPort, config.Config.RpcRegisterName.OpenImOauth2Name, 0, chat.Start)
	if err != nil {
		panic(err)
	}
}
