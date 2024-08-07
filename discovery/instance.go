package discovery

import (
	"fmt"
	"net"
	"strings"
)

type Server struct {
	Name    string
	group   string
	Addr    string
	Port    uint64
	Version string // 一般不考虑用版本控制
	Tags    []string
}

// 服务目标路径
func (srv *Server) target() string {
	if srv.Version == "" {
		return fmt.Sprintf("%s", srv.Name)
	}
	return fmt.Sprintf("%s/%s", srv.Name, srv.Version)
}

// 服务唯一key
func (srv *Server) key() string {
	return fmt.Sprintf("%s/%s", srv.target(), srv.Addr)
}

// 获取本机ip地址
func getHostIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println("get current host ip err: ", err)
		return ""
	}
	addr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(addr.String(), ":")[0]
	return ip
}
