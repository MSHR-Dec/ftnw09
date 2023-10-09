/*
Copyright 2023.

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
	"github.com/MSHR-Dec/ftnw09/internal/driver"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mshrv1 "github.com/MSHR-Dec/ftnw09/api/v1"
)

// SSpyFamillllyReconciler reconciles a SSpyFamilllly object
type SSpyFamillllyReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ProbeInterval time.Duration
}

//+kubebuilder:rbac:groups=mshr.whisffee.net,resources=sspyfamillllies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mshr.whisffee.net,resources=sspyfamillllies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mshr.whisffee.net,resources=sspyfamillllies/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SSpyFamilllly object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *SSpyFamillllyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var ssf mshrv1.SSpyFamilllly
	err := r.Get(ctx, req.NamespacedName, &ssf)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "unable to get SSpyFamilllly", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if !ssf.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	dbClient := driver.NewMySQLClient(
		ssf.Spec.DBUser,
		ssf.Spec.DBPassword,
		ssf.Spec.DBHost,
		strconv.Itoa(ssf.Spec.DBPort),
		ssf.Spec.TargetDatabase,
	)
	mv, err := dbClient.FindLatestMigrationVersion()
	if err != nil {
		logger.Error(err, "unable to find latest migration version")
		return ctrl.Result{}, err
	}
	logger.Info(strconv.Itoa(mv))

	err = r.reconcileConfigMap(ctx, ssf, mv)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcileDeployment(ctx, ssf, mv)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.reconcileService(ctx, ssf, mv)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, ssf)
}

type ticker struct {
	events chan event.GenericEvent

	interval time.Duration
}

func (t *ticker) Start(ctx context.Context) error {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			t.events <- event.GenericEvent{}
		}
	}
}

func (t *ticker) NeedLeaderElection() bool {
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *SSpyFamillllyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register the ticker as a runnable managed by the manager.
	// The storage bucket controller's reconciliation will be triggered by the events send by the ticker.
	events := make(chan event.GenericEvent)
	err := mgr.Add(&ticker{
		events:   events,
		interval: r.ProbeInterval,
	})
	if err != nil {
		return err
	}

	// The storage bucket controller watches the events send by the ticker.
	// When it receives an event from ticker, reconcile requests for all storage buckets will be enqueued (= polling).
	source := source.Channel{
		Source:         events,
		DestBufferSize: 0,
	}

	handler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
		sspyfamillllies := mshrv1.SSpyFamillllyList{}
		mgr.GetCache().List(ctx, &sspyfamillllies)

		var requests []reconcile.Request
		for _, ssf := range sspyfamillllies.Items {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      ssf.Name,
					Namespace: ssf.Namespace,
				},
			})
		}

		return requests
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&mshrv1.SSpyFamilllly{}).
		WatchesRawSource(&source, handler).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *SSpyFamillllyReconciler) reconcileConfigMap(ctx context.Context, ssf mshrv1.SSpyFamilllly, migrationVersion int) error {
	logger := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(ssf.Namespace)
	cm.SetName(ssf.Name)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		content := fmt.Sprintf(`schemaspy.t=mysql
schemaspy.dp=/app/mysql-connector-java.jar
schemaspy.host=%s
schemaspy.port=%d
schemaspy.db=%s
schemaspy.s=%s
schemaspy.u=%s
schemaspy.p=%s
schemaspy.o=/schema`,
			ssf.Spec.DBHost,
			ssf.Spec.DBPort,
			ssf.Spec.TargetDatabase,
			ssf.Spec.TargetDatabase,
			ssf.Spec.DBUser,
			ssf.Spec.DBPassword)
		cm.Data["schemaspy.properties"] = content
		return nil
	})

	if err != nil {
		logger.Error(err, "unable to create or update ConfigMap")
		return err
	}
	if op != controllerutil.OperationResultNone {
		logger.Info("reconcile ConfigMap successfully", "op", op)
	}
	return nil
}

func (r *SSpyFamillllyReconciler) reconcileDeployment(ctx context.Context, ssf mshrv1.SSpyFamilllly, migrationVersion int) error {
	logger := log.FromContext(ctx)

	dep := buildDeployment(ssf.Name, ssf.Namespace, migrationVersion)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dep)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: ssf.Namespace, Name: ssf.Name}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, "sspy-familllly-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "sspy-familllly-controller",
		Force:        pointer.Bool(true),
	})

	if err != nil {
		logger.Error(err, "unable to create or update Deployment")
		return err
	}
	logger.Info("reconcile Deployment successfully", "name", ssf.Name)
	return nil
}

func (r *SSpyFamillllyReconciler) reconcileService(ctx context.Context, ssf mshrv1.SSpyFamilllly, migrationVersion int) error {
	logger := log.FromContext(ctx)

	svc := buildService(ssf.Name, ssf.Namespace, migrationVersion)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svc)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current corev1.Service
	err = r.Get(ctx, client.ObjectKey{Namespace: ssf.Namespace, Name: ssf.Name}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := corev1apply.ExtractService(&current, "sspy-familllly-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(svc, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "sspy-familllly-controller",
		Force:        pointer.Bool(true),
	})
	if err != nil {
		logger.Error(err, "unable to create or update Service")
		return err
	}

	logger.Info("reconcile Service successfully", "name", ssf.Name)
	return nil
}

//func (r *SSpyFamillllyReconciler) _reconcileConfigMap(ctx context.Context, ssf mshrv1.SSpyFamilllly) error {
//	logger := log.FromContext(ctx)
//
//	cm := &corev1.ConfigMap{}
//	cm.SetNamespace(ssf.Namespace)
//	cm.SetName("ssf-" + ssf.Name)
//
//	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
//		if cm.Data == nil {
//			cm.Data = make(map[string]string)
//		}
//		for name, content := range ssf.Spec.Markdowns {
//			cm.Data[name] = content
//		}
//		return nil
//	})
//
//	if err != nil {
//		logger.Error(err, "unable to create or update ConfigMap")
//		return err
//	}
//	if op != controllerutil.OperationResultNone {
//		logger.Info("reconcile ConfigMap successfully", "op", op)
//	}
//	return nil
//}
//
//func (r *SSpyFamillllyReconciler) _reconcileDeployment(ctx context.Context, ssf mshrv1.SSpyFamilllly) error {
//	logger := log.FromContext(ctx)
//
//	depName := "ssf-" + ssf.Name
//	viewerImage := "peaceiris/mdbook:latest"
//	if len(ssf.Spec.ViewerImage) != 0 {
//		viewerImage = ssf.Spec.ViewerImage
//	}
//
//	dep := appsv1apply.Deployment(depName, ssf.Namespace).
//		WithLabels(map[string]string{
//			"app.kubernetes.io/name":       "ssfbook",
//			"app.kubernetes.io/instance":   ssf.Name,
//			"app.kubernetes.io/created-by": "sspy-familllly-controller",
//		}).
//		WithSpec(appsv1apply.DeploymentSpec().
//			WithReplicas(ssf.Spec.Replicas).
//			WithSelector(metav1apply.LabelSelector().WithMatchLabels(map[string]string{
//				"app.kubernetes.io/name":       "ssfbook",
//				"app.kubernetes.io/instance":   ssf.Name,
//				"app.kubernetes.io/created-by": "sspy-familllly-controller",
//			})).
//			WithTemplate(corev1apply.PodTemplateSpec().
//				WithLabels(map[string]string{
//					"app.kubernetes.io/name":       "ssfbook",
//					"app.kubernetes.io/instance":   ssf.Name,
//					"app.kubernetes.io/created-by": "sspy-familllly-controller",
//				}).
//				WithSpec(corev1apply.PodSpec().
//					WithContainers(corev1apply.Container().
//						WithName("ssfbook").
//						WithImage(viewerImage).
//						WithImagePullPolicy(corev1.PullIfNotPresent).
//						WithCommand("mdbook").
//						WithArgs("serve", "--hostname", "0.0.0.0").
//						WithVolumeMounts(corev1apply.VolumeMount().
//							WithName("markdowns").
//							WithMountPath("/book/src"),
//						).
//						WithPorts(corev1apply.ContainerPort().
//							WithName("http").
//							WithProtocol(corev1.ProtocolTCP).
//							WithContainerPort(3000),
//						).
//						WithLivenessProbe(corev1apply.Probe().
//							WithHTTPGet(corev1apply.HTTPGetAction().
//								WithPort(intstr.FromString("http")).
//								WithPath("/").
//								WithScheme(corev1.URISchemeHTTP),
//							),
//						).
//						WithReadinessProbe(corev1apply.Probe().
//							WithHTTPGet(corev1apply.HTTPGetAction().
//								WithPort(intstr.FromString("http")).
//								WithPath("/").
//								WithScheme(corev1.URISchemeHTTP),
//							),
//						),
//					).
//					WithVolumes(corev1apply.Volume().
//						WithName("markdowns").
//						WithConfigMap(corev1apply.ConfigMapVolumeSource().
//							WithName("ssf-" + ssf.Name),
//						),
//					),
//				),
//			),
//		)
//
//	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dep)
//	if err != nil {
//		return err
//	}
//	patch := &unstructured.Unstructured{
//		Object: obj,
//	}
//
//	var current appsv1.Deployment
//	err = r.Get(ctx, client.ObjectKey{Namespace: ssf.Namespace, Name: depName}, &current)
//	if err != nil && !errors.IsNotFound(err) {
//		return err
//	}
//
//	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, "sspy-familllly-controller")
//	if err != nil {
//		return err
//	}
//
//	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
//		return nil
//	}
//
//	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
//		FieldManager: "sspy-familllly-controller",
//		Force:        pointer.Bool(true),
//	})
//
//	if err != nil {
//		logger.Error(err, "unable to create or update Deployment")
//		return err
//	}
//	logger.Info("reconcile Deployment successfully", "name", ssf.Name)
//	return nil
//}
//
//func (r *SSpyFamillllyReconciler) _reconcileService(ctx context.Context, ssf mshrv1.SSpyFamilllly) error {
//	logger := log.FromContext(ctx)
//	svcName := "ssf-" + ssf.Name
//
//	svc := corev1apply.Service(svcName, ssf.Namespace).
//		WithLabels(map[string]string{
//			"app.kubernetes.io/name":       "ssfbook",
//			"app.kubernetes.io/instance":   ssf.Name,
//			"app.kubernetes.io/created-by": "sspy-familllly-controller",
//		}).
//		WithSpec(corev1apply.ServiceSpec().
//			WithSelector(map[string]string{
//				"app.kubernetes.io/name":       "ssfbook",
//				"app.kubernetes.io/instance":   ssf.Name,
//				"app.kubernetes.io/created-by": "sspy-familllly-controller",
//			}).
//			WithType(corev1.ServiceTypeClusterIP).
//			WithPorts(corev1apply.ServicePort().
//				WithProtocol(corev1.ProtocolTCP).
//				WithPort(80).
//				WithTargetPort(intstr.FromInt(3000)),
//			),
//		)
//
//	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svc)
//	if err != nil {
//		return err
//	}
//	patch := &unstructured.Unstructured{
//		Object: obj,
//	}
//
//	var current corev1.Service
//	err = r.Get(ctx, client.ObjectKey{Namespace: ssf.Namespace, Name: svcName}, &current)
//	if err != nil && !errors.IsNotFound(err) {
//		return err
//	}
//
//	currApplyConfig, err := corev1apply.ExtractService(&current, "sspy-familllly-controller")
//	if err != nil {
//		return err
//	}
//
//	if equality.Semantic.DeepEqual(svc, currApplyConfig) {
//		return nil
//	}
//
//	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
//		FieldManager: "sspy-familllly-controller",
//		Force:        pointer.Bool(true),
//	})
//	if err != nil {
//		logger.Error(err, "unable to create or update Service")
//		return err
//	}
//
//	logger.Info("reconcile Service successfully", "name", ssf.Name)
//	return nil
//}

func (r *SSpyFamillllyReconciler) updateStatus(ctx context.Context, ssf mshrv1.SSpyFamilllly) (ctrl.Result, error) {
	var dep appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: ssf.Namespace, Name: ssf.Name}, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	var status mshrv1.SSpyFamillllyStatus
	if dep.Status.AvailableReplicas == 0 {
		status = mshrv1.SSpyFamillllyNotReady
	} else {
		status = mshrv1.SSpyFamillllyAvailable
	}

	if ssf.Status != status {
		ssf.Status = status
		err = r.Status().Update(ctx, &ssf)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if ssf.Status != mshrv1.SSpyFamillllyHealthy {
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}
