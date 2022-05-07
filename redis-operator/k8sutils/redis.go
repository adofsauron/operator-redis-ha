package k8sutils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"encoding/base64"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	appsv1alpha1 "redis-operator/api/v1alpha1"
	"redis-operator/ctlconfig"
)

const (
	ETCD_FILE_CERT   = "/usr/bin/etcd-client.crt"
	ETCD_FILE_KEY    = "/usr/bin/etcd-key.crt"
	ETCD_FILE_CACERT = "/usr/bin/etcd-ca.crt"
)

type RedisDetails struct {
	PodName   string
	Namespace string
}

// generateRedisManagerLogger will generate logging interface for Redis operations
func generateRedisManagerLogger(namespace, name string) logr.Logger {
	reqLogger := log.WithValues("Request.RedisManager.Namespace", namespace, "Request.RedisManager.Name", name)
	return reqLogger
}

// getContainerID will return the id of container from pod
func getContainerID(cr *appsv1alpha1.OperatorRedisHA, podName string) (int, *corev1.Pod) {
	logger := generateRedisManagerLogger(cr.Namespace, cr.ObjectMeta.Name)
	pod, err := generateK8sClient().CoreV1().Pods(cr.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Could not get pod info")
	}

	targetContainer := -1
	for containerID, tr := range pod.Spec.Containers {
		logger.Info("Pod Counted successfully", "Count", containerID, "Container Name", tr.Name)
		if tr.Name == cr.ObjectMeta.Name {
			targetContainer = containerID
			break
		}
	}
	return targetContainer, pod
}

// executeCommand will execute the commands in pod
func executeCommand(cr *appsv1alpha1.OperatorRedisHA, cmd []string, podName string) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)
	logger := generateRedisManagerLogger(cr.Namespace, cr.ObjectMeta.Name)
	config, err := generateK8sConfig()
	if err != nil {
		logger.Error(err, "Could not find pod to execute")
		return "", err
	}
	targetContainer, pod := getContainerID(cr, podName)
	if targetContainer < 0 {
		logger.Error(err, "Could not find pod to execute")
		return "", errors.New("Could not find pod to execute")
	}

	req := generateK8sClient().CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(cr.Namespace).SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Container: pod.Spec.Containers[targetContainer].Name,
		Command:   cmd,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		logger.Error(err, "Failed to init executor")
		return "", err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &execOut,
		Stderr: &execErr,
		Tty:    false,
	})
	if err != nil {
		logger.Error(err, "Could not execute command", "Command", cmd, "Output", execOut.String(), "Error", execErr.String())
		return "", err
	}

	logger.Info("Successfully executed the command", "Command", cmd, "Output", execOut.String())
	return execOut.String(), nil
}

// getRedisServerIP will return the IP of redis service
func getRedisServerIP(redisInfo RedisDetails) string {
	logger := generateRedisManagerLogger(redisInfo.Namespace, redisInfo.PodName)
	redisPod, err := generateK8sClient().CoreV1().Pods(redisInfo.Namespace).Get(context.TODO(), redisInfo.PodName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Error in getting redis pod IP")
	}

	redisIP := redisPod.Status.PodIP
	// If we're NOT IPv4, assume were IPv6..
	if net.ParseIP(redisIP).To4() == nil {
		logger.Info("Redis is IPv6", "ip", redisIP, "ipv6", net.ParseIP(redisIP).To16())
		redisIP = fmt.Sprintf("[%s]", redisIP)
	}

	logger.Info("Successfully got the ip for redis", "ip", redisIP)
	return redisIP
}

func ExecuteRedisCreateHA(cr *appsv1alpha1.OperatorRedisHA) error {
	logger := generateRedisManagerLogger(cr.Namespace, cr.ObjectMeta.Name)

	cmd := []string{"redis-cli", "-p", "6379", "slaveof"}
	pod := RedisDetails{
		PodName:   cr.ObjectMeta.Name + "-0",
		Namespace: cr.Namespace,
	}
	cmd = append(cmd, getRedisServerIP(pod))
	cmd = append(cmd, "6379")

	logger.Info("Redis HA creation command is", "Command", cmd)
	out, err := executeCommand(cr, cmd, cr.ObjectMeta.Name+"-1") // pod-1 slaveof pod-0
	if nil != err {
		return err
	}

	if !strings.Contains(out, "OK") {
		err := errors.New("ExecuteRedisCreateHA fail, out not ok")
		logger.Error(err, fmt.Sprintf("ExecuteRedisCreateHA fail, out not ok, out: %s", out))
		return err
	}

	return nil
}

func ExecuteRedisSetEtcdCrd(cr *appsv1alpha1.OperatorRedisHA) error {
	logger := generateRedisManagerLogger(cr.Namespace, cr.ObjectMeta.Name)

	cfg := ctlconfig.GetconfigParam()

	strbytes_cert := []byte(cfg.ETCD_VALUE_CERT)
	encoded_ceret := base64.StdEncoding.EncodeToString(strbytes_cert)

	strbytes_key := []byte(cfg.ETCD_VALUE_KEY)
	encoded_key := base64.StdEncoding.EncodeToString(strbytes_key)

	strbytes_cacert := []byte(cfg.ETCD_VALUE_CACERT)
	encoded_caceret := base64.StdEncoding.EncodeToString(strbytes_cacert)


	cmd := []string{}
	cmd = append(cmd, "/usr/bin/etcd-save-crt.sh", encoded_ceret, encoded_key, encoded_caceret)

	logger.Info(fmt.Sprintf("ExecuteRedisSetEtcdCrd cmd: %s", cmd))

	for i := 0; i <= 1; i++ {
		podName := cr.ObjectMeta.Name + "-" + strconv.Itoa(i)
		_, err := executeCommand(cr, cmd, podName)
		if nil != err {
			logger.Error(err, fmt.Sprintf("ExecuteRedisSetEtcdCrd fail, executeCommand err: %v", err))
			return err
		}
	}

	return nil
}
