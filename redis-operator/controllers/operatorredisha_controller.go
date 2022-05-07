/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "redis-operator/api/v1alpha1"
	"redis-operator/k8sutils"
)

// OperatorRedisHAReconciler reconciles a OperatorRedisHA object
type OperatorRedisHAReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger logr.Logger
}

//+kubebuilder:rbac:groups=apps.operator-redis-ha.org,resources=operatorredishas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.operator-redis-ha.org,resources=operatorredishas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.operator-redis-ha.org,resources=operatorredishas/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *OperatorRedisHAReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.ReconcileHandleInstance(ctx, req)
}

func (r *OperatorRedisHAReconciler) ReconcileHandleInstance(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger = log.FromContext(ctx) // 每次都使用当前上下文

	cr := &appsv1alpha1.OperatorRedisHA{}
	{
		err := r.Client.Get(context.TODO(), req.NamespacedName, cr)
		if err != nil {
			r.Logger.Info(fmt.Sprintf(".Client.Get fail: %v", err))
			if errors.IsNotFound(err) {
				return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_NORMAL}, nil
			}
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_NORMAL}, err
		}
	}

	r.Logger.Info(fmt.Sprintf("ReconcileHandleStatus::Reconcile CRStatus: %d\n", cr.Status.CRStatus))

	switch cr.Status.CRStatus {
	case STATUS_ON_START:
		return r.ReconcileCheckSts(ctx, req, cr)
	case STATUS_IN_CHECK_DELETE:
		return r.ReconcileCheckDelete(ctx, req, cr)
	case STATUS_IN_PROCESS_DELETE:
		return r.ReconcileProcessDelete(ctx, req, cr)
	case STATUS_IN_CHECK_STS:
		return r.ReconcileCheckSts(ctx, req, cr)
	case STATUS_IN_CREATE_STS:
		return r.ReconcileInCreatekSts(ctx, req, cr)
	case STATUS_IN_CHECK_PODS:
		return r.ReconcileCheckPods(ctx, req, cr)
	case STATUS_IN_CHECK_SERVICE:
		return r.ReconcileCheckService(ctx, req, cr)
	case STATUS_IN_CREATE_SERVICE:
		return r.ReconcileInCreateService(ctx, req, cr)
	case STATUS_IN_SET_ETCD_CRT:
		return r.ReconcileSetEtcdCrt(ctx, req, cr)
	case STATUS_IN_CHECK_REDIS_HA:
		return r.ReconcileCheckRedisHA(ctx, req, cr)
	case STATUS_IN_CREATE_REDIS_HA:
		return r.ReconcileCreateRedisHA(ctx, req, cr)
	case STATUS_IN_FIX_REDIS_SERVER:
		return r.ReconcileCheckRedisHA(ctx, req, cr)
	case STATUS_IN_FORCE_REDO_POD:
		return r.ReconcileCheckRedisHA(ctx, req, cr)
	case STATUS_IN_CHECK_NORMAL:
		return r.ReconcileCheckNormal(ctx, req, cr)
	default:
		r.Logger.Info(fmt.Sprintf("ReconcileHandleStatus fail, unknow status: %d\n", cr.Status.CRStatus))
		return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_NORMAL}, nil
	}
}

