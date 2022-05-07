package k8sutils

import (
	"context"
	"fmt"
	"sort"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisv1alpha1 "redis-operator/api/v1alpha1"
)

const (
	redisExporterContainer = "redis-exporter"
	graceTime              = 15
)

const (
	STSReplicaNum = 2 // HA规格定死一主一从
)

// statefulSetParameters will define statefulsets input params
type statefulSetParameters struct {
	Replicas              int32
	Metadata              metav1.ObjectMeta
	NodeSelector          map[string]string
	SecurityContext       *corev1.PodSecurityContext
	PriorityClassName     string
	Affinity              *corev1.Affinity
	Tolerations           *[]corev1.Toleration
	EnableMetrics         bool
	PersistentVolumeClaim corev1.PersistentVolumeClaim
	ExternalConfig        *string
}

// containerParameters will define container input params
type containerParameters struct {
	Image                        string
	ImagePullPolicy              corev1.PullPolicy
	Resources                    *corev1.ResourceRequirements
	RedisExporterImage           string
	RedisExporterImagePullPolicy corev1.PullPolicy
	RedisExporterResources       *corev1.ResourceRequirements
	RedisExporterEnv             *[]corev1.EnvVar
	Role                         string
	EnabledPassword              *bool
	SecretName                   *string
	SecretKey                    *string
	PersistenceEnabled           *bool
	ReadinessProbe               *corev1.Probe
	LivenessProbe                *corev1.Probe
}

// CreateOrUpdateStateFul method will create or update Redis service
func CreateOrUpdateStateFul(namespace string, stsMeta metav1.ObjectMeta, params statefulSetParameters, ownerDef metav1.OwnerReference, containerParams containerParameters) error {
	logger := statefulSetLogger(namespace, stsMeta.Name)
	storedStateful, err := GetStatefulSet(namespace, stsMeta.Name)
	statefulSetDef := generateStatefulSetsDef(stsMeta, params, ownerDef, containerParams)
	if err != nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(statefulSetDef); err != nil {
			logger.Error(err, "Unable to patch redis statefulset with comparison object")
			return err
		}
		if errors.IsNotFound(err) {
			return createStatefulSet(namespace, statefulSetDef)
		}
		return err
	}
	return patchStatefulSet(storedStateful, statefulSetDef, namespace)
}

// patchStateFulSet will patch Redis Kubernetes StateFulSet
func patchStatefulSet(storedStateful *appsv1.StatefulSet, newStateful *appsv1.StatefulSet, namespace string) error {
	logger := statefulSetLogger(namespace, storedStateful.Name)

	// We want to try and keep this atomic as possible.
	newStateful.ResourceVersion = storedStateful.ResourceVersion
	newStateful.CreationTimestamp = storedStateful.CreationTimestamp
	newStateful.ManagedFields = storedStateful.ManagedFields

	patchResult, err := patch.DefaultPatchMaker.Calculate(storedStateful, newStateful,
		patch.IgnoreStatusFields(),
		patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus(),
		patch.IgnoreField("kind"),
		patch.IgnoreField("apiVersion"),
	)
	if err != nil {
		logger.Error(err, "Unable to patch redis statefulset with comparison object")
		return err
	}
	if !patchResult.IsEmpty() {
		logger.Info("Changes in statefulset Detected, Updating...", "patch", string(patchResult.Patch))
		// Field is immutable therefore we MUST keep it as is.
		newStateful.Spec.VolumeClaimTemplates = storedStateful.Spec.VolumeClaimTemplates
		for key, value := range storedStateful.Annotations {
			if _, present := newStateful.Annotations[key]; !present {
				newStateful.Annotations[key] = value
			}
		}
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newStateful); err != nil {
			logger.Error(err, "Unable to patch redis statefulset with comparison object")
			return err
		}
		return updateStatefulSet(namespace, newStateful)
	}
	logger.Info("Reconciliation Complete, no Changes required.")
	return nil
}

