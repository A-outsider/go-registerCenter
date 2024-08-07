package discovery

import (
	"bytes"
	"dream_program/config"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

//func init(){
//	config.initConfig()
//}

func InitConfig(name, group string, cfg any) {

	//create ServerConfig
	serverConfig := []constant.ServerConfig{
		*constant.NewServerConfig(config.Get().Nacos.Host, uint64(config.Get().Nacos.Port)),
	}

	//create ClientConfig
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("log"),
		constant.WithCacheDir("cache"),
		constant.WithLogLevel("debug"),
	)

	// create config client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	//get config
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: name,
		Group:  group,
	})
	if err != nil {
		panic(err.Error())
	}

	// 赋值
	getConfigData(cfg, content)

	// 监听操作
	err = client.ListenConfig(vo.ConfigParam{
		DataId: name,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("配置文件发生了变化...")
			fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
			getConfigData(cfg, data)
		},
	})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getConfigData(cfg any, content string) {
	var vipers = viper.New()
	vipers.SetConfigType("yaml")
	err := vipers.ReadConfig(bytes.NewBuffer([]byte(content)))
	if err != nil {
		fmt.Println("error reading config: %w", err)
	}
	err = vipers.Unmarshal(cfg)
	if err != nil {
		fmt.Println("error unmarshaling config: %w", err)
	}
}
