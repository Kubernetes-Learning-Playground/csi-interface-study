package bootstrap

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golanglearning/new_project/csi_practice/pkg/service"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"net"
	"os"
)

type MyDriver struct {
	NodeId string
}

func NewMyDriver(nodeId string) *MyDriver {
	return &MyDriver{NodeId: nodeId}
}

/*
	IdentityService、NodeService、ControllerService --> 称为"自定义组件": 实现CSI插件必须用户自行实现
	node-driver-registrar、csi-attacher、csi-provisioner  --> 都是"外部组件" 是以sidecar形式与自定义组件配合部署
*/

func (d *MyDriver) Start() {
	ctlSvc := service.NewControllerService()
	identitySvc := service.NewIdentityService()
	nodeSvc := service.NewNodeService(d.NodeId)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(DumpLog), // 类似插件
	}

	// grpc服务
	grpcServer := grpc.NewServer(opts...)
	csi.RegisterControllerServer(grpcServer, ctlSvc)
	csi.RegisterIdentityServer(grpcServer, identitySvc)
	csi.RegisterNodeServer(grpcServer, nodeSvc)

	proto := "unix"
	addr := "/csi/csi.sock"

	if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
		klog.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
	}

	// 协议定为 unix:///csi/csi.sock
	listener, err := net.Listen(proto, addr)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}
	// 启动grpc server
	klog.Info("grpc server start...")
	grpcServer.Serve(listener)
}
