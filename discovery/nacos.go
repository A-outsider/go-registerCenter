package discovery

import (
	"dream_program/config"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"os"
	"os/signal"
	"syscall"
)

type Nacos struct {
	server       *Server
	namingClient naming_client.INamingClient
	ConfigClient config_client.IConfigClient // TODO 考虑要不要放一起
}

// 创建nacos实例
func NewNamingNacos() (nacos *Nacos) {
	nacos = new(Nacos)

	//create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(config.Get().Nacos.Host, uint64(config.Get().Nacos.Port)),
	}

	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("log"),
		constant.WithCacheDir("cache"),
		constant.WithLogLevel("debug"),
	)

	// 实例化
	var err error
	nacos.namingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(fmt.Errorf("create Nacos Client failed: %s", err))
	}
	return
}

func RegisterEndpoint(name string, port int) {
	// 创建服务实例
	cli := NewNamingNacos()
	cli.server = &Server{Name: name, Port: uint64(port)}

	// 注册服务中心
	ok, err := cli.RegisterService()
	if !ok || err != nil {
		fmt.Println("register Service Instance failed!")
		panic(err)
	}

	// 手动关闭服务
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
		<-ch
		cli.DestroyEndpoint()
	}()

	//心跳保活: nacos内置实现了
}

func (nacos *Nacos) RegisterService() (ok bool, err error) {
	return nacos.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          getHostIp(),
		Port:        nacos.server.Port,
		ServiceName: nacos.server.Name,
		Weight:      10,   // 表示服务实例的权重
		Enable:      true, // 是否启用服务发现
		Healthy:     true, // 健康检查机制
		Ephemeral:   true, //表示服务实例是否为短暂的
	})
}

func (nacos *Nacos) DestroyEndpoint() {
	param := vo.DeregisterInstanceParam{
		Ip:          getHostIp(),
		Port:        nacos.server.Port,
		ServiceName: nacos.server.Name,
		Ephemeral:   true,
	}
	ok, err := nacos.namingClient.DeregisterInstance(param)
	if !ok || err != nil {
		fmt.Println("Destory Service Instance failed!")
		panic(err)
	}
}

// ------------------------------------------- 服务发现------------------------------------------------------------------------
// 获取一个健康的实例（加权轮训负载均衡）
func (nacos *Nacos) GetHealthyInstance(serviceName string) (instance *model.Instance) {
	instances, err := nacos.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	})
	if err != nil {
		panic("SelectOneHealthyInstance failed!")
	}
	return instances
}

// 获取实例列表
func (nacos *Nacos) SelectInstances(serviceName string) (instances []model.Instance) {
	instances, _ = nacos.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
	})
	return
}
