package discovery

import (
	"dream_program/config"
	"fmt"
	"github.com/hashicorp/consul/api"
	"os"
	"os/signal"
	"syscall"

	"log"
)

type Consul struct {
	server *Server
	cli    *api.Client
}

// 获取consul实例
func NewConsul() (consul *Consul, err error) {
	consul = new(Consul)
	server := new(Server)
	con := api.DefaultConfig()
	con.Address = config.Get().Consul.Addr

	client, err := api.NewClient(con)
	if err != nil {
		return nil, err
	}
	consul.cli = client
	consul.server = server
	return consul, nil
}

// 服务注册具体流程
func RegisterConsulEndpoint(serviceName, host, tag string, port int) (err error) {

	consul, err := NewConsul()
	if err != nil {
		log.Fatalf("连接到Consul失败: %v", err)
	}

	consul.server.Name = serviceName
	consul.server.Port = uint64(port)
	consul.server.Addr = getHostIp() // 注意这里的addr指定的host地址
	consul.server.Tags = []string{"dream", tag}

	err = consul.register()
	if err != nil {
		log.Fatalf("注册服务到Consul失败: %v", err)
	}

	// 手动关闭服务
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
		<-ch
		consul.Deregister(fmt.Sprintf("%s-%s-%d", consul.server.Name, consul.server.Addr, consul.server.Port))
	}()

	return
}

func (consul *Consul) register() error {
	check := &api.AgentServiceCheck{
		TCP:                            fmt.Sprintf("%s:%d", consul.server.Addr, consul.server.Port), // 这里一定是外部可以访问的地址
		Timeout:                        "30s",                                                        //表示Consul健康检查在等待服务响应的时间可以长达30秒。
		Interval:                       "60s",                                                        //健康检查间隔
		DeregisterCriticalServiceAfter: "1m",                                                         // 最小超时时间为1分钟，
	}

	srv := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", consul.server.Name, consul.server.Addr, consul.server.Port), //TODO 后面可在增加前缀提升扩展性
		Name:    consul.server.Name,                                                                  // 服务名称
		Tags:    consul.server.Tags,                                                                  // 打标签
		Address: consul.server.Addr,
		Port:    int(consul.server.Port),
		Check:   check,
	}

	return consul.cli.Agent().ServiceRegister(srv)
}

// ---------------consul注销服务------------------------
func (c *Consul) Deregister(serviceID string) error {
	return c.cli.Agent().ServiceDeregister(serviceID)
}
