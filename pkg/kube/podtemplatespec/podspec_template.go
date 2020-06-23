package podtemplatespec

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Modification func(*corev1.PodTemplateSpec)

const (
	notFound = -1
)

func New(templateMods ...Modification) corev1.PodTemplateSpec {
	podTemplateSpec := corev1.PodTemplateSpec{}
	for _, templateMod := range templateMods {
		templateMod(&podTemplateSpec)
	}
	return podTemplateSpec
}

func Apply(templateMods ...Modification) Modification {
	return func(template *corev1.PodTemplateSpec) {
		for _, f := range templateMods {
			f(template)
		}
	}
}

func NOOP() Modification {
	return func(spec *corev1.PodTemplateSpec) {}
}

func WithContainer(name string, containerfunc func(*corev1.Container)) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		idx := findIndexByName(name, podTemplateSpec.Spec.Containers)
		if idx == notFound {
			// if we are attempting to modify a container that does not exist, we will add a new one
			podTemplateSpec.Spec.Containers = append(podTemplateSpec.Spec.Containers, corev1.Container{})
			idx = len(podTemplateSpec.Spec.Containers) - 1
		}
		c := &podTemplateSpec.Spec.Containers[idx]
		containerfunc(c)
	}
}

func WithContainerByIndex(index int, funcs ...func(container *corev1.Container)) func(podTemplateSpec *corev1.PodTemplateSpec) {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		if index >= len(podTemplateSpec.Spec.Containers) {
			podTemplateSpec.Spec.Containers = append(podTemplateSpec.Spec.Containers, corev1.Container{})
		}
		c := &podTemplateSpec.Spec.Containers[index]
		for _, f := range funcs {
			f(c)
		}
	}
}

func WithInitContainer(name string, containerfunc func(*corev1.Container)) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		idx := findIndexByName(name, podTemplateSpec.Spec.InitContainers)
		if idx == notFound {
			// if we are attempting to modify a container that does not exist, we will add a new one
			podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, corev1.Container{})
			idx = len(podTemplateSpec.Spec.InitContainers) - 1
		}
		c := &podTemplateSpec.Spec.InitContainers[idx]
		containerfunc(c)
	}
}

func WithInitContainerByIndex(index int, funcs ...func(container *corev1.Container)) func(podTemplateSpec *corev1.PodTemplateSpec) {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		if index >= len(podTemplateSpec.Spec.Containers) {
			podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, corev1.Container{})
		}
		c := &podTemplateSpec.Spec.InitContainers[index]
		for _, f := range funcs {
			f(c)
		}
	}
}

func WithPodLabels(labels map[string]string) Modification {
	if labels == nil {
		labels = map[string]string{}
	}
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.ObjectMeta.Labels = labels
	}
}

func WithServiceAccount(serviceAccountName string) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.ServiceAccountName = serviceAccountName
	}
}

func WithVolume(volume corev1.Volume) Modification {
	return func(template *corev1.PodTemplateSpec) {
		for _, v := range template.Spec.Volumes {
			if v.Name == volume.Name {
				return
			}
		}
		template.Spec.Volumes = append(template.Spec.Volumes, volume)
	}
}

func findIndexByName(name string, containers []corev1.Container) int {
	for idx, c := range containers {
		if c.Name == name {
			return idx
		}
	}
	return notFound
}

func WithTerminationGracePeriodSeconds(seconds int) Modification {
	s := int64(seconds)
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.TerminationGracePeriodSeconds = &s
	}
}

func WithFsGroup(fsGroup int) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		spec := &podTemplateSpec.Spec
		fsGroup64 := int64(fsGroup)
		spec.SecurityContext = &corev1.PodSecurityContext{
			FSGroup: &fsGroup64,
		}
	}
}

func WithImagePullSecrets(name string) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.ImagePullSecrets = append(podTemplateSpec.Spec.ImagePullSecrets, corev1.LocalObjectReference{
			Name: name,
		})
	}
}

func WithTopologyKey(topologyKey string, idx int) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[idx].PodAffinityTerm.TopologyKey = topologyKey
	}
}

func WithAffinity(stsName, antiAffinityLabelKey string, weight int) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.Affinity =
			&corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
						Weight: int32(weight),
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{antiAffinityLabelKey: stsName}},
						},
					}},
				},
			}
	}
}

func WithNodeAffinity(nodeAffinity *corev1.NodeAffinity) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.Affinity.NodeAffinity = nodeAffinity
	}
}

func WithPodAffinity(podAffinity *corev1.PodAffinity) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.Affinity.PodAffinity = podAffinity
	}
}

func WithTolerations(tolerations []corev1.Toleration) Modification {
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Spec.Tolerations = tolerations
	}
}

func WithAnnotations(annotations map[string]string) Modification {
	if annotations == nil {
		annotations = map[string]string{}
	}
	return func(podTemplateSpec *corev1.PodTemplateSpec) {
		podTemplateSpec.Annotations = annotations
	}
}