// generateStatefulSetsDef generates the statefulsets definition of Redis
func generateStatefulSetsDef(stsMeta metav1.ObjectMeta, params statefulSetParameters, ownerDef metav1.OwnerReference, containerParams containerParameters) *appsv1.StatefulSet {
	statefulset := &appsv1.StatefulSet{
		TypeMeta:   generateMetaInformation("StatefulSet", "apps/v1"),
		ObjectMeta: stsMeta,
		Spec: appsv1.StatefulSetSpec{
			Selector:    LabelSelectors(stsMeta.GetLabels()),
			ServiceName: stsMeta.Name,
			Replicas:    &params.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      stsMeta.GetLabels(),
					Annotations: generateStatefulSetsAnots(stsMeta),
					// Annotations: stsMeta.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers:        generateContainerDef(stsMeta.GetName(), containerParams, params.EnableMetrics, params.ExternalConfig),
					NodeSelector:      params.NodeSelector,
					SecurityContext:   params.SecurityContext,
					PriorityClassName: params.PriorityClassName,
					Affinity:          params.Affinity,
				},
			},
		},
	}
	if params.Tolerations != nil {
		statefulset.Spec.Template.Spec.Tolerations = *params.Tolerations
	}
	if containerParams.PersistenceEnabled != nil && *containerParams.PersistenceEnabled {
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, createPVCTemplate(stsMeta, params.PersistentVolumeClaim))
	}
	if params.ExternalConfig != nil {
		statefulset.Spec.Template.Spec.Volumes = getExternalConfig(*params.ExternalConfig)
	}
	AddOwnerRefToObject(statefulset, ownerDef)
	return statefulset
}

// getExternalConfig will return the redis external configuration
func getExternalConfig(configMapName string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "external-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		},
	}
}

// createPVCTemplate will create the persistent volume claim template
func createPVCTemplate(stsMeta metav1.ObjectMeta, storageSpec corev1.PersistentVolumeClaim) corev1.PersistentVolumeClaim {
	pvcTemplate := storageSpec
	pvcTemplate.CreationTimestamp = metav1.Time{}
	pvcTemplate.Name = stsMeta.GetName()
	pvcTemplate.Labels = stsMeta.GetLabels()
	// We want the same annoations as the StatefulSet here
	pvcTemplate.Annotations = generateStatefulSetsAnots(stsMeta)
	if storageSpec.Spec.AccessModes == nil {
		pvcTemplate.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	} else {
		pvcTemplate.Spec.AccessModes = storageSpec.Spec.AccessModes
	}
	pvcVolumeMode := corev1.PersistentVolumeFilesystem
	if storageSpec.Spec.VolumeMode != nil {
		pvcVolumeMode = *storageSpec.Spec.VolumeMode
	}
	pvcTemplate.Spec.VolumeMode = &pvcVolumeMode
	pvcTemplate.Spec.Resources = storageSpec.Spec.Resources
	pvcTemplate.Spec.Selector = storageSpec.Spec.Selector
	return pvcTemplate
}

// generateContainerDef generates container definition for Redis
func generateContainerDef(name string, containerParams containerParameters, enableMetrics bool, externalConfig *string) []corev1.Container {
	containerDefinition := []corev1.Container{
		{
			Name:            name,
			Image:           containerParams.Image,
			ImagePullPolicy: containerParams.ImagePullPolicy,
			Env: getEnvironmentVariables(
				containerParams.Role,
				containerParams.EnabledPassword,
				containerParams.SecretName,
				containerParams.SecretKey,
				containerParams.PersistenceEnabled,
				containerParams.RedisExporterEnv,
			),
			Resources:      *containerParams.Resources,
			ReadinessProbe: getProbeInfo(),
			LivenessProbe:  getProbeInfo(),
			VolumeMounts:   getVolumeMount(name, containerParams.PersistenceEnabled, externalConfig),
		},
	}
	if containerParams.ReadinessProbe != nil {
		containerDefinition[0].ReadinessProbe = containerParams.ReadinessProbe
	} else {
		containerDefinition[0].ReadinessProbe = getProbeInfo()
	}
	if containerParams.LivenessProbe != nil {
		containerDefinition[0].LivenessProbe = containerParams.LivenessProbe
	} else {
		containerDefinition[0].LivenessProbe = getProbeInfo()
	}

	if containerParams.Resources != nil {
		containerDefinition[0].Resources = *containerParams.Resources
	}
	if enableMetrics {
		containerDefinition = append(containerDefinition, enableRedisMonitoring(containerParams))
	}

	return containerDefinition
}

// enableRedisMonitoring will add Redis Exporter as sidecar container
func enableRedisMonitoring(params containerParameters) corev1.Container {
	exporterDefinition := corev1.Container{
		Name:            redisExporterContainer,
		Image:           params.RedisExporterImage,
		ImagePullPolicy: params.RedisExporterImagePullPolicy,
		Env: getEnvironmentVariables(
			params.Role,
			params.EnabledPassword,
			params.SecretName,
			params.SecretKey,
			params.PersistenceEnabled,
			params.RedisExporterEnv,
		),
		Resources:    *params.RedisExporterResources,
		VolumeMounts: getVolumeMount("", nil, nil), // We need/want the tls-certs but we DON'T need the PVC (if one is available)
	}
	if params.RedisExporterResources != nil {
		exporterDefinition.Resources = *params.RedisExporterResources
	}
	return exporterDefinition
}