func (r *OperatorRedisHAReconciler) ReconcileCheckSts(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {

	isStsCreated, err := k8sutils.CheckStatefulSetExist(cr.Namespace, cr.ObjectMeta.Name)
	if nil != err {
		r.Logger.Error(err, fmt.Sprintf("ReconcileInCreatekSts CheckStatefulSetExist fail, err: %v", err))
		return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
	}

	// sts只允许创建,不允许运行时修改
	if !isStsCreated {
		cr.Status.CRStatus = STATUS_IN_CREATE_STS
	} else {
		cr.Status.CRStatus = STATUS_IN_CHECK_PODS
	}

	{
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileCheckSts Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileInCreatekSts(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {
	{
		err := k8sutils.CreateStatefulSet(cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileInCreatekSts CreateStatefulSet fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err // 出错后继续创建
		}
	}

	{
		cr.Status.CRStatus = STATUS_IN_CHECK_PODS
		err := r.Status().Update(ctx, cr) // TODO: 如果update失败, 那下次要回滚?
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileInCreatekSts Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileCheckPods(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {
	{
		checkRet, err := k8sutils.CheckStatefulSetPods(cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileCheckPods CheckStatefulSetPods fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err // 出错后继续检查
		}

		if !checkRet {
			r.Logger.Info(fmt.Sprintf("WARN: ReconcileCheckPods CheckStatefulSetPods fail"))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil // 出错后继检查
		}
	}

	{
		cr.Status.CRStatus = STATUS_IN_CHECK_SERVICE
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileCheckPods Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileCheckService(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {
	isServiceExist, err := k8sutils.CheckServiceExist(cr)
	if nil != err {
		r.Logger.Error(err, fmt.Sprintf("ReconcileCheckService fail, CheckServiceExist err: %v", err))
		return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
	}

	if !isServiceExist {
		cr.Status.CRStatus = STATUS_IN_CREATE_SERVICE
	} else {
		cr.Status.CRStatus = STATUS_IN_SET_ETCD_CRT
	}

	{
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Info(fmt.Sprintf("ReconcileCheckService Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileInCreateService(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {
	{
		err := k8sutils.CreatekService(cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileInCreateService fail, CreatekService err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, err
		}
	}

	{
		cr.Status.RedisAddr = cr.Name + "." + cr.Namespace + ".svc.cluster.local" // 创建完service后即开发连接
		cr.Status.RedisPort = 6379

		cr.Status.CRStatus = STATUS_IN_SET_ETCD_CRT
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileInCreateService Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileCheckRedisHA(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {

	if !cr.Status.BeSlaveOf {
		cr.Status.CRStatus = STATUS_IN_CREATE_REDIS_HA
	} else {
		cr.Status.CRStatus = STATUS_IN_CHECK_NORMAL
	}

	{
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileInCreateService Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileSetEtcdCrt(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {

	if !cr.Status.BeSetEtcdCrt {
		err := k8sutils.ExecuteRedisSetEtcdCrd(cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileSetEtcdCrt fail, ExecuteRedisSetEtcdCrd err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
		}

		cr.Status.BeSetEtcdCrt = true
		cr.Status.CRStatus = STATUS_IN_CHECK_REDIS_HA
	} else {
		cr.Status.CRStatus = STATUS_IN_CHECK_REDIS_HA
	}

	{
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileSetEtcdCrt Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileCreateRedisHA(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {

	err := k8sutils.ExecuteRedisCreateHA(cr)
	if nil != err {
		r.Logger.Error(err, fmt.Sprintf("ReconcileCreateRedisHA fail, ExecuteRedisCreateHA err: %v", err))
		return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
	}

	{
		cr.Status.BeSlaveOf = true
		cr.Status.CRStatus = STATUS_IN_CHECK_NORMAL
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Error(err, fmt.Sprintf("ReconcileCreateRedisHA Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileCheckDelete(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {
	isBeDeleted := k8sutils.CheckRedisFinalizer(cr)
	if isBeDeleted {
		cr.Status.CRStatus = STATUS_IN_PROCESS_DELETE // 如果设置了删除则开始去执行删除
	} else {
		cr.Status.CRStatus = STATUS_IN_CHECK_NORMAL
	}

	err := r.Status().Update(ctx, cr)
	if nil != err {
		r.Logger.Info(fmt.Sprintf("ReconcileCheckDelete Update fail, err: %v", err))
		return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil

	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileProcessDelete(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {
	err := k8sutils.ProcessRedisFinalizer(cr, r.Client)
	if nil != err {
		r.Logger.Info(fmt.Sprintf("ReconcileProcessDelete fail, err: %v", err))
		return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil
	}

	{
		cr.Status.CRStatus = STATUS_IN_CHECK_NORMAL
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Info(fmt.Sprintf("ReconcileProcessDelete Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_EVENT}, nil

		}
	}

	return ctrl.Result{}, nil
}

func (r *OperatorRedisHAReconciler) ReconcileCheckNormal(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.OperatorRedisHA) (ctrl.Result, error) {

	if k8sutils.CheckRedisFinalizer(cr) {
		cr.Status.CRStatus = STATUS_IN_PROCESS_DELETE
		err := r.Status().Update(ctx, cr)
		if nil != err {
			r.Logger.Info(fmt.Sprintf("ReconcileCheckNormal Update fail, err: %v", err))
			return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_NORMAL}, nil
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * TIME_INTERVAL_NORMAL}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OperatorRedisHAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.OperatorRedisHA{}).
		Complete(r)
}
