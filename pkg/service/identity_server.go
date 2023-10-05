package service

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
 Identity Server 身分认证服务
 NodePlugin 与 ControllerPlugin 都必须实现
 driver-registrar组件会调用此接口把CSI driver 注册到kubelet中
*/

// IdentityServer 用于 Kubernetes 与 CSI 插件协调版本信息
// 暴露插件的名称和能力
type IdentityServer struct {
	myDriver *MyDriver
}

var _ csi.IdentityServer = &IdentityServer{}

func NewIdentityService(driver *MyDriver) *IdentityServer {
	return &IdentityServer{myDriver: driver}
}

// GetPluginCapabilities 返回driver提供的能力，比如是否提供 ControllerServer,volume 访问能力
func (i *IdentityServer) GetPluginCapabilities(ctx context.Context, request *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	capList := []csi.PluginCapability_Service_Type{
		csi.PluginCapability_Service_CONTROLLER_SERVICE,
		csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
	}
	var caps []*csi.PluginCapability
	for _, capObj := range capList {
		c := &csi.PluginCapability{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{
					Type: capObj,
				},
			},
		}
		caps = append(caps, c)
	}
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}

// Probe 探针
func (i *IdentityServer) Probe(ctx context.Context, request *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	s := wrappers.BoolValue{Value: true}
	return &csi.ProbeResponse{
		Ready: &s,
	}, nil
}

// GetPluginInfo 返回driver的信息
func (i *IdentityServer) GetPluginInfo(ctx context.Context, request *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	if i.myDriver.Name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if i.myDriver.Version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}
	return &csi.GetPluginInfoResponse{
		Name:          i.myDriver.Name,
		VendorVersion: i.myDriver.Version,
	}, nil
}
