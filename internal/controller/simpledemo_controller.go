/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sAppv1 "demo.com/pkg/simple/api/v1"
)

var WaitRequeue = 10 * time.Second

// SimpleDemoReconciler reconciles a SimpleDemo object
type SimpleDemoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.demo.com,resources=simpledemoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.demo.com,resources=simpledemoes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.demo.com,resources=simpledemoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimpleDemo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *SimpleDemoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "pp-deployment", req.NamespacedName)
	logger.Info("reconciling  start 开始 Reconcile")

	// 1. 获取资源对象
	md := new(sAppv1.SimpleDemo)
	err := r.Client.Get(ctx, req.NamespacedName, md)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// 防止污染缓存
	mdCopy := md.DeepCopy()

	// ======== 处理 deployment =======
	// 2. 获取 deployment 资源对象
	deploy := new(appsv1.Deployment)
	err = r.Client.Get(ctx, req.NamespacedName, deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			// 2.1 不存在对象
			// 2.1.1 创建 deployment
			errCreate := r.createDeployment(ctx, mdCopy)
			if errCreate != nil {
				return ctrl.Result{}, errCreate
			}
			// 创建 肯定没有成功
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeDeployment,
				fmt.Sprintf(sAppv1.ConditionMessageDeploymentNOTFmt, req.NamespacedName),
				sAppv1.ConditionStatusFalse,
				sAppv1.ConditionReasonDeploymentNOTReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else {
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeDeployment,
				fmt.Sprintf("Deployment %s, err %s", req.NamespacedName, err.Error()),
				sAppv1.ConditionStatusFalse,
				sAppv1.ConditionReasonDeploymentNOTReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}
	} else {
		// 2.2 存在对象
		// 2.2.1 更新 deployment
		err = r.updateDeployment(ctx, mdCopy, req)
		if err != nil {
			return ctrl.Result{}, err
		}

		if deploy.Status.ReadyReplicas == deploy.Status.Replicas {
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeDeployment,
				fmt.Sprintf(sAppv1.ConditionMessageDeploymentOKFmt, req.NamespacedName),
				sAppv1.ConditionStatusTrue,
				sAppv1.ConditionReasonDeploymentReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else {
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeDeployment,
				fmt.Sprintf(sAppv1.ConditionMessageDeploymentNOTFmt, req.NamespacedName),
				sAppv1.ConditionStatusFalse,
				sAppv1.ConditionReasonDeploymentNOTReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		}

	}

	// ======== 处理 service
	// 3. 获取 service 资源对象
	svc := new(corev1.Service)
	err = r.Client.Get(ctx, req.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			// 3.1 不存在

			if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeIngress {
				// 3.1.1 mode 为 ingress
				// 3.1.1.1 创建 clusterIP
				err = r.createService(ctx, mdCopy)
				if err != nil {
					return ctrl.Result{}, err
				}

			} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeNodePort {
				// 3.1.2 mode 为 nodePort
				// 3.1.2.1 创建 nodePort 模式 service
				err = r.createNodePort(ctx, mdCopy)
				if err != nil {
					return ctrl.Result{}, err
				}
			} else {
				// 返回不支持的 MODE
				return ctrl.Result{}, sAppv1.ErrNotSupportMode
			}
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeService,
				fmt.Sprintf(sAppv1.ConditionMessageServiceOKFmt, req.NamespacedName),
				sAppv1.ConditionStatusTrue,
				sAppv1.ConditionReasonServiceReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}

		} else {
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeService,
				fmt.Sprintf("Service %s, err %s", req.NamespacedName, err.Error()),
				sAppv1.ConditionStatusFalse,
				sAppv1.ConditionReasonServiceNOTReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}
	} else {
		// 3.2 存在
		if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeIngress {
			// 3.2.1 mode 为 ingress
			// 3.2.1.1 更新 clusterIP
			err = r.updateService(ctx, mdCopy, req)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeNodePort {
			// 3.2.2 mode 为 nodePort
			// 3.2.2.1 更新 nodePort 模式 service
			err = r.updateNodePort(ctx, mdCopy, req)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, sAppv1.ErrNotSupportMode
		}
		_, errStatus := r.updateStatus(ctx,
			mdCopy,
			sAppv1.ConditionTypeService,
			fmt.Sprintf(sAppv1.ConditionMessageServiceOKFmt, req.NamespacedName),
			sAppv1.ConditionStatusTrue,
			sAppv1.ConditionReasonServiceReady)
		if errStatus != nil {
			return ctrl.Result{}, errStatus
		}
	}

	// ========= 处理 ingress
	// 4. 获取 ingress 资源
	ig := new(netv1.Ingress)
	err = r.Client.Get(ctx, req.NamespacedName, ig)
	if err != nil {
		if errors.IsNotFound(err) {
			// 4.1 不存在
			if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeIngress {
				// 4.1.1 mode 为 ingress
				// 4.1.1.1 创建 ingress
				err = r.createIngress(ctx, mdCopy)
				if err != nil {
					return ctrl.Result{}, err
				}
				_, errStatus := r.updateStatus(ctx,
					mdCopy,
					sAppv1.ConditionTypeIngress,
					fmt.Sprintf(sAppv1.ConditionMessageIngressNOTFmt, req.NamespacedName),
					sAppv1.ConditionStatusFalse,
					sAppv1.ConditionReasonIngressNOTReady)
				if errStatus != nil {
					return ctrl.Result{}, errStatus
				}
			} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeNodePort {
				// 4.1.2 mode 为 nodePort
				// 4.1.2.1 退出 因为此时已经创建好了
				/*
					可能存在 ingress 不存在但是 有 condition 情况
				*/
				r.deleteStatus(mdCopy, sAppv1.ConditionTypeIngress)
				/*
					这里的修改无法执行到 apiserver 中去 所以
					当已经存在子资源的时候 无法删除
				*/
				return ctrl.Result{}, nil
			}
		} else {
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeIngress,
				fmt.Sprintf("Ingress %s, err %s", req.NamespacedName, err.Error()),
				sAppv1.ConditionStatusFalse,
				sAppv1.ConditionReasonIngressNOTReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
			return ctrl.Result{}, err
		}
	} else {
		// 4.2 存在
		if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeIngress {
			// 4.2.1 mode 为 ingress
			// 4.2.1.1 更新 ingress
			err = r.updateIngress(ctx, mdCopy, req)
			if err != nil {
				return ctrl.Result{}, err
			}
			_, errStatus := r.updateStatus(ctx,
				mdCopy,
				sAppv1.ConditionTypeIngress,
				fmt.Sprintf(sAppv1.ConditionMessageIngressOKFmt, req.NamespacedName),
				sAppv1.ConditionStatusTrue,
				sAppv1.ConditionReasonIngressReady)
			if errStatus != nil {
				return ctrl.Result{}, errStatus
			}
		} else if strings.ToLower(mdCopy.Spec.Expose.Mode) == sAppv1.ModeNodePort {
			// 4.2.2 mode 为 nodePort
			// 4.2.2.1 删除 ingress 因为是 nodePort 不需要 ingress
			err = r.deleteIngress(ctx, mdCopy)
			if err != nil {
				return ctrl.Result{}, err
			}
			r.deleteStatus(mdCopy, sAppv1.ConditionTypeIngress)
		}
	}

	sus, errStatus := r.updateStatus(ctx,
		mdCopy,
		"",
		"",
		"",
		"")
	if errStatus != nil {
		logger.Info("!!!!=====nil")
		return ctrl.Result{}, errStatus
	} else if !sus {
		// 如果系统状态还没有成功 则等待 10s
		logger.Info("reconciling  end 结束 end end end")
		return ctrl.Result{RequeueAfter: WaitRequeue}, nil
	}

	logger.Info("reconciling  end 结束 end end end")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleDemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sAppv1.SimpleDemo{}).
		Owns(&appsv1.Deployment{}).
		//WithEventFilter(predicate.Funcs{
		//	CreateFunc: func(e event.CreateEvent) bool {
		//		return e.Object.GetName() == "specific-deployment-name"
		//	},
		//	UpdateFunc: func(e event.UpdateEvent) bool {
		//		return e.ObjectNew.GetName() == "specific-deployment-name"
		//	},
		//	DeleteFunc: func(e event.DeleteEvent) bool {
		//		return e.Object.GetName() == "specific-deployment-name"
		//	},
		//	GenericFunc: func(e event.GenericEvent) bool {
		//		return e.Object.GetName() == "specific-deployment-name"
		//	},
		//}).
		Owns(&netv1.Ingress{}).
		Owns(&corev1.Service{}).
		Named("simpledemo").
		Complete(r)
}

