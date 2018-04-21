package kube

import (
	log "github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

const (
	// The domain when --domain is not provided by user. will be used for configuring Host based ingress rule
	defaultIngressDomain = "datacol.io"

	ingressAnnotationName = "kubernetes.io/ingress.class"
	ingressClassName      = "nginx"

	ingressDefaultNamespace = "ingress-nginx"
	nginxAppName            = "ingress-nginx"
	nginxConfigName         = "nginx-configuration"
	defaultBackendName      = "default-http-backend"

	// inspired by https://github.com/kubernetes/ingress-nginx/blob/master/deploy/rbac.yaml
	ingRoleName                 = "nginx-ingress-role"
	nginxControllerName         = "nginx-ingress-controller"
	ingServiceAccountName       = "nginx-ingress-serviceaccount"
	ingClusterRoleName          = "nginx-ingress-clusterrole"
	nginxRoleBindingName        = "nginx-ingress-role-nisa-binding"
	nginxClusterRoleBindingName = "nginx-ingress-clusterrole-nisa-binding"
)

// FIXME: Right now deploying the nginx controller in the same namespace as of an stack and all of the apps in the apps can use the load-balancer
type awsIngress struct {
	Client          *kubernetes.Clientset
	ParentNamespace string // the stack namespace
	Namespace       string // the namespace in which we manage ingress controller
	err             error
}

func (r *awsIngress) namespace() string {
	return ingressNamespace(r.ParentNamespace)
}

func (r *awsIngress) CreateOrUpdate() error {
	ns := r.namespace()

	if ns != r.ParentNamespace {
		r.setupNamespace(ns)
	}

	r.setupRBACRoles(ns)
	r.setupDefaultBackend(ns)
	r.setupIngressController(ns)

	return r.err
}

func (r *awsIngress) Remove() error {
	ns := r.namespace()
	log.Debugf("Cleaning up ingress controller from namespace: %s", ns)

	checkError := func(err error) {
		if err != nil {
			log.Error(err)
		}
	}
	gop, dop := metav1.GetOptions{}, &metav1.DeleteOptions{}

	// Remove default backend
	checkError(r.Client.Core().Services(ns).Delete(defaultBackendName, dop))

	if dp, err := r.Client.Extensions().Deployments(ns).Get(defaultBackendName, gop); err != nil {
		log.Warn(err)
	} else {
		*dp.Spec.Replicas = 0
		if _, err = r.Client.Extensions().Deployments(ns).Update(dp); err != nil {
			log.Warn(err)
		}

		checkError(r.Client.Extensions().Deployments(ns).Delete(defaultBackendName, dop))
	}

	checkError(r.Client.Core().ConfigMaps(ns).Delete(nginxControllerName, dop))

	checkError(r.Client.Core().Services(ns).Delete(nginxControllerName, dop))

	if dp, err := r.Client.Extensions().Deployments(ns).Get(nginxControllerName, gop); err != nil {
		log.Warn(err)
	} else {
		*dp.Spec.Replicas = 0
		if _, err = r.Client.Extensions().Deployments(ns).Update(dp); err != nil {
			log.Warn(err)
		}

		checkError(r.Client.Extensions().Deployments(ns).Delete(nginxControllerName, dop))
	}

	checkError(r.Client.Core().ServiceAccounts(ns).Delete(ingServiceAccountName, dop))

	checkError(r.Client.RbacV1beta1().Roles(ns).Delete(ingRoleName, dop))
	checkError(r.Client.RbacV1beta1().RoleBindings(ns).Delete(nginxRoleBindingName, dop))

	checkError(r.Client.RbacV1beta1().ClusterRoles().Delete(ingClusterRoleName, dop))
	checkError(r.Client.RbacV1beta1().ClusterRoleBindings().Delete(nginxClusterRoleBindingName, dop))

	if ns != r.ParentNamespace {
		checkError(r.Client.Core().Namespaces().Delete(ns, dop))
	}

	return nil
}

func (r *awsIngress) setupNamespace(ns string) {
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}

	_, err := r.Client.Core().Namespaces().Create(&namespace)
	if !kerrors.IsAlreadyExists(err) {
		r.err = err
	}
}

