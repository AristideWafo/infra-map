package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func TestK8sConnections_ServiceToPods(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{
				{TargetRef: &corev1.ObjectReference{Kind: "Pod", UID: types.UID("pod-uid-1")}},
				{TargetRef: &corev1.ObjectReference{Kind: "Pod", UID: types.UID("pod-uid-2")}},
				{TargetRef: nil}, // adresse sans pod cible — ignorée
			},
		}},
	})

	conns, err := K8sConnections(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, conns, 2)

	assert.Equal(t, "svc-default-api", conns[0].FromID)
	assert.Equal(t, "pod-uid-1", conns[0].ToID)
	assert.Equal(t, "service", conns[0].Type)
	assert.Equal(t, "svc-default-api-pod-uid-1", conns[0].ID)
	// Pas de service mesh : télémétrie à zéro
	assert.Zero(t, conns[0].Traffic)
}

func TestK8sConnections_EmptyCluster(t *testing.T) {
	conns, err := K8sConnections(context.Background(), fake.NewSimpleClientset())
	require.NoError(t, err)
	assert.Empty(t, conns)
}

func pathType(t networkingv1.PathType) *networkingv1.PathType { return &t }

func TestIngressConnections_RuleBackendsAndDefaultBackend(t *testing.T) {
	client := fake.NewSimpleClientset(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "default"},
		Spec: networkingv1.IngressSpec{
			DefaultBackend: &networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{Name: "fallback"},
			},
			Rules: []networkingv1.IngressRule{{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/api",
							PathType: pathType(networkingv1.PathTypePrefix),
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{Name: "api"},
							},
						}},
					},
				},
			}},
		},
	})

	conns, err := IngressConnections(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, conns, 2)

	assert.Equal(t, "ingress-default-web", conns[0].FromID)
	assert.Equal(t, "svc-default-fallback", conns[0].ToID)
	assert.Equal(t, "ingress", conns[0].Type)

	assert.Equal(t, "ingress-default-web", conns[1].FromID)
	assert.Equal(t, "svc-default-api", conns[1].ToID)
}

func TestIngressConnections_NoIngresses(t *testing.T) {
	conns, err := IngressConnections(context.Background(), fake.NewSimpleClientset())
	require.NoError(t, err)
	assert.Empty(t, conns)
}