func (r *SimpleDemoReconciler) createDeployment(ctx context.Context, md *sAppv1.SimpleDemo) error {
	deploy, err := NewDeployment(md)
	if err != nil {
		return err
	}
	err = controllerutil.SetControllerReference(md, deploy, r.Scheme)
	if err != nil {
		return err
	}
	return r.Client.Create(ctx, deploy)
}

func (r *SimpleDemoReconciler) updateDeployment(ctx context.Context, md *sAppv1.SimpleDemo, req ctrl.Request) error {
	// 1. 得到应该创建的资源对象
	desiredDeploy, err := NewDeployment(md)
	if err != nil {
		return err
	}
	// 2. 获取现有的 Deployment
	currentDeploy := &appsv1.Deployment{}
	err = r.Client.Get(ctx, req.NamespacedName, currentDeploy)
	if err != nil {
		return err
	}

	err = r.Update(ctx, desiredDeploy, client.DryRunAll)
	if err != nil {
		return err
	}

	// 3. 比较现有和新的 Deployment 对象，确保需要更新
	needUpdate := !reflect.DeepEqual(currentDeploy.Spec, desiredDeploy.Spec)

	if !needUpdate {
		return nil
	}
	// 4. 更新 `currentDeploy` 的 Spec
	currentDeploy.Spec = desiredDeploy.Spec
	return r.Client.Update(ctx, currentDeploy)
}

