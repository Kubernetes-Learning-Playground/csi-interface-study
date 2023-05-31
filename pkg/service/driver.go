package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	"net"
	"os"
)

type MyDriverOptions struct {
	NodeID                string
	DriverName            string
	Endpoint              string
	DefaultOnDeletePolicy string
}

type MyDriver struct {
	Name     string
	NodeID   string
	Version  string
	Endpoint string

	Cscap []*csi.ControllerServiceCapability
	Nscap []*csi.NodeServiceCapability
}

func NewMyDriver(opt *MyDriverOptions) *MyDriver {
	m := &MyDriver{
		NodeID:   opt.NodeID,
		Name:     opt.DriverName,
		Version:  "v0.0.1",
		Endpoint: opt.Endpoint,
	}

	m.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME, //删除和创建volume
		//csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME, // 包含attach过程
	})

	m.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
		csi.NodeServiceCapability_RPC_UNKNOWN,
	})

	return m
}

func (d *MyDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		csc = append(csc, NewControllerServiceCapability(c))
	}
	d.Cscap = csc
}

func (d *MyDriver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	d.Nscap = nsc
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

/*
 IdentityService、NodeService、ControllerService --> 称为"自定义组件": 实现CSI插件必须用户自行实现
 node-driver-registrar、csi-attacher、csi-provisioner  --> 都是"外部组件" 是以sidecar形式与自定义组件配合部署
*/

func (d *MyDriver) Start() {
	ctlSvc := NewControllerService(d)
	identitySvc := NewIdentityService(d)
	nodeSvc := NewNodeService(d)

	opts := []grpc.ServerOption{
		//grpc.UnaryInterceptor(DumpLog), // 类似插件
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
