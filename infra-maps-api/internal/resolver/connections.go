package resolver

import (
	"context"
	"fmt"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// IngressConnections découvre Ingress → Service : chaque backend référencé
// par une règle (ou le defaultBackend) produit un lien "ingress". Comme pour
// Service → Pod, aucune télémétrie de service mesh, uniquement la topologie.
func IngressConnections(ctx context.Context, client kubernetes.Interface) ([]*models.Connection, error) {
	ingresses, err := client.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("k8s list ingresses: %w", err)
	}

	var conns []*models.Connection
	for _, ing := range ingresses.Items {
		ingID := "ingress-" + ing.Namespace + "-" + ing.Name

		addBackend := func(svcName string) {
			if svcName == "" {
				return
			}
			svcID := "svc-" + ing.Namespace + "-" + svcName
			conns = append(conns, &models.Connection{
				ID:       ingID + "-" + svcID,
				FromID:   ingID,
				ToID:     svcID,
				Protocol: "tcp",
				Type:     "ingress",
			})
		}

		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
			addBackend(ing.Spec.DefaultBackend.Service.Name)
		}
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service != nil {
					addBackend(path.Backend.Service.Name)
				}
			}
		}
	}
	return conns, nil
}

// K8sConnections découvre la topologie logique Service → Pod via les
// Endpoints. Sans service mesh, traffic/latency/errors restent à 0
// (DATA_MODELS.md) : c'est une topologie, pas de la télémétrie.
func K8sConnections(ctx context.Context, client kubernetes.Interface) ([]*models.Connection, error) {
	endpoints, err := client.CoreV1().Endpoints("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("k8s list endpoints: %w", err)
	}

	var conns []*models.Connection
	for _, ep := range endpoints.Items {
		svcID := "svc-" + ep.Namespace + "-" + ep.Name
		for _, subset := range ep.Subsets {
			for _, addr := range subset.Addresses {
				if addr.TargetRef == nil || addr.TargetRef.Kind != "Pod" {
					continue
				}
				podUID := string(addr.TargetRef.UID)
				conns = append(conns, &models.Connection{
					ID:       svcID + "-" + podUID,
					FromID:   svcID,
					ToID:     podUID,
					Protocol: "tcp",
					Type:     "service",
				})
			}
		}
	}
	return conns, nil
}
