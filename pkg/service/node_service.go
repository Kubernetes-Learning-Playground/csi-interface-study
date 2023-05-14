package service

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// NodeService：用于将 Volume 存储卷挂载到指定的目录中以便 Kubelet 创建容器时使用
//（需要监听在 /var/lib/kubelet/plugins/[SanitizedCSIDriverName]/csi.sock）
// 真正的执行 mount、unmount。所以它必须在每台机器上都存在(daemonset)
type NodeService struct {
	nodeID string
}

var _ csi.NodeServer = &NodeService{}

func NewNodeService(nodeID string) *NodeService {
	return &NodeService{nodeID: nodeID}
}

// NodeUnstageVolume NodeStageVolume的逆操作，将一个存储卷从临时目录umount掉
func (n *NodeService) NodeUnstageVolume(ctx context.Context, request *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	//TODO implement me
	return nil, status.Error(codes.Unimplemented, "")
}

// NodePublishVolume 将存储卷从临时目录mount到目标目录（pod目录）
func (n *NodeService) NodePublishVolume(ctx context.Context, request *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	//TODO implement me
	klog.Infof("NodePublishVolume")
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume 将存储卷从pod目录umount掉
func (n *NodeService) NodeUnpublishVolume(ctx context.Context, request *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	//TODO implement me
	klog.Infof("NodeUnpublishVolume")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetVolumeStats 返回可用于该卷的卷容量统计信息。
func (n *NodeService) NodeGetVolumeStats(ctx context.Context, request *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	//TODO implement me
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeExpandVolume node上执行卷扩容
func (n *NodeService) NodeExpandVolume(ctx context.Context, request *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	//TODO implement me
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetCapabilities 返回Node插件的功能点，如是否支持stage/unstage功能
func (n *NodeService) NodeGetCapabilities(ctx context.Context, request *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	//TODO implement me
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

// NodeGetInfo 获取节点信息
func (n *NodeService) NodeGetInfo(ctx context.Context, request *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	//TODO implement me
	klog.Infoln("NodeGetInfo")
	return &csi.NodeGetInfoResponse{
		NodeId: n.nodeID,
	}, nil
}

// NodeStageVolume 如果存储卷没有格式化，首先要格式化。
// 然后把存储卷mount到一个临时的目录（这个目录通常是节点上的一个全局目录）。
// 再通过NodePublishVolume将存储卷mount到pod的目录中。
// mount过程分为2步，原因是为了支持多个pod共享同一个volume（如NFS）。
// 如果使用云盘，
// 就会将云硬盘格式化成对应文件系统 将volume mount到一个全局的目录
func (n *NodeService) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	//TODO implement me
	return nil, status.Error(codes.Unimplemented, "")
}
