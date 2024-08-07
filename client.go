package rpc

import (
	"dream_program/config"
	"dream_program/discovery"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

var conns []*grpc.ClientConn //客户端列表

type run struct {
	err error //实现错误传播
}

// 运行代码，不出现错误时才运行
func (r *run) Run(fun func() (*grpc.ClientConn, error)) {
	if r.err != nil {
		return
	} //顺序执行 ,实现错误传播
	conn, err := fun()
	r.err = err
	conns = append(conns, conn)
}

func (r *run) Error() error {
	return r.err
}

// 初始化所有Client
func InitGRPCClients() error {
	r := new(run)
	r.Run(InitSubmitGRPC)

	return r.Error()
}

// 关闭所有Client
func CloseGPRCClients() {
	for _, conn := range conns {
		_ = conn.Close()
	}
}

// 创建GRPC客户端连接 nacos版
func NewGRPCClient(name string) (*grpc.ClientConn, error) {
	// 获取服务实例
	client := discovery.NewNamingNacos()

	instance := client.GetHealthyInstance(name)
	addr := fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	fmt.Println("addr:", addr)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// 创建GRPC客户端连接 consul版
func NewGRPCClient(name string) (*grpc.ClientConn, error) {
	// dail
	conn, err := grpc.Dial(fmt.Sprintf("consul://%s/%s?healthy=true", config.Get().Consul.Addr, name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`))
	if err != nil {
		return conn, err
	}
	fmt.Println(conn.Target())
	return conn, nil
}

// 创建GRPC客户端连接 etcd版
func NewGRPCClient(addr, name string) (*grpc.ClientConn, error) {
	// etcd
	etcdCli, err := clientv3.NewFromURL(addr) //创建 etcd 客户端
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(etcdCli) //创建 etcd 解析器
	if err != nil {
		return nil, err
	}

	// dial
	conn, err := grpc.Dial("etcd:///"+name,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)), //指定默认的服务配置，使用轮询负载均衡策略。
		//grpc.WithChainUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor()),                    // 链式拦截器
	)
	if err != nil {
		return conn, err
	}

	return conn, nil
}
