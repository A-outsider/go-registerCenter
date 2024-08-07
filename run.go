package rpc

import (
	"dream_program/discovery"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

// Run 运行微服务，传入一个注册函数
func Run(host string, port int, name string, register func(grpcServ *grpc.Server)) error {

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("address: ", addr)

	// 注册consul节点
	err := discovery.RegisterConsulEndpoint(name, host, "", port)
	if err != nil {
		return err
	}

	// 注册nacos节点
	discovery.RegisterEndpoint(name, port)

	// 注册etcd节点
	discovery.RegisterEtcdEndpoint(name, addr, "")

	// 建一个tcp连接
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	//注册一个grpc的服务
	grpcServ := grpc.NewServer()
	//grpc.ChainUnaryInterceptor(grpc_opentracing.UnaryServerInterceptor()),

	//启动grpc的健康检查
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(grpcServ, healthcheck)

	//将grpc服务与对应的节点绑定
	register(grpcServ)

	//用tcp升级为grpc连接 ,并启动服务
	if err := grpcServ.Serve(listen); err != nil {
		return err
	}
	return nil
}
