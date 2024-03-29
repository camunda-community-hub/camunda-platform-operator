/*
Copyright 2021.

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
	"strings"

	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	camundacloudv1 "io.camnda/operator/api/v1"
)

// ZeebeReconciler reconciles a Zeebe object
type ZeebeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const statefulset_name = "zeebe"

//+kubebuilder:rbac:groups=camunda-cloud.io.camunda,resources=zeebes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=camunda-cloud.io.camunda,resources=zeebes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=camunda-cloud.io.camunda,resources=zeebes/finalizers,verbs=update

// CRUD apps: deployments and statefulsets
// +kubebuilder:rbac:groups=apps,resources=statefulsets;deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status;deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Zeebe object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ZeebeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var zeebe camundacloudv1.Zeebe
	if err := r.Get(ctx, req.NamespacedName, &zeebe); err != nil {
		logger.Error(err, "unable to fetch Zeebe resource")

		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "Operator",
		"app.kubernetes.io/name":       "zeebe-cluster",
		"app.kubernetes.io/app":        statefulset_name,
		"app.kubernetes.io/component":  "broker",
		"app":                          statefulset_name,
	}

	brokerConfigMap := &v12.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zeebe-configmap",
			Labels:    labels,
			Namespace: req.Namespace,
		},
		Data: map[string]string{
			"startup.sh": "" +
				"#!/usr/bin/env bash\n" +
				"set -eux -o pipefail\n" +
				"export ZEEBE_BROKER_CLUSTER_NODEID=$(echo $K8S_NAME | tr -d \"[:alpha:]-\")\n" +
				"exec /usr/local/zeebe/bin/broker",
		},
	}

	if err := ctrl.SetControllerReference(&zeebe, brokerConfigMap, r.Scheme); err != nil {
		logger.Error(err, "unable to construct config map from zeebe CRD")
		// don't bother requeuing until we get a change to the spec
		return ctrl.Result{}, nil
	}

	if err := r.Create(ctx, brokerConfigMap); err != nil {
		logger.Error(err, "unable to create config map for Zeebe", "configmap", brokerConfigMap)
		return ctrl.Result{}, err
	}

	logger.V(1).Info("created configmap for Zeebe", "configmap", brokerConfigMap)

	brokerService := r.createBrokerService(labels, req.Namespace)

	if err := ctrl.SetControllerReference(&zeebe, brokerService, r.Scheme); err != nil {
		logger.Error(err, "unable to construct service from zeebe CRD")
		// don't bother requeuing until we get a change to the spec
		return ctrl.Result{}, nil
	}

	if err := r.Create(ctx, brokerService); err != nil {
		logger.Error(err, "unable to create service for Zeebe", "service", brokerService)
		return ctrl.Result{}, err
	}

	logger.V(1).Info("created service for Zeebe", "service", brokerService)

	brokerStatefulSet := r.createBrokerStatefulset(zeebe, labels, req)

	if err := ctrl.SetControllerReference(&zeebe, brokerStatefulSet, r.Scheme); err != nil {
		logger.Error(err, "unable to construct statefulset from zeebe CRD")
		// don't bother requeuing until we get a change to the spec
		return ctrl.Result{}, nil
	}

	if err := r.Create(ctx, brokerStatefulSet); err != nil {
		logger.Error(err, "unable to create statefulset for Zeebe", "statefulset", brokerStatefulSet)
		return ctrl.Result{}, err
	}

	logger.V(1).Info("created statefulset for Zeebe", "statefulset", brokerStatefulSet)

	// We return an empty result and no error,
	// which indicates to controller-runtime that we’ve successfully reconciled
	// this object and don’t need to try again until there’s some changes.
	return ctrl.Result{}, nil
}

func (r *ZeebeReconciler) createBrokerStatefulset(zeebe camundacloudv1.Zeebe, labels map[string]string, req ctrl.Request) *v1.StatefulSet {
	storageClassName := "ssd"
	backendSpec := zeebe.Spec.Broker.Backend
	brokerStatefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      statefulset_name,
			Namespace: req.Namespace,
		},
		Spec: v1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: backendSpec.Replicas,
			Template: createPodSpecTemplate(labels, zeebe.Spec, req.Namespace),
			VolumeClaimTemplates: []v12.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: v12.PersistentVolumeClaimSpec{
						AccessModes: []v12.PersistentVolumeAccessMode{
							v12.ReadWriteOnce,
						},
						StorageClassName: &storageClassName,
						Resources: v12.ResourceRequirements{
							Requests: v12.ResourceList{
								"storage": *resource.NewQuantity(128*1024*1024, resource.DecimalExponent),
							},
						},
					},
				},
			},
		},
	}
	return brokerStatefulSet
}

func (r *ZeebeReconciler) createBrokerService(labels map[string]string, namespace string) *v12.Service {
	brokerService := &v12.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      statefulset_name,
			Namespace: namespace,
		},
		Spec: v12.ServiceSpec{
			ClusterIP:                v12.ClusterIPNone,
			PublishNotReadyAddresses: true,
			Type:                     v12.ServiceTypeClusterIP,
			Ports: []v12.ServicePort{
				{
					Port:     9600,
					Protocol: v12.ProtocolTCP,
					Name:     "http",
				},
				{
					Port:     26502,
					Protocol: v12.ProtocolTCP,
					Name:     "internal",
				},
				{
					Port:     26501,
					Protocol: v12.ProtocolTCP,
					Name:     "command",
				},
			},
			Selector: labels,
		},
	}
	return brokerService
}

func createPodSpecTemplate(labels map[string]string, zeebeSpec camundacloudv1.ZeebeSpec, namespace string) v12.PodTemplateSpec {
	podAddressFormat := "%s-%d.%s.%s.svc.cluster.local:26502"

	replicas := *zeebeSpec.Broker.Backend.Replicas
	podAddresses := make([]string, replicas)
	var podIndex int32
	for podIndex = 0; podIndex < replicas; podIndex++ {
		podAddresses[podIndex] = fmt.Sprintf(podAddressFormat, statefulset_name, podIndex, statefulset_name, namespace)
	}

	backendSpec := zeebeSpec.Broker.Backend
	envs := []v12.EnvVar{
		{
			Name:  "ZEEBE_BROKER_GATEWAY_ENABLE",
			Value: fmt.Sprintf("%t", !zeebeSpec.Gateway.Standalone),
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_PARTITIONSCOUNT",
			Value: fmt.Sprintf("%d", *zeebeSpec.Broker.Partitions.Count),
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_REPLICATIONFACTOR",
			Value: fmt.Sprintf("%d", *zeebeSpec.Broker.Partitions.Replication),
		},
		{ // TODO SUPPORT BIGGER CLUSTER
			Name:  "ZEEBE_BROKER_CLUSTER_NODEID",
			Value: fmt.Sprintf("%d", 0),
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_CLUSTERSIZE",
			Value: fmt.Sprintf("%d", *backendSpec.Replicas),
		},
		{
			Name: "K8S_NAME",
			ValueFrom: &v12.EnvVarSource{
				FieldRef: &v12.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name:  "K8S_SERVICE_NAME",
			Value: statefulset_name,
		},
		{
			Name: "K8S_NAMESPACE",
			ValueFrom: &v12.EnvVarSource{
				FieldRef: &v12.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name:  "ZEEBE_BROKER_NETWORK_ADVERTISEDHOST",
			Value: "$(K8S_NAME).$(K8S_SERVICE_NAME).$(K8S_NAMESPACE).svc.cluster.local",
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_CLUSTERNAME",
			Value: "$(K8S_NAMESPACE)",
		},
		{
			Name:  "ZEEBE_BROKER_GATEWAY_CLUSTER_HOST",
			Value: "$(ZEEBE_BROKER_NETWORK_ADVERTISEDHOST)",
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_INITIALCONTACTPOINTS",
			Value: strings.Join(podAddresses, ","),
		},
		{
			Name:  "ZEEBE_LOG_STACKDRIVER_SERVICENAME",
			Value: "zeebe",
		},
		{
			Name:  "ZEEBE_LOG_STACKDRIVER_SERVICEVERSION",
			Value: backendSpec.ImageTag,
		},
	}

	for _, env := range backendSpec.OverrideEnv {
		envs = append(envs, env)
	}

	return v12.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v12.PodSpec{
			Containers: []v12.Container{
				{
					Name:            statefulset_name,
					Image:           fmt.Sprintf("%s:%s", backendSpec.ImageName, backendSpec.ImageTag),
					ImagePullPolicy: v12.PullAlways,
					Env:             envs,
					Ports: []v12.ContainerPort{
						{
							ContainerPort: 9600,
							Name:          "http",
						},
						{
							ContainerPort: 26501,
							Name:          "command",
						},
						{
							ContainerPort: 26502,
							Name:          "internal",
						},
					},
					ReadinessProbe: &v12.Probe{
						Handler: v12.Handler{
							HTTPGet: &v12.HTTPGetAction{
								Path: "/ready",
								Port: intstr.IntOrString{
									IntVal: 9600,
								},
							},
						},
						PeriodSeconds:    10,
						SuccessThreshold: 1,
						TimeoutSeconds:   1,
					},
					Resources: backendSpec.Resources,
					VolumeMounts: []v12.VolumeMount{
						// TODO add config map
						//{
						//	Name:      "config",
						//	MountPath: " /usr/local/zeebe/config/application.yaml",
						//	SubPath:   "application.yaml",
						//},
						{
							Name:      "config",
							MountPath: "/usr/local/bin/startup.sh",
							SubPath:   "startup.sh",
						},
						{
							Name:      "data",
							MountPath: "/usr/local/zeebe/data",
						},
					},
				},
			},
			Volumes: []v12.Volume{
				{
					Name: "config",
					VolumeSource: v12.VolumeSource{
						ConfigMap: &v12.ConfigMapVolumeSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: "zeebe-configmap",
							},
							DefaultMode: getIntPointer(0744),
						},
					},
				},
			},
		},
	}
}

func getIntPointer(val int32) *int32 {
	return &val
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZeebeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&camundacloudv1.Zeebe{}).
		Complete(r)
}