func (r *SimpleDemoReconciler) createService(ctx context.Context, md *sAppv1.SimpleDemo) error {
	service, err := NewService(md)
	if err != nil {
		return err
	}
	err = controllerutil.SetControllerReference(md, service, r.Scheme)
	if err != nil {
		return err
	}
	return r.Client.Create(ctx, service)
}

func (r *SimpleDemoReconciler) createNodePort(ctx context.Context, md *sAppv1.SimpleDemo) error {
	np, err := NewNodePort(md)
	if err != nil {
		return err
	}
	err = controllerutil.SetControllerReference(md, np, r.Scheme)
	if err != nil {
		return err
	}
	return r.Client.Create(ctx, np)
}

func (r *SimpleDemoReconciler) updateService(ctx context.Context, md *sAppv1.SimpleDemo, req ctrl.Request) error {
	desiredService, err := NewService(md)
	if err != nil {
		return err
	}
	// 2. 获取现有的 Deployment
	currentService := &corev1.Service{}
	err = r.Client.Get(ctx, req.NamespacedName, currentService)
	if err != nil {
		return err
	}
	err = r.Update(ctx, desiredService, client.DryRunAll)
	if err != nil {
		return err
	}
	// 3. 比较现有和新的 Deployment 对象，确保需要更新
	needUpdate := !reflect.DeepEqual(currentService.Spec, desiredService.Spec)

	if !needUpdate {
		return nil
	}
	// 4. 更新 `currentDeploy` 的 Spec
	currentService.Spec = desiredService.Spec

	return r.Client.Update(ctx, currentService)
}

func (r *SimpleDemoReconciler) updateNodePort(ctx context.Context, md *sAppv1.SimpleDemo, req ctrl.Request) error {
	desiredService, err := NewNodePort(md)
	if err != nil {
		return err
	}
	// 2. 获取现有的 Deployment
	currentService := &corev1.Service{}
	err = r.Client.Get(ctx, req.NamespacedName, currentService)
	if err != nil {
		return err
	}
	err = r.Update(ctx, desiredService, client.DryRunAll)
	if err != nil {
		return err
	}
	// 3. 比较现有和新的 Deployment 对象，确保需要更新
	needUpdate := !reflect.DeepEqual(currentService.Spec, desiredService.Spec)

	if !needUpdate {
		return nil
	}
	// 4. 更新 `currentDeploy` 的 Spec
	currentService.Spec = desiredService.Spec

	return r.Client.Update(ctx, currentService)

}

func (r *SimpleDemoReconciler) createIngress(ctx context.Context, md *sAppv1.SimpleDemo) error {
	ig, err := NewIngress(md)
	if err != nil {
		return err
	}
	err = controllerutil.SetControllerReference(md, ig, r.Scheme)
	if err != nil {
		return err
	}
	return r.Client.Create(ctx, ig)
}

