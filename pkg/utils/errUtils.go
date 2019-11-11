package utils

import "k8s.io/klog"

func HandleErr(err error) {
	if err != nil {
		klog.Error(err)
		panic(err)
	}
}
