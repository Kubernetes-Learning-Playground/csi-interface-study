### nfs-csi-practice

#### 使用官方推荐的nfs csi
[项目地址](https://github.com/kubernetes-csi/csi-driver-nfs)

1. 除了目标k8s集群外，需要再找一台可以与此k8s集群互通的服务器，并先安装nfs服务 [参考](https://cloud.tencent.com/developer/article/1720669)

2. 开启网络端口 [参考](https://www.onitroad.com/jc/linux/other/nfs-server-default-port.html)

3. 安装对应的组件
```bash
[root@vm-0-12-centos csi-nfs-practice]# kubectl apply -f yamls/csi4.1
deployment.apps/csi-nfs-controller unchanged
csidriver.storage.k8s.io/nfs.csi.k8s.io unchanged
daemonset.apps/csi-nfs-node unchanged
serviceaccount/csi-nfs-controller-sa unchanged
serviceaccount/csi-nfs-node-sa unchanged
clusterrole.rbac.authorization.k8s.io/nfs-external-provisioner-role unchanged
clusterrolebinding.rbac.authorization.k8s.io/nfs-csi-provisioner-binding unchanged
```
4. 安装 StorageClass 与启动 deployment 等