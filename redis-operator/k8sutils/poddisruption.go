package k8sutils

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateService method will create or update Redis service
func CreateOrUpdatePodDisruptionBudget(pdbDef *policyv1.PodDisruptionBudget) error {
	logger := pdbLogger(pdbDef.Namespace, pdbDef.Name)
	storedPDB, err := GetPodDisruptionBudget(pdbDef.Namespace, pdbDef.Name)
	if err != nil {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(pdbDef); err != nil {
			logger.Error(err, "Unable to patch redis PodDisruptionBudget with comparison object")
			return err
		}
		if errors.IsNotFound(err) {
			return createPodDisruptionBudget(pdbDef.Namespace, pdbDef)
		}
		return err
	}
	return patchPodDisruptionBudget(storedPDB, pdbDef, pdbDef.Namespace)
}

// patchPodDisruptionBudget will patch Redis Kubernetes PodDisruptionBudgets
func patchPodDisruptionBudget(storedPdb *policyv1.PodDisruptionBudget, newPdb *policyv1.PodDisruptionBudget, namespace string) error {
	logger := pdbLogger(namespace, storedPdb.Name)
	// We want to try and keep this atomic as possible.
	newPdb.ResourceVersion = storedPdb.ResourceVersion
	newPdb.CreationTimestamp = storedPdb.CreationTimestamp
	newPdb.ManagedFields = storedPdb.ManagedFields

	// newPdb.Kind = "PodDisruptionBudget"
	// newPdb.APIVersion = "policy/v1"
	storedPdb.Kind = "PodDisruptionBudget"
	storedPdb.APIVersion = "policy/v1"

	patchResult, err := patch.DefaultPatchMaker.Calculate(storedPdb, newPdb,
		patch.IgnorePDBSelector(),
		patch.IgnoreStatusFields(),
	)
	if err != nil {
		logger.Error(err, "Unable to patch redis PodDisruption with comparison object")
		return err
	}
	if !patchResult.IsEmpty() {
		logger.Info("Changes in PodDisruptionBudget Detected, Updating...",
			"patch", string(patchResult.Patch),
			"Current", string(patchResult.Current),
			"Original", string(patchResult.Original),
			"Modified", string(patchResult.Modified),
		)
		for key, value := range storedPdb.Annotations {
			if _, present := newPdb.Annotations[key]; !present {
				newPdb.Annotations[key] = value
			}
		}
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newPdb); err != nil {
			logger.Error(err, "Unable to patch redis PodDisruptionBudget with comparison object")
			return err
		}
		return updatePodDisruptionBudget(namespace, newPdb)
	}
	return nil
}

// createPodDisruptionBudget is a method to create PodDisruptionBudgets in Kubernetes
func createPodDisruptionBudget(namespace string, pdb *policyv1.PodDisruptionBudget) error {
	logger := pdbLogger(namespace, pdb.Name)
	_, err := generateK8sClient().PolicyV1().PodDisruptionBudgets(namespace).Create(context.TODO(), pdb, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "Redis PodDisruptionBudget creation failed")
		return err
	}
	logger.Info("Redis PodDisruptionBudget creation was successful")
	return nil
}

// updatePodDisruptionBudget is a method to update PodDisruptionBudgets in Kubernetes
func updatePodDisruptionBudget(namespace string, pdb *policyv1.PodDisruptionBudget) error {
	logger := pdbLogger(namespace, pdb.Name)
	_, err := generateK8sClient().PolicyV1().PodDisruptionBudgets(namespace).Update(context.TODO(), pdb, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "Redis PodDisruptionBudget update failed")
		return err
	}
	logger.Info("Redis PodDisruptionBudget update was successful", "PDB.Spec", pdb.Spec)
	return nil
}

// deletePodDisruptionBudget is a method to delete PodDisruptionBudgets in Kubernetes
func deletePodDisruptionBudget(namespace string, pdbName string) error {
	logger := pdbLogger(namespace, pdbName)
	err := generateK8sClient().PolicyV1().PodDisruptionBudgets(namespace).Delete(context.TODO(), pdbName, metav1.DeleteOptions{})
	if err != nil {
		logger.Error(err, "Redis PodDisruption deletion failed")
		return err
	}
	logger.Info("Redis PodDisruption delete was successful")
	return nil
}

// GetPodDisruptionBudget is a method to get PodDisruptionBudgets in Kubernetes
func GetPodDisruptionBudget(namespace string, pdb string) (*policyv1.PodDisruptionBudget, error) {
	logger := pdbLogger(namespace, pdb)
	getOpts := metav1.GetOptions{
		TypeMeta: generateMetaInformation("PodDisruptionBudget", "policy/v1"),
	}
	pdbInfo, err := generateK8sClient().PolicyV1().PodDisruptionBudgets(namespace).Get(context.TODO(), pdb, getOpts)
	if err != nil {
		logger.Info("Redis PodDisruptionBudget get action failed")
		return nil, err
	}
	logger.Info("Redis PodDisruptionBudget get action was successful")
	return pdbInfo, err
}

// pdbLogger will generate logging interface for PodDisruptionBudgets
func pdbLogger(namespace string, name string) logr.Logger {
	reqLogger := log.WithValues("Request.PodDisruptionBudget.Namespace", namespace, "Request.PodDisruptionBudget.Name", name)
	return reqLogger
}
