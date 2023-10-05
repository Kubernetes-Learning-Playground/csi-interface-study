## csi-interface-study CSI接口练习

### CSI 基础知识
类似于 CRI，CSI 也是基于 gRPC 实现。详细的 CSI SPEC 可以参考 [这里](https://github.com/container-storage-interface)，它要求插件开发者要实现三个 gRPC 服务：
- Identity Server：用于 Kubernetes 与 CSI 插件协调版本信息
- Controller Server：用于创建、删除以及管理 Volume 存储卷
- Node Server：用于将 Volume 存储卷挂载到指定的目录中以便 Kubelet 创建容器时使用(需要监听在 /var/lib/kubelet/plugins/[SanitizedCSIDriverName]/csi.sock)
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
即可实现自定义 CSI 插件。

除上述的接口需要实现外，在部署 CSI 应用时，也需要采用 deployment statefulSet daemonSet 这类的编排容器部署。
官方提供 External 组件，作为 k8s 跟 csi-driver 的桥梁 [仓库](https://github.com/orgs/kubernetes-csi/repositories?type=all)：
- node-driver-registrar

CSI node-driver-registrar 是一个 sidecar 容器，可从 CSI driver 获取信息(使用 NodeGetInfo)，并使用 kubelet 插件注册机制在该节点上的 kubelet 中对其进行注册。


- external-attacher

它是一个 sidecar 容器，用于监视 Kubernetes VolumeAttachment 对象并针对 CSI 插件触发 **ControllerPublish** 和 **ControllerUnpublish** 操作


- external-provisioner

它是一个 sidecar 容器，用于监听 Kubernetes PersistentVolumeClaim 对象并针对 CSI 插件触发 **CreateVolume** 和 **DeleteVolume** 操作。
external-attacher 还支持快照数据源。如果将快照 CRD 资源指定为 PVC 对象上的数据源，则此 sidecar 容器通过获取 **SnapshotContent** 对象获取有关快照的信息，并填充数据源字段，该字段向存储系统指示应使用指定的快照填充新卷。


- external-resizer

它是一个 sidecar 容器，用于监听 Kubernetes api-server 上的 PersistentVolumeClaim 对象，如果用户请求在 PersistentVolumeClaim 对象上请求更多存储，则会针对 CSI 插件触发 **ControllerExpandVolume** 操作。


- external-snapshotter

它是一个 sidecar 容器，用于监听 Kubernetes api-server 上的 **VolumeSnapshot**和**VolumeSnapshotContent** CRD 对象。创建新的 **VolumeSnapshot**对象（引用与此驱动程序对应的 SnapshotClass CRD 对象）将导致 sidecar 容器提供新的快照。该 Sidecar 侦听指示成功创建 VolumeSnapshot 的服务，并立即创建 VolumeSnapshotContent 资源。


- livenessprobe

它是一个 sidecar 容器，用于监视CSI驱动程序的运行状况，并通过 Liveness Probe 机制将其报告给 Kubernetes。 这使 Kubernetes 能够自动检测驱动程序问题并重新启动 Pod 以尝试解决问题。

#### 项目部署
- docker 镜像
```bash
[root@vm-0-12-centos mycsi]# docker build -t mycsi:v1 .
Sending build context to Docker daemon  13.29MB
Step 1/15 : FROM golang:1.18.7-alpine3.15 as builder
 ---> 33c97f935029
Step 2/15 : WORKDIR /app
 ---> Using cache
```
- 部署 k8s
```bash
[root@vm-0-12-centos mycsi]# kubectl apply -f deploy/driver.yaml
deployment.apps/mycsi unchanged
```