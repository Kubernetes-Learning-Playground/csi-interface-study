package mock

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// 已经废弃

type FakeVolumes []*csi.Volume

func (fv FakeVolumes) Delete(id string) {
	for i, v := range fv {
		if v.VolumeId == id {
			fv = append(fv[:i], fv[i+1:]...)
			return
		}
	}
}

func (fv FakeVolumes) List() []*csi.ListVolumesResponse_Entry {
	ret := make([]*csi.ListVolumesResponse_Entry, 0)
	for _, v := range fv {
		ret = append(ret, &csi.ListVolumesResponse_Entry{
			Volume: v,
		})
	}
	return ret
}

func (fv FakeVolumes) Create() *csi.Volume {
	v := &csi.Volume{
		VolumeId:      "my-nfs-csi-volume-" + time.Now().Format("20060102150405"),
		CapacityBytes: 10 * 1024 * 1024 * 1024, //统一 使用 10G
	}
	fv = append(fv, v)
	return v
}

func (fv FakeVolumes) Get(id string) (*csi.Volume, error) {
	for _, v := range fv {
		if v.VolumeId == id {
			return v, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "found no volume")
}