func (r *awsIngress) setupRBACRoles(ns string) {
	if r.err != nil {
		return
	}

	account := v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: ingServiceAccountName}}

	_, err := r.Client.Core().ServiceAccounts(ns).Create(&account)
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			r.err = err
			return
		}
	}

	clusterRole := rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: ingClusterRoleName},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "endpoints", "nodes", "pods", "secrets"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"ingresses/status"},
				Verbs:     []string{"update"},
			},
		},
	}

	role := rbac.Role{
		ObjectMeta: metav1.ObjectMeta{Name: ingRoleName},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "pods", "secrets", "namespaces"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				Verbs:         []string{"get", "update"},
				ResourceNames: []string{"ingress-controller-leader-nginx"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create"},
			},
		},
	}

	roleBinding := rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: nginxRoleBindingName,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     ingRoleName,
		},
		Subjects: []rbac.Subject{{
			Kind:      rbac.ServiceAccountKind,
			Name:      ingServiceAccountName,
			Namespace: ns,
		}},
	}

	clusterRoleBinding := rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: nginxClusterRoleBindingName},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Name:     ingClusterRoleName,
			Kind:     "ClusterRole",
		},
		Subjects: []rbac.Subject{{
			Kind:      rbac.ServiceAccountKind,
			Name:      ingServiceAccountName,
			Namespace: ns,
		}},
	}

	log.Debugf("creating role in %s: %s", ns, toJson(role))

	_, err = r.Client.Rbac().Roles(ns).Create(&role)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.err = err
		return
	}

	log.Debugf("creating clusterrole in %s: %s", ns, toJson(clusterRole))

	_, err = r.Client.Rbac().ClusterRoles().Create(&clusterRole)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.err = err
		return
	}

	log.Debugf("creating rolebinding in %s: %s", ns, toJson(roleBinding))

	_, err = r.Client.Rbac().RoleBindings(ns).Create(&roleBinding)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.err = err
		return
	}

	log.Debugf("creating clusterrolebinding in %s: %s", ns, toJson(clusterRoleBinding))

	_, err = r.Client.Rbac().ClusterRoleBindings().Create(&clusterRoleBinding)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.err = err
		return
	}
}

func (r *awsIngress) setupDefaultBackend(ns string) {
	if r.err != nil {
		log.Errorf("Skipping default-backend setup setup on AWS b/c of %v", r.err)
		return
	}

	replica, labels := int32(1), map[string]string{appLabel: defaultBackendName}
	cpuLimit := resource.NewQuantity(10, resource.DecimalSI)
	reqLimit := resource.NewQuantity(20, resource.BinarySI)

	defaultBackendSvc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   defaultBackendName,
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			}},
			Selector: map[string]string{
				appLabel: defaultBackendName,
			},
		},
	}

	defaultbackendDp := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   defaultBackendName,
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replica,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  defaultBackendName,
						Image: "gcr.io/google_containers/defaultbackend:1.4",
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path:   "/healthz",
									Port:   intstr.FromInt(8080),
									Scheme: v1.URISchemeHTTP,
								},
							},
							InitialDelaySeconds: 30,
							TimeoutSeconds:      5,
						},

						Ports: []v1.ContainerPort{{
							ContainerPort: 8080,
							Protocol:      v1.ProtocolTCP,
						}},
					}},
				},
			},
		},
	}

	if false {
		//FIXME: Not enabling the the resources for now
		defaultbackendDp.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    *cpuLimit,
				v1.ResourceMemory: *reqLimit,
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    *cpuLimit,
				v1.ResourceMemory: *reqLimit,
			},
		}
	}

	if _, r.err = r.Client.Extensions().Deployments(ns).Get(defaultbackendDp.Name, metav1.GetOptions{}); r.err != nil {
		if kerrors.IsNotFound(r.err) {
			_, r.err = r.Client.Extensions().Deployments(ns).Create(&defaultbackendDp)
		}
	}

	if r.err != nil {
		return
	}

	if _, r.err = r.Client.Core().Services(ns).Get(defaultBackendSvc.Name, metav1.GetOptions{}); r.err != nil {
		if kerrors.IsNotFound(r.err) {
			_, r.err = r.Client.Core().Services(ns).Create(&defaultBackendSvc)
		}
	}

}

