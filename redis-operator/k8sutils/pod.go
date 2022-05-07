package k8sutils

import (
	"github.com/go-logr/logr"
)

func podLogger(namespace string, name string) logr.Logger {
	reqLogger := log.WithValues("Request.Service.Namespace", namespace, "Request.Service.Name", name)
	return reqLogger
}

func PodTestCp() {

}