func (r *SimpleDemoReconciler) updateIngress(ctx context.Context, md *sAppv1.SimpleDemo, req ctrl.Request) error {
	desiredIngress, err := NewIngress(md)
	if err != nil {
		return err
	}
	// 2. 获取现有的 Deployment
	currentIngress := &netv1.Ingress{}
	err = r.Client.Get(ctx, req.NamespacedName, currentIngress)
	if err != nil {
		return err
	}
	err = r.Update(ctx, desiredIngress, client.DryRunAll)
	if err != nil {
		return err
	}
	// 3. 比较现有和新的 Deployment 对象，确保需要更新
	needUpdate := !reflect.DeepEqual(currentIngress.Spec, desiredIngress.Spec)

	if !needUpdate {
		return nil
	}
	// 4. 更新 `currentDeploy` 的 Spec
	currentIngress.Spec = desiredIngress.Spec

	return r.Client.Update(ctx, currentIngress)
}

func (r *SimpleDemoReconciler) deleteIngress(ctx context.Context, md *sAppv1.SimpleDemo) error {
	ig, err := NewIngress(md)
	if err != nil {
		return err
	}
	err = controllerutil.SetControllerReference(md, ig, r.Scheme)
	if err != nil {
		return err
	}
	return r.Client.Delete(ctx, ig)
}

func (r *SimpleDemoReconciler) updateStatus(
	ctx context.Context, md *sAppv1.SimpleDemo, conditionType, message, status, reason string) (bool, error) {

	/*
		整体的逻辑是 根据不同种类的 condition 都加入到 conditions 中
		然后检查是否有 false 的 condition
		如果有 false 那么更新 status 的总体情况 且返回 success
		如果有 true  那么更新 status 为好的 返回 faile
	*/
	if conditionType != "" {
		var condition *sAppv1.Condition
		for i := range md.Status.Conditions {
			if md.Status.Conditions[i].Type == conditionType {
				condition = &md.Status.Conditions[i]
			}
		}
		if condition != nil {
			if condition.Status != status ||
				condition.Message != message ||
				condition.Reason != reason {
				// update
				condition.Status = status
				condition.Message = message
				condition.Reason = reason
			}
		} else {
			md.Status.Conditions = append(md.Status.Conditions, createCondition(conditionType, message, status, reason))
		}
	}

	m, re, p, sus := isSuccess(md.Status.Conditions)
	if sus {
		md.Status.Message = sAppv1.StatusMessageSuccess
		md.Status.Reason = sAppv1.StatusReasonSuccess
		md.Status.Phase = sAppv1.StatusPhaserComplete
	} else {
		md.Status.Message = m
		md.Status.Reason = re
		md.Status.Phase = p
	}
	// 执行更新
	return sus, r.Client.Status().Update(ctx, md)
}

func isSuccess(conditions []sAppv1.Condition) (message, reason, parse string, sus bool) {
	if len(conditions) == 0 {
		return "", "", "", false
	}
	for i := range conditions {
		if conditions[i].Status == sAppv1.ConditionStatusFalse {
			return conditions[i].Message, conditions[i].Reason, conditions[i].Status, false
		}
	}

	return "", "", "", true
}

func createCondition(conditionType, message, status, reason string) sAppv1.Condition {
	return sAppv1.Condition{
		Type:               conditionType,
		Status:             status,
		Message:            message,
		Reason:             reason,
		LastTransitionTime: metav1.Now(),
	}
}

// 可以多次删除 幂等性 不需要错误返回
func (r *SimpleDemoReconciler) deleteStatus(md *sAppv1.SimpleDemo, conditionType string) {

	for i := range md.Status.Conditions {
		if md.Status.Conditions[i].Type == conditionType {
			// 执行删除
			md.Status.Conditions = deleteCondition(md.Status.Conditions, i)
		}
	}
}

func deleteCondition(conditions []sAppv1.Condition, i int) []sAppv1.Condition {
	// 切片中的元素顺序不敏感

	// 1 要删除元素的索引值不能大于切片长度
	if i >= len(conditions) {
		return []sAppv1.Condition{}
	}
	// 2 如果切片长度为 1 且索引值为 0 直接清空
	if len(conditions) == 1 && i == 0 {
		return conditions[:0]
	}
	// 3 如果长度-1为索引值 删除最后一个元素
	if len(conditions)-1 == i {
		conditions = conditions[:len(conditions)-1]
		return conditions
	}
	// 4 默认处理方式 交换索引元素和最后一个元素 删除最后一个元素
	conditions[i], conditions[len(conditions)-1] = conditions[len(conditions)-1], conditions[i]
	conditions = conditions[:len(conditions)-1]
	return conditions
}
