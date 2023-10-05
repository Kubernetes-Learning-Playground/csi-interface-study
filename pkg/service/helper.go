package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"os"
)

// MountTemp 挂载临时目录，并创建子文件夹
// createSubDir 是否创建子文件夹  如果是 false 则会尝试删除子文件夹
func MountTemp(basePath, pvName string, mounter mount.Interface, createSubDir bool) error {

	tmpPath := "/tmp/"
	volCap := &csi.VolumeCapability{
		AccessType: &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		},
	}
	opts := volCap.GetMount().GetMountFlags()

	// 下面检查目录
	nn, err := mounter.IsLikelyNotMountPoint(tmpPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(tmpPath, 0777)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			nn = true
		}
	}
	if !nn {
		return status.Error(codes.Internal, "Unable to handle tmp directory for temporary mounting")
	}

	// 挂载到临时目录
	err = mounter.Mount(basePath, tmpPath, "nfs", opts)
	if err != nil {
		status.Error(codes.Internal, err.Error())
	}

	defer func() {
		err := mount.CleanupMountPoint(tmpPath, mounter, true)
		if err != nil {

			klog.Warningf("cs hook err: ", err)
		}
	}()

	if createSubDir {
		//一旦挂载成功， 那我们就可以再/tmp/pvc-xxx-xx-x-
		if err = os.Mkdir(tmpPath+pvName, 0777); err != nil && !os.IsExist(err) {
			return status.Errorf(codes.Internal, "Unable to create sub folder: %v", err.Error())
		}
	} else {
		if err = os.RemoveAll(tmpPath + pvName); err != nil {
			return status.Errorf(codes.Internal, "Unable to delete sub folder: %v", err.Error())
		}
	}

	return nil

}
