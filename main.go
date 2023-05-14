package main

import (
	"flag"
	"golanglearning/new_project/csi_practice/pkg/bootstrap"
	"k8s.io/klog/v2"
)

var (
	nodeID = ""
)

func main() {

	// 节点名由外部注入
	flag.StringVar(&nodeID, "nodeid", "", "--nodeid=xxx")
	klog.InitFlags(nil)
	flag.Parse()

	driver := bootstrap.NewMyDriver(nodeID)
	driver.Start()
}
