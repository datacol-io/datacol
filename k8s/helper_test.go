package kube

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestMergeIngressRules(t *testing.T) {
	{
		ing1 := v1beta1.Ingress{Spec: v1beta1.IngressSpec{
			Rules: ingressRulesManifest("app1", "", intstr.FromInt(80), []string{"a1.com"}),
		}}

		ing2 := v1beta1.Ingress{Spec: v1beta1.IngressSpec{
			Rules: ingressRulesManifest("app2", "", intstr.FromInt(80), []string{"a2.com"}),
		}}

		ing := mergeIngressRules(&ing1, &ing2)
		assert.Len(t, ing.Spec.Rules, 2)

		ing3 := mergeIngressRules(ing, &ing2)
		assert.Len(t, ing3.Spec.Rules, 2)

		ing2.Spec.Rules[0].Host = "b1.com"
		ing4 := mergeIngressRules(ing3, &ing2)
		assert.Len(t, ing4.Spec.Rules, 2)
	}

	{
		ing1 := v1beta1.Ingress{Spec: v1beta1.IngressSpec{
			Rules: ingressRulesManifest("app1", "", intstr.FromInt(80), []string{"a1.com", "a2.com"}),
		}}

		ing2 := v1beta1.Ingress{Spec: v1beta1.IngressSpec{
			Rules: ingressRulesManifest("app1", "", intstr.FromInt(80), []string{"a1.com", "a2.com", "b1.com"}),
		}}

		ing := mergeIngressRules(&ing1, &ing2)
		assert.Len(t, ing.Spec.Rules, 3)
	}

	{
		ing1 := v1beta1.Ingress{Spec: v1beta1.IngressSpec{
			Rules: ingressRulesManifest("app1", "", intstr.FromInt(80), []string{"a1.com", "a2.com"}),
		}}

		ing2 := v1beta1.Ingress{Spec: v1beta1.IngressSpec{
			Rules: ingressRulesManifest("app1", "", intstr.FromInt(80), []string{"a1.com"}),
		}}

		ing := mergeIngressRules(&ing1, &ing2)
		assert.Len(t, ing.Spec.Rules, 1)
	}
}
