package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
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