// getVolumeMount gives information about persistence mount
func getVolumeMount(name string, persistenceEnabled *bool, externalConfig *string) []corev1.VolumeMount {
	var VolumeMounts []corev1.VolumeMount

	if persistenceEnabled != nil && *persistenceEnabled {
		VolumeMounts = append(VolumeMounts, corev1.VolumeMount{
			Name:      name,
			MountPath: "/data",
		})
	}

	if externalConfig != nil {
		VolumeMounts = append(VolumeMounts, corev1.VolumeMount{
			Name:      "external-config",
			MountPath: "/etc/redis/external.conf.d",
		})
	}

	return VolumeMounts
}

// getProbeInfo generates probe information for Redis
func getProbeInfo() *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: graceTime,
		PeriodSeconds:       15,
		FailureThreshold:    5,
		TimeoutSeconds:      5,
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/usr/bin/redis-cli",
					"-p",
					"6379",
					"ping",
				},
			},
		},
	}
}

// getEnvironmentVariables returns all the required Environment Variables
func getEnvironmentVariables(role string, enabledPassword *bool, secretName *string, secretKey *string, persistenceEnabled *bool, extraEnv *[]corev1.EnvVar) []corev1.EnvVar {
	logger := statefulSetLogger("namespace", "sts")

	envVars := []corev1.EnvVar{
		{Name: "SERVER_MODE", Value: role},
		{Name: "SETUP_MODE", Value: role},
	}

	// etcd addr
	// TODO 
	// etcdAddr, err := GetKubeSystemEtcdEndpoints()
	etcdAddr := "192.168.58.201:2379,192.168.58.201:2379"
	var err error 
	if nil != err {
		logger.Error(err, fmt.Sprintf("GetKubeSystemEtcdEndpoints err: %v", err))
	} else {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ETCD_ADDR",
			Value: etcdAddr,
		})
	}

	redisHost := "redis://localhost:6379"
	envVars = append(envVars, corev1.EnvVar{
		Name:  "REDIS_ADDR",
		Value: redisHost,
	})

	if enabledPassword != nil && *enabledPassword {
		envVars = append(envVars, corev1.EnvVar{
			Name: "REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: *secretName,
					},
					Key: *secretKey,
				},
			},
		})
	}
	if persistenceEnabled != nil && *persistenceEnabled {
		envVars = append(envVars, corev1.EnvVar{Name: "PERSISTENCE_ENABLED", Value: "true"})
	}

	if extraEnv != nil {
		envVars = append(envVars, *extraEnv...)
	}
	sort.SliceStable(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
	return envVars
}

// createStatefulSet is a method to create statefulset in Kubernetes
func createStatefulSet(namespace string, stateful *appsv1.StatefulSet) error {
	logger := statefulSetLogger(namespace, stateful.Name)
	_, err := generateK8sClient().AppsV1().StatefulSets(namespace).Create(context.TODO(), stateful, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Redis stateful creation failed, ns: %s, sts: %s, err: %v", namespace, stateful.Name, err))
		return err
	}
	logger.Info(fmt.Sprintf("Redis stateful successfully created, ns: %s, sts: %s", namespace, stateful.Name))
	return nil
}

// updateStatefulSet is a method to update statefulset in Kubernetes
func updateStatefulSet(namespace string, stateful *appsv1.StatefulSet) error {
	logger := statefulSetLogger(namespace, stateful.Name)
	// logger.Info(fmt.Sprintf("Setting Statefulset to the following: %s", stateful))
	_, err := generateK8sClient().AppsV1().StatefulSets(namespace).Update(context.TODO(), stateful, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "Redis stateful update failed")
		return err
	}
	logger.Info("Redis stateful successfully updated ")
	return nil
}

