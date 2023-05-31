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
	DefaultDriverName = "mycsi.practice.com"
)

func main() {

	// 节点名由外部注入
	flag.StringVar(&nodeID, "nodeid", "", "--nodeid=xxx")
	flag.StringVar(&endpoint, "endpoint", "unix:///csi/csi.sock", "CSI endpoint")
	flag.StringVar(&driverName, "drivername", DefaultDriverName, "name of the driver")
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
