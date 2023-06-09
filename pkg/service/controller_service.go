package service

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github/mycsi/csi_practice/pkg/service/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"os"
	"strings"
)

var VolumeSet mock.FakeVolumes

func init() {
	VolumeSet = make(mock.FakeVolumes, 0)
}

// ControllerService：用于创建、删除以及管理 Volume 存储卷
// Controller Service (NFS)  "mount –t xxxxxx -- NodeService"
// 用于实现创建/删除 volume 等 不需要在特定宿主机完成的操作、譬如和云商的API进行交互 以及attach操作等
type ControllerService struct {
	myDriver *MyDriver
	mounter  mount.Interface //依然要初始化这个
}

// ControllerService 对象需要实现csi.ControllerServer接口
var _ csi.ControllerServer = &ControllerService{}

func NewControllerService(driver *MyDriver) *ControllerService {
	return &ControllerService{myDriver: driver, mounter: mount.New("")}
}

// CreateVolume 创建存储卷 (Provision操作)
func (cs *ControllerService) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.Info("create volume...")
	klog.Info("volume name: ", req.GetName())
	klog.Info("param: ", req.GetParameters())

	// TODO: 改成可配置 storageClass中配置
	basePath := "10.0.0.8:/home/test" //  根目录
	tmpPath := "/tmp/"
	volCap := &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
	}
	opts := volCap.GetMount().GetMountFlags()

	// 检查目录
	nn, err := cs.mounter.IsLikelyNotMountPoint(tmpPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Info("如果没有，需要创建:", tmpPath)
			err = os.MkdirAll(tmpPath, 0777)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			nn = true
		}
	}
	if !nn {
		return nil, status.Error(codes.Internal, "无法处理tmp目录进行临时挂载")
	}

	// 挂载到临时目录
	err = cs.mounter.Mount(basePath, tmpPath, "nfs", opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer func() {
		err := mount.CleanupMountPoint(tmpPath, cs.mounter, true)
		if err != nil {
			klog.Warningf("cs 反挂出错", err)
		}
	}()
	//一旦挂载成功， 那我们就可以再/tmp/pvc-xxx-xx-x-
	if err = os.Mkdir(tmpPath+req.GetName(), 0777); err != nil && !os.IsExist(err) {
		return nil, status.Errorf(codes.Internal, "failed to make subdirectory: %v", err.Error())
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      "jtthink-volume-" + req.GetName(),
			CapacityBytes: 0,
		},
	}, nil
}

// DeleteVolume 删除存储卷
func (cs *ControllerService) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.Info("delete volume...")
	VolumeSet.Delete(request.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume 挂载存储卷 (Attach操作)
// 将存储卷挂载到目标节点上
func (cs *ControllerService) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.Info("publish volume...")
	return &csi.ControllerPublishVolumeResponse{}, nil
}

// ControllerUnpublishVolume 卸载存储卷
func (cs *ControllerService) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.Info("unPublish volume...")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities 验证Controller信息
func (cs *ControllerService) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateVolumeCapabilities not implemented")
}

// ListVolumes 列出存储卷
func (cs *ControllerService) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.Info("list volume...")
	return &csi.ListVolumesResponse{
		Entries: VolumeSet.List(),
	}, nil
}

// GetCapacity 存储卷可用量信息
func (cs *ControllerService) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return &csi.GetCapacityResponse{
		AvailableCapacity: 100 * 1024 * 1024,
	}, nil
}

// ControllerGetCapabilities controller插件的功能点，如是否支持GetCapacity接口，是否支持snapshot功能等
func (cs *ControllerService) ControllerGetCapabilities(ctx context.Context, request *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.myDriver.Cscap,
	}, nil
}

// CreateSnapshot 创建存储卷快照，并创建快照对象(VolumeSnapshot)
func (cs *ControllerService) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSnapshot not implemented")

}

// DeleteSnapshot 删除存储卷快照
func (cs *ControllerService) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSnapshot not implemented")
}

// ListSnapshots 列出快照
func (cs ControllerService) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSnapshots not implemented")
}

// ControllerExpandVolume 扩容
func (cs *ControllerService) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerExpandVolume not implemented")
}

// ControllerGetVolume 获得卷
func (cs *ControllerService) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	v, err := VolumeSet.Get(request.VolumeId)
	if err != nil {
		return nil, err
	}
	return &csi.ControllerGetVolumeResponse{
		Volume: v,
	}, nil
}

// ////////////////////////////////以下是自定义函数

const (
	paramServer           = "server"
	paramShare            = "share"
	paramSubDir           = "subdir"
	mountOptionsField     = "mountoptions"
	mountPermissionsField = "mountpermissions"
	pvcNameKey            = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey       = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey             = "csi.storage.k8s.io/pv/name"
	pvcNameMetadata       = "${pvc.metadata.name}"
	pvcNamespaceMetadata  = "${pvc.metadata.namespace}"
	pvNameMetadata        = "${pv.metadata.name}"
	separator             = "#"
)

type nfsVolume struct {
	// Volume id
	id string
	// Address of the NFS server.
	// Matches paramServer.
	server string
	// Base directory of the NFS server to create volumes under
	// Matches paramShare.
	baseDir string
	// Subdirectory of the NFS server to create volumes under
	subDir string
	// size of volume
	size int64
	// pv name when subDir is not empty
	uuid string
}

const (
	idServer = iota
	idBaseDir
	idSubDir
	idUUID
	totalIDElements // Always last
)

func replaceWithMap(str string, m map[string]string) string {
	for k, v := range m {
		if k != "" {
			str = strings.ReplaceAll(str, k, v)
		}
	}
	return str
}

// 官方的一个 拼凑ID的方式
func getVolumeIDFromNfsVol(vol *nfsVolume) string {
	idElements := make([]string, totalIDElements)
	idElements[idServer] = strings.Trim(vol.server, "/")
	idElements[idBaseDir] = strings.Trim(vol.baseDir, "/")
	idElements[idSubDir] = strings.Trim(vol.subDir, "/")
	idElements[idUUID] = vol.uuid
	return strings.Join(idElements, separator)
}
func newNFSVolume(name string, size int64, params map[string]string) (*nfsVolume, error) {
	var server, baseDir, subDir string
	subDirReplaceMap := map[string]string{}

	for k, v := range params {
		switch strings.ToLower(k) {
		case paramServer:
			server = v
		case paramShare:
			baseDir = v
		case paramSubDir:
			subDir = v
		case pvcNamespaceKey:
			subDirReplaceMap[pvcNamespaceMetadata] = v
		case pvcNameKey:
			subDirReplaceMap[pvcNameMetadata] = v
		case pvNameKey:
			subDirReplaceMap[pvNameMetadata] = v
		}
	}

	if server == "" {
		return nil, fmt.Errorf("%v is a required parameter", paramServer)
	}

	vol := &nfsVolume{
		server:  server,
		baseDir: baseDir,
		size:    size,
	}
	if subDir == "" {
		// use pv name by default if not specified
		vol.subDir = name
	} else {
		// replace pv/pvc name namespace metadata in subDir
		vol.subDir = replaceWithMap(subDir, subDirReplaceMap)
		// make volume id unique if subDir is provided
		vol.uuid = name
	}
	vol.id = getVolumeIDFromNfsVol(vol)
	return vol, nil
}
