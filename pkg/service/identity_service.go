package service

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
)

/*
	Identity Service 身分认证服务
	NodePlugin 与 ControllerPlugin都必须实现
	driver-registrar组件会调用此接口把CSI driver 注册到kubelet中
*/

// IdentityService：用于 Kubernetes 与 CSI 插件协调版本信息
// 暴露插件的名称和能力
type IdentityService struct{}

var _ csi.IdentityServer = &IdentityService{}

func NewIdentityService() *IdentityService {
	return &IdentityService{}
}

// GetPluginCapabilities 返回driver提供的能力，比如是否提供 Controller Service,volume 访问能能力
func (i *IdentityService) GetPluginCapabilities(ctx context.Context, request *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	//TODO implement me
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
func (i *IdentityService) Probe(ctx context.Context, request *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	status := wrappers.BoolValue{Value: true}
	//TODO implement me
	return &csi.ProbeResponse{
		Ready: &status,
	}, nil
}

// GetPluginInfo 返回driver的信息
func (i *IdentityService) GetPluginInfo(ctx context.Context, request *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	//TODO implement me
	return &csi.GetPluginInfoResponse{
		Name:          "mycsi.jtthink.com",
		VendorVersion: "v1.0",
	}, nil
}
