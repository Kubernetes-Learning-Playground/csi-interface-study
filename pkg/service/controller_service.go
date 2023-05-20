package service

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golanglearning/new_project/csi_practice/pkg/service/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

var VolumeSet mock.FakeVolumes

func init() {
	VolumeSet = make(mock.FakeVolumes, 0)
}

// ControllerService：用于创建、删除以及管理 Volume 存储卷
// Controller Service (NFS)  "mount –t xxxxxx -- NodeService"
// 用于实现创建/删除 volume 等 不需要在特定宿主机完成的操作、譬如和云商的API进行交互 以及attach操作等
type ControllerService struct{}

// ControllerService 对象需要实现csi.ControllerServer接口
var _ csi.ControllerServer = &ControllerService{}

func NewControllerService() *ControllerService {
	return &ControllerService{}
}

// CreateVolume 创建存储卷
func (*ControllerService) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	return &csi.CreateVolumeResponse{
		Volume: VolumeSet.Create(),
	}, nil
}

// DeleteVolume 删除存储卷
func (c *ControllerService) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.Info("delete volume...")
	VolumeSet.Delete(request.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume 发布存储卷
func (c *ControllerService) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.Info("publish volume...")
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (c *ControllerService) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.Info("unPublish volume...")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities 验证存储卷
func (c *ControllerService) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateVolumeCapabilities not implemented")
}

// ListVolumes 列出存储卷
func (c *ControllerService) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.Info("list volume...")
	return &csi.ListVolumesResponse{
		Entries: VolumeSet.List(),
	}, nil
}

// GetCapacity 存储卷可用量信息
func (c *ControllerService) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return &csi.GetCapacityResponse{
		AvailableCapacity: 100 * 1024 * 1024,
	}, nil
}

// ControllerGetCapabilities controller插件的功能点，如是否支持GetCapacity接口，是否支持snapshot功能等
func (c *ControllerService) ControllerGetCapabilities(ctx context.Context, request *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	capList := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME, //删除和创建volume
		//csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME, // 包含attach过程
	}
	var caps []*csi.ControllerServiceCapability
	for _, capObj := range capList {
		c := &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: capObj,
				},
			},
		}
		caps = append(caps, c)
	}
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}

// CreateSnapshot 创建快照
func (c *ControllerService) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSnapshot not implemented")

}

// DeleteSnapshot 删除快照
func (c *ControllerService) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSnapshot not implemented")
}

// ListSnapshots 列出快照
func (c ControllerService) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSnapshots not implemented")
}

// ControllerExpandVolume 扩容
func (c *ControllerService) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerExpandVolume not implemented")
}

// ControllerGetVolume 获得卷
func (c *ControllerService) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	v, err := VolumeSet.Get(request.VolumeId)
	if err != nil {
		return nil, err
	}
	return &csi.ControllerGetVolumeResponse{
		Volume: v,
	}, nil
}