// GetStateFulSet is a method to get statefulset in Kubernetes
func GetStatefulSet(namespace string, stateful string) (*appsv1.StatefulSet, error) {
	logger := statefulSetLogger(namespace, stateful)
	getOpts := metav1.GetOptions{
		TypeMeta: generateMetaInformation("StatefulSet", "apps/v1"),
	}
	statefulInfo, err := generateK8sClient().AppsV1().StatefulSets(namespace).Get(context.TODO(), stateful, getOpts)
	if err != nil {
		logger.Info("Redis statefulset get action failed")
		return nil, err
	}
	logger.Info("Redis statefulset get action was successful")
	return statefulInfo, nil
}

// statefulSetLogger will generate logging interface for Statfulsets
func statefulSetLogger(namespace string, name string) logr.Logger {
	reqLogger := log.WithValues("Request.StatefulSet.Namespace", namespace, "Request.StatefulSet.Name", name)
	return reqLogger
}

func CheckStatefulSetExist(namespace string, stateful string) (bool, error) {
	logger := statefulSetLogger(namespace, stateful)
	_, err := GetStatefulSet(namespace, stateful)
	if nil != err {
		if errors.IsNotFound(err) {
			return false, nil
		}

		logger.Error(err, fmt.Sprintf("CreateStatefulSet fail, CheckStatefulSetExist fail, err: %v", err))
		return false, err
	}

	return true, nil
}

// generateRedisStandalone generates Redis standalone information
func generateRedisParams(cr *redisv1alpha1.OperatorRedisHA, replicas int32, affinity *corev1.Affinity) statefulSetParameters {
	res := statefulSetParameters{
		Metadata:     cr.ObjectMeta,
		Replicas:     replicas,
		NodeSelector: cr.Spec.NodeSelector,
		Affinity:     affinity,
		Tolerations:  cr.Spec.Tolerations,
	}

	if cr.Spec.Storage != nil {
		res.PersistentVolumeClaim = cr.Spec.Storage.VolumeClaimTemplate
	}

	return res
}

// generateRedisStandaloneContainerParams generates Redis container information
func generateRedisClusterContainerParams(cr *redisv1alpha1.OperatorRedisHA) containerParameters {
	trueProperty := true
	containerProp := containerParameters{
		Image:           cr.Spec.KubernetesConfig.Image,
		ImagePullPolicy: cr.Spec.KubernetesConfig.ImagePullPolicy,
		Resources:       cr.Spec.KubernetesConfig.Resources,
	}

	if cr.Spec.Storage != nil {
		containerProp.PersistenceEnabled = &trueProperty
	}
	return containerProp
}

func CreateStatefulSet(cr *redisv1alpha1.OperatorRedisHA) error {
	logger := statefulSetLogger(cr.Namespace, cr.ObjectMeta.Name)

	isExist, err := CheckStatefulSetExist(cr.Namespace, cr.ObjectMeta.Name)
	if nil != err {
		logger.Error(err, fmt.Sprintf("CreateStatefulSet fail, CheckStatefulSetExist fail, err: %v", err))
		return err
	}

	if isExist {
		return nil
	}

	stateFulName := cr.ObjectMeta.Name
	labels := getRedisLabels(stateFulName, cr.ObjectMeta.Labels)
	annotations := generateStatefulSetsAnots(cr.ObjectMeta)
	objectMetaInfo := generateObjectMetaInformation(stateFulName, cr.Namespace, labels, annotations)
	params := generateRedisParams(cr, STSReplicaNum, cr.Spec.Affinity)
	ownerDef := redisClusterAsOwner(cr)
	containerParams := generateRedisClusterContainerParams(cr)
	statefulSetDef := generateStatefulSetsDef(objectMetaInfo, params, ownerDef, containerParams)

	return createStatefulSet(cr.Namespace, statefulSetDef)
}

func CheckStatefulSetPods(cr *redisv1alpha1.OperatorRedisHA) (bool, error) {
	namespace := cr.Namespace
	stateful := cr.ObjectMeta.Name

	logger := statefulSetLogger(namespace, stateful)
	statefulSec, err := GetStatefulSet(namespace, stateful)
	if nil != err {
		if errors.IsNotFound(err) {
			return false, nil
		}

		logger.Error(err, fmt.Sprintf("CheckStatefulSetPods fail, CheckStatefulSetExist fail, err: %v", err))
		return false, err
	}

	// TODO: 需要继续检查pod是否为running
	if STSReplicaNum != statefulSec.Status.CurrentReplicas {
		logger.Info(fmt.Sprintf("WARN: CheckStatefulSetPods fail, CurrentReplicas[%d] not STSReplicaNum[%d]",
			statefulSec.Status.CurrentReplicas, STSReplicaNum))
		return false, nil
	}

	return true, nil
}
