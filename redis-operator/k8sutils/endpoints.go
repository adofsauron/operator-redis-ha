package k8sutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NS_KUBE_SYSTEM       = "kube-system"
	SVC_KUBE_SYSTEM_ETCD = "coc-kube-etcd"
)

func epLogger(namespace string, name string) logr.Logger {
	reqLogger := log.WithValues("Request.Service.Namespace", namespace, "Request.Endpoints.Name", name)
	return reqLogger
}

func getEndpoints(namespace string, ep string) (*corev1.Endpoints, error) {
	logger := epLogger(namespace, ep)
	getOpts := metav1.GetOptions{
		TypeMeta: generateMetaInformation("Endpoints", "v1"),
	}
	epInfo, err := generateK8sClient().CoreV1().Endpoints(namespace).Get(context.TODO(), ep, getOpts)
	if err != nil {
		logger.Info(fmt.Sprintf("get endpoints is failed: %v", err))
		return nil, err
	}
	return epInfo, nil
}

func GetKubeSystemEtcdEndpoints() (string, error) {
	logger := serviceLogger(NS_KUBE_SYSTEM, SVC_KUBE_SYSTEM_ETCD)

	epEtcd, err := getEndpoints(NS_KUBE_SYSTEM, SVC_KUBE_SYSTEM_ETCD)
	if err != nil {
		logger.Error(err, fmt.Sprintf("GetKubeSystemEtcdSVC fail, err: %v", err))
		return "", err
	}

	if 0 >= len(epEtcd.Subsets) {
		logger.Info(fmt.Sprintf("ERROR: GetKubeSystemEtcdEndpoints fail, 0 >= len(epEtcd.Subsets[0]"))
		return "", errors.New("0 >= len(epEtcd.Subsets[0]")
	}

	if 0 >= len(epEtcd.Subsets[0].Ports) {
		logger.Info(fmt.Sprintf("ERROR: GetKubeSystemEtcdEndpoints fail, 0 >= len(epEtcd.Subsets[0].Ports"))
		return "", errors.New("0 >= len(epEtcd.Subsets[0].Ports")
	}

	if 0 >= len(epEtcd.Subsets[0].Addresses) {
		logger.Info(fmt.Sprintf("ERROR: GetKubeSystemEtcdEndpoints fail, 0 >= len(epEtcd.Subsets[0].Addresses"))
		return "", errors.New("0 >= len(epEtcd.Subsets[0].Addresses")
	}

	addr := ""
	port := epEtcd.Subsets[0].Ports[0].Port
	for _, set := range epEtcd.Subsets[0].Addresses {
		ip := set.IP

		ip_port := fmt.Sprintf("%s:%d", ip, port)
		addr = addr + ip_port
		addr = addr + ","
	}

	addrLen := len(addr)
	addr = addr[:addrLen-1]
	return addr, nil
}
