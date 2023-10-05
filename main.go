package main

import (
	"flag"
	"github/mycsi/csi_practice/pkg/service"
	"k8s.io/klog/v2"
)

var (
	nodeID     = ""
	endpoint   = ""
	driverName = ""
)

const (
	// DefaultDriverName csi 默认插件名称，使用 mycsi.practice.com
	DefaultDriverName = "mycsi.practice.com"
	DefaultEndpoint   = "unix:///csi/csi.sock"
	DefaultNodeID     = "node1"
)

func main() {

	// 节点名由外部注入
	flag.StringVar(&nodeID, "nodeid", DefaultNodeID, "--nodeid=xxx, node name")
	flag.StringVar(&endpoint, "endpoint", DefaultEndpoint, "--endpoint=xxx, CSI endpoint, ex: unix:///csi/csi.sock")
	flag.StringVar(&driverName, "drivername", DefaultDriverName, "--drivername=xxx, csi driver name")
	klog.InitFlags(nil)
	flag.Parse()

	// TODO : 改成配置文件
	driverOptions := service.MyDriverOptions{
		NodeID:     nodeID,
		DriverName: driverName,
		Endpoint:   endpoint,
	}
	klog.Info("bootstrap start...")
	driver := service.NewMyDriver(&driverOptions)
	driver.Start()
}