func (r *awsIngress) setupIngressController(ns string) {
	if r.err != nil {
		log.Errorf("Skipping nginx-controller setup setup on AWS b/c of %v", r.err)
		return
	}

	replica, labels := int32(1), map[string]string{appLabel: nginxAppName}

	nginxConfig := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nginxConfigName,
			Labels: labels,
		},
		Data: map[string]string{
			"use-proxy-protocol": "true",
		},
	}

	if _, r.err = r.Client.Core().ConfigMaps(ns).Get(nginxConfig.Name, metav1.GetOptions{}); r.err != nil {
		if kerrors.IsNotFound(r.err) {
			_, r.err = r.Client.Core().ConfigMaps(ns).Create(&nginxConfig)
		}
	}

	if r.err != nil {
		return
	}

	nginxControllerService := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: nginxAppName,
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-proxy-protocol":          "*",
				"service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout": "3600",
			},
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeLoadBalancer,
			Selector: labels,
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http"),
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromString("https"),
				},
			},
		},
	}

	probeHandler := v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path:   "/healthz",
			Port:   intstr.FromInt(10254),
			Scheme: v1.URISchemeHTTP,
		},
	}

	nginxControllerDp := v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: nginxControllerName,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nginxControllerName,
					Labels:    labels,
					Namespace: ns,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: ingServiceAccountName,
					Containers: []v1.Container{{
						Name:            nginxControllerName,
						Image:           "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.10.2",
						ImagePullPolicy: v1.PullIfNotPresent,
						Env: []v1.EnvVar{
							{
								Name: "POD_NAME",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name: "POD_NAMESPACE",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
						},
						LivenessProbe: &v1.Probe{
							Handler:             probeHandler,
							InitialDelaySeconds: 30,
							TimeoutSeconds:      5,
						},
						ReadinessProbe: &v1.Probe{
							Handler:          probeHandler,
							FailureThreshold: 3,
						},
						Args: []string{
							"/nginx-ingress-controller",
							"--default-backend-service=$(POD_NAMESPACE)/" + defaultBackendName,
							"--configmap=$(POD_NAMESPACE)/" + nginxConfigName,
							"--publish-service=$(POD_NAMESPACE)/" + nginxAppName,
							"--annotations-prefix=kubernetes.io",
						},
						Ports: []v1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
								Protocol:      v1.ProtocolTCP,
							},
							{
								Name:          "https",
								ContainerPort: 443,
								Protocol:      v1.ProtocolTCP,
							},
						},
					}},
				},
			},
		},
	}

	//Note: Nginx controller pod will expect the default backend service to be created

	if _, r.err = r.Client.Core().Services(ns).Get(nginxControllerService.Name, metav1.GetOptions{}); r.err != nil {
		if kerrors.IsNotFound(r.err) {
			_, r.err = r.Client.Core().Services(ns).Create(&nginxControllerService)
		}
	}

	if r.err != nil {
		return
	}

	if _, r.err = r.Client.Extensions().Deployments(ns).Get(nginxControllerDp.Name, metav1.GetOptions{}); r.err != nil {
		if kerrors.IsNotFound(r.err) {
			_, r.err = r.Client.Extensions().Deployments(ns).Create(&nginxControllerDp)
		}
	}
}
