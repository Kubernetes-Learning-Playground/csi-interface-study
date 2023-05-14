### csi-interface-study CSI接口练习

类似于 CRI，CSI 也是基于 gRPC 实现。详细的 CSI SPEC 可以参考 这里，它要求插件开发者要实现三个 gRPC 服务：
- Identity Service：用于 Kubernetes 与 CSI 插件协调版本信息
- Controller Service：用于创建、删除以及管理 Volume 存储卷
- Node Service：用于将 Volume 存储卷挂载到指定的目录中以便 Kubelet 创建容器时使用（需要监听在 /var/lib/kubelet/plugins/[SanitizedCSIDriverName]/csi.sock）
```go
// ControllerServer is the server API for Controller service.
type ControllerServer interface {
    CreateVolume(context.Context, *CreateVolumeRequest) (*CreateVolumeResponse, error)
    DeleteVolume(context.Context, *DeleteVolumeRequest) (*DeleteVolumeResponse, error)
    ControllerPublishVolume(context.Context, *ControllerPublishVolumeRequest) (*ControllerPublishVolumeResponse, error)
    ControllerUnpublishVolume(context.Context, *ControllerUnpublishVolumeRequest) (*ControllerUnpublishVolumeResponse, error)
    ValidateVolumeCapabilities(context.Context, *ValidateVolumeCapabilitiesRequest) (*ValidateVolumeCapabilitiesResponse, error)
    ListVolumes(context.Context, *ListVolumesRequest) (*ListVolumesResponse, error)
    GetCapacity(context.Context, *GetCapacityRequest) (*GetCapacityResponse, error)
    ControllerGetCapabilities(context.Context, *ControllerGetCapabilitiesRequest) (*ControllerGetCapabilitiesResponse, error)
    CreateSnapshot(context.Context, *CreateSnapshotRequest) (*CreateSnapshotResponse, error)
    DeleteSnapshot(context.Context, *DeleteSnapshotRequest) (*DeleteSnapshotResponse, error)
    ListSnapshots(context.Context, *ListSnapshotsRequest) (*ListSnapshotsResponse, error)
    ControllerExpandVolume(context.Context, *ControllerExpandVolumeRequest) (*ControllerExpandVolumeResponse, error)
    ControllerGetVolume(context.Context, *ControllerGetVolumeRequest) (*ControllerGetVolumeResponse, error)
}

// NodeServer is the server API for Node service.
type NodeServer interface {
    NodeStageVolume(context.Context, *NodeStageVolumeRequest) (*NodeStageVolumeResponse, error)
    NodeUnstageVolume(context.Context, *NodeUnstageVolumeRequest) (*NodeUnstageVolumeResponse, error)
    NodePublishVolume(context.Context, *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error)
    NodeUnpublishVolume(context.Context, *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error)
    NodeGetVolumeStats(context.Context, *NodeGetVolumeStatsRequest) (*NodeGetVolumeStatsResponse, error)
    NodeExpandVolume(context.Context, *NodeExpandVolumeRequest) (*NodeExpandVolumeResponse, error)
    NodeGetCapabilities(context.Context, *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error)
    NodeGetInfo(context.Context, *NodeGetInfoRequest) (*NodeGetInfoResponse, error)
}

// IdentityServer is the server API for Identity service.
type IdentityServer interface {
    GetPluginInfo(context.Context, *GetPluginInfoRequest) (*GetPluginInfoResponse, error)
    GetPluginCapabilities(context.Context, *GetPluginCapabilitiesRequest) (*GetPluginCapabilitiesResponse, error)
    Probe(context.Context, *ProbeRequest) (*ProbeResponse, error)
}
```
上面接口列表是由[仓库](https://github.com/container-storage-interface/spec/blob/master/lib/go/csi/csi.pb.go) 中获取，用户需要实现分别这些接口，
即可实现自定义CSI插件。

除上述的接口需要实现外，在部署CSI应用时，也需要采用deployment statefulSet daemonSet这类的编排容器部署。
官方提供 External 组件，作为k8s api跟csi driver的桥梁 [仓库](https://github.com/orgs/kubernetes-csi/repositories?type=all)：
- node-driver-registrar

CSI node-driver-registrar是一个sidecar容器，可从CSI driver获取驱动程序信息（使用NodeGetInfo），并使用kubelet插件注册机制在该节点上的kubelet中对其进行注册。


- external-attacher

它是一个sidecar容器，用于监视Kubernetes VolumeAttachment对象并针对驱动程序端点触发CSI ControllerPublish和ControllerUnpublish操作


- external-provisioner

它是一个sidecar容器，用于监视Kubernetes PersistentVolumeClaim对象并针对驱动程序端点触发CSI CreateVolume和DeleteVolume操作。
external-attacher还支持快照数据源。 如果将快照CRD资源指定为PVC对象上的数据源，则此sidecar容器通过获取SnapshotContent对象获取有关快照的信息，并填充数据源字段，该字段向存储系统指示应使用指定的快照填充新卷 。


- external-resizer

它是一个sidecar容器，用于监视Kubernetes API服务器上的PersistentVolumeClaim对象的改动，如果用户请求在PersistentVolumeClaim对象上请求更多存储，则会针对CSI端点触发ControllerExpandVolume操作。


- external-snapshotter

它是一个sidecar容器，用于监视Kubernetes API服务器上的VolumeSnapshot和VolumeSnapshotContent CRD对象。创建新的VolumeSnapshot对象（引用与此驱动程序对应的SnapshotClass CRD对象）将导致sidecar容器提供新的快照。该Sidecar侦听指示成功创建VolumeSnapshot的服务，并立即创建VolumeSnapshotContent资源。


- livenessprobe

它是一个sidecar容器，用于监视CSI驱动程序的运行状况，并通过Liveness Probe机制将其报告给Kubernetes。 这使Kubernetes能够自动检测驱动程序问题并重新启动Pod以尝试解决问题。