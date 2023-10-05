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
	"strings"
)

var VolumeSet mock.FakeVolumes

func init() {
	VolumeSet = make(mock.FakeVolumes, 0)
}

// ControllerServer 用于创建、删除以及管理 Volume 存储卷
// Controller Service (NFS)  "mount –t xxxxxx -- NodeService"
// 用于实现创建/删除 volume 等 不需要在特定宿主机完成的操作、譬如和云商的API进行交互 以及attach操作等
type ControllerServer struct {
	myDriver     *MyDriver
	mounter      mount.Interface
	storeRecords map[string]map[string]string
}

// ControllerService 对象需要实现csi.ControllerServer接口
var _ csi.ControllerServer = &ControllerServer{}

func NewControllerService(driver *MyDriver) *ControllerServer {
	return &ControllerServer{
		myDriver:     driver,
		mounter:      mount.New(""),
		storeRecords: map[string]map[string]string{},
	}
}

// CreateVolume 创建存储卷(Provision操作)
// 由 sidecar 容器(external provisioner)调用
// 当用户创建 pvc 后，csi 服务会自动调用此方法，
// 创建出一个符合容量大小的 pv，让 pv controller 自己去 bound
func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.Info("create volume...")
	klog.Info("volume name: ", req.GetName()) // pv 名称
	klog.Info("param: ", req.GetParameters())

	var serverIp, shareDir string
	// 从 storageClass parameters 中获取
	if req.GetParameters() != nil {
		serverIp = req.GetParameters()["server"]
		shareDir = req.GetParameters()["share"]
		if serverIp == "" {
			fmt.Println("empty server ip...")
			serverIp = "10.0.0.8"
		}
		if shareDir == "" {
			fmt.Println("empty share dir...")
			shareDir = "/home/test"
		}
		// 存储
		cs.storeRecords[req.GetName()] = map[string]string{
			"server": serverIp,
			"share":  shareDir,
		}
	}

	basePath := fmt.Sprintf("%s:%s", serverIp, shareDir)
	// 创建子目录，给多个 pod 使用
	// 避免多个 pod 挂载时，全都挂载到根目录中
	err := MountTemp(basePath, req.GetName(), cs.mounter, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      "my-nfs-csi-volume-" + req.GetName(),
			CapacityBytes: 0,
		},
	}, nil
}

// DeleteVolume 删除存储卷
/*
# 当进行 kubectl delete pvc mycsi-pvc 会调用此方法
[root@vm-0-12-centos ~]# kubectl get pvc
NAME                STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
myappdata-myapp-0   Pending                                                                        gluster-dynamic   315d
mycsi-pvc           Bound     pvc-92efb72c-493c-4591-abc2-efd6e5694c1c   2Gi        RWO            mycsi-sc          34m
nfs-pvc             Bound     pvc-3f52cf3f-6bc7-42b9-9d03-1ec1d33db45b   10Gi       RWX            nfs-csi           138d
[root@vm-0-12-centos ~]# kubectl delete pvc mycsi-pvc
persistentvolumeclaim "mycsi-pvc" deleted
*/
func (cs *ControllerServer) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.Info("delete volume...")

	vid := request.GetVolumeId()
	pvName := strings.Replace(vid, "my-nfs-csi-volume-", "", -1)
	storeMap := cs.storeRecords[vid]
	basePath := fmt.Sprintf("%s:%s", storeMap["server"], storeMap["share"])
	if err := MountTemp(basePath, pvName, cs.mounter, false); err != nil {
		klog.Warningf("delete volume error...")
		return nil, err
	}
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume 挂载存储卷 (Attach操作)
// 将存储卷挂载到目标节点上
func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.Info("publish volume...")
	return &csi.ControllerPublishVolumeResponse{}, nil
}

// ControllerUnpublishVolume 卸载存储卷
func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.Info("unPublish volume...")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities 验证Controller信息
func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateVolumeCapabilities not implemented")
}

// ListVolumes 列出存储卷
func (cs *ControllerServer) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.Info("list volume...")
	return &csi.ListVolumesResponse{
		Entries: VolumeSet.List(),
	}, nil
}

// GetCapacity 存储卷可用量信息
func (cs *ControllerServer) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return &csi.GetCapacityResponse{
		AvailableCapacity: 100 * 1024 * 1024,
	}, nil
}

// ControllerGetCapabilities controller插件的功能点，如是否支持GetCapacity接口，是否支持snapshot功能等
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context, request *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.myDriver.Cscap,
	}, nil
}

// CreateSnapshot 创建存储卷快照，并创建快照对象(VolumeSnapshot)
func (cs *ControllerServer) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSnapshot not implemented")

}

// DeleteSnapshot 删除存储卷快照
func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSnapshot not implemented")
}

// ListSnapshots 列出快照
func (cs *ControllerServer) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSnapshots not implemented")
}

// ControllerExpandVolume 扩容
func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerExpandVolume not implemented")
}

// ControllerGetVolume 获得卷
func (cs *ControllerServer) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
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

// 官方的一个拼凑ID的方式
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
