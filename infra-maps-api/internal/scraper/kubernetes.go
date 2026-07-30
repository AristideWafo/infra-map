package scraper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const clusterNodeID = "cluster-kubernetes"

// Kubernetes scrape la topologie du cluster : cluster → nodes → pods.
// Les métriques CPU/RAM fines viennent de Prometheus (kube-state-metrics) en
// complément ; ici on porte la topologie et les statuts K8s.
type Kubernetes struct {
	client kubernetes.Interface
	now    func() time.Time

	mu      sync.RWMutex
	lastErr error
}

// NewKubernetes construit le scraper depuis un ServiceAccount in-cluster ou un
// kubeconfig local.
func NewKubernetes(inCluster bool, kubeconfig string) (*Kubernetes, error) {
	var cfg *rest.Config
	var err error
	if inCluster {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, fmt.Errorf("k8s config: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("k8s client: %w", err)
	}
	return &Kubernetes{client: client, now: time.Now}, nil
}

// NewKubernetesWithClient est le constructeur de test (fake clientset).
func NewKubernetesWithClient(client kubernetes.Interface) *Kubernetes {
	return &Kubernetes{client: client, now: time.Now}
}

func (k *Kubernetes) Name() string { return "kubernetes" }

func (k *Kubernetes) Health() error {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.lastErr
}

func (k *Kubernetes) setErr(err error) {
	k.mu.Lock()
	k.lastErr = err
	k.mu.Unlock()
}

func (k *Kubernetes) Scrape(ctx context.Context) ([]*models.UnifiedNode, error) {
	seen := k.now()

	nodeList, err := k.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("k8s list nodes: %w", err)
		k.setErr(err)
		return nil, err
	}
	podList, err := k.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("k8s list pods: %w", err)
		k.setErr(err)
		return nil, err
	}

	cluster := &models.UnifiedNode{
		ID: clusterNodeID, Name: "kubernetes", Type: models.NodeTypeCluster,
		Status: models.StatusHealthy, Source: "kubernetes", LastSeen: seen,
	}
	podCount := len(podList.Items)
	cluster.Pods = &podCount

	nodes := []*models.UnifiedNode{cluster}
	nodeUIDByName := map[string]string{}

	for i := range nodeList.Items {
		n := &nodeList.Items[i]
		uid := string(n.UID)
		nodeUIDByName[n.Name] = uid
		nodes = append(nodes, &models.UnifiedNode{
			ID: uid, Name: n.Name, Type: models.NodeTypeNode, ParentID: clusterNodeID,
			Status: nodeStatus(n), Source: "kubernetes", LastSeen: seen,
			CreatedAt: n.CreationTimestamp.Time, RawLabels: n.Labels,
		})
	}

	for i := range podList.Items {
		p := &podList.Items[i]
		parent := nodeUIDByName[p.Spec.NodeName]
		if parent == "" {
			parent = clusterNodeID // pod pas encore schedulé
		}
		restarts := podRestarts(p)
		pod := &models.UnifiedNode{
			ID: string(p.UID), Name: p.Name, Type: models.NodeTypePod, ParentID: parent,
			Status: podStatus(p, restarts), Source: "kubernetes", LastSeen: seen,
			CreatedAt: p.CreationTimestamp.Time, Namespace: p.Namespace,
			RawLabels: p.Labels,
		}
		pod.Restarts = &restarts
		nodes = append(nodes, pod)
	}

	k.setErr(nil)
	return nodes, nil
}

func (k *Kubernetes) Connections(ctx context.Context) ([]*models.Connection, error) {
	return nil, nil // Connection resolver — commit dédié
}

func nodeStatus(n *corev1.Node) models.NodeStatus {
	for _, cond := range n.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				return models.StatusHealthy
			}
			return models.StatusCritical
		}
	}
	return models.StatusUnknown
}

func podStatus(p *corev1.Pod, restarts int) models.NodeStatus {
	for _, cs := range p.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
			return models.StatusCritical
		}
	}
	switch p.Status.Phase {
	case corev1.PodFailed:
		return models.StatusCritical
	case corev1.PodPending, corev1.PodUnknown:
		return models.StatusUnknown
	}
	if restarts > 0 {
		return models.StatusWarning
	}
	return models.StatusHealthy
}

func podRestarts(p *corev1.Pod) int {
	total := 0
	for _, cs := range p.Status.ContainerStatuses {
		total += int(cs.RestartCount)
	}
	return total
}
