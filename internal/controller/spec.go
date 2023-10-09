package controller

import (
	corev1 "k8s.io/api/core/v1"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"strconv"
)

func buildDeployment(name string, namespace string, migrationVersion int) *appsv1apply.DeploymentApplyConfiguration {
	instanceLabel := "schemaspy-" + strconv.Itoa(migrationVersion)

	return appsv1apply.Deployment(name, namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "schemaspy",
			"app.kubernetes.io/instance":   instanceLabel,
			"app.kubernetes.io/component":  "ops",
			"app.kubernetes.io/part-of":    "migration",
			"app.kubernetes.io/managed-by": "sspy-familllly-controller",
		}).
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(1).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(map[string]string{
				"app.kubernetes.io/name": "schemaspy",
			})).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(map[string]string{
					"app.kubernetes.io/name":       "schemaspy",
					"app.kubernetes.io/instance":   instanceLabel,
					"app.kubernetes.io/component":  "ops",
					"app.kubernetes.io/part-of":    "migration",
					"app.kubernetes.io/managed-by": "sspy-familllly-controller",
				}).
				WithSpec(corev1apply.PodSpec().
					WithInitContainers(
						corev1apply.Container().
							WithName("schemaspy").
							WithImage("schemaspy:latest").
							WithImagePullPolicy(corev1.PullIfNotPresent).
							WithCommand(
								"java",
								"-jar",
								"schemaspy.jar",
								"-configFile",
								"/schemaspy.properties",
							).
							WithVolumeMounts(
								corev1apply.VolumeMount().
									WithName("properties").
									WithMountPath("/schemaspy.properties").
									WithSubPath("schemaspy.properties"),
								corev1apply.VolumeMount().
									WithName("output").
									WithMountPath("/schema"),
							),
					).
					WithContainers(corev1apply.Container().
						WithName("nginx").
						WithImage("nginx:latest").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPorts(corev1apply.ContainerPort().
							WithContainerPort(80),
						).
						WithVolumeMounts(corev1apply.VolumeMount().
							WithName("output").
							WithMountPath("/usr/share/nginx/html"),
						),
					).
					WithVolumes(
						corev1apply.Volume().
							WithName("properties").
							WithConfigMap(corev1apply.ConfigMapVolumeSource().
								WithName(name),
							),
						corev1apply.Volume().
							WithName("output").
							WithEmptyDir(corev1apply.EmptyDirVolumeSource()),
					),
				),
			),
		)
}

func buildService(name string, namespace string, migrationVersion int) *corev1apply.ServiceApplyConfiguration {
	instanceLabel := "schemaspy-" + strconv.Itoa(migrationVersion)

	return corev1apply.Service(name, namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "schemaspy",
			"app.kubernetes.io/instance":   instanceLabel,
			"app.kubernetes.io/component":  "ops",
			"app.kubernetes.io/part-of":    "migration",
			"app.kubernetes.io/managed-by": "sspy-familllly-controller",
		}).
		WithSpec(corev1apply.ServiceSpec().
			WithSelector(map[string]string{
				"app.kubernetes.io/name": "schemaspy",
			}).
			WithType(corev1.ServiceTypeNodePort).
			WithPorts(corev1apply.ServicePort().
				WithProtocol(corev1.ProtocolTCP).
				WithPort(80).
				WithNodePort(30080),
			),
		)
}
