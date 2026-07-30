package scraper

import (
	"context"
	"testing"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func k8sNode(name, uid string, ready bool) *corev1.Node {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(uid)},
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: status},
		}},
	}
}

func k8sPod(name, uid, ns, nodeName string, phase corev1.PodPhase, restarts int32, crashLoop bool) *corev1.Pod {
	cs := corev1.ContainerStatus{RestartCount: restarts}
	if crashLoop {
		cs.State = corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(uid), Namespace: ns},
		Spec:       corev1.PodSpec{NodeName: nodeName},
		Status:     corev1.PodStatus{Phase: phase, ContainerStatuses: []corev1.ContainerStatus{cs}},
	}
}

func TestKubernetes_ScrapeTopology(t *testing.T) {
	client := fake.NewSimpleClientset(
		k8sNode("worker-01", "node-uid-1", true),
		k8sPod("api-1", "pod-uid-1", "default", "worker-01", corev1.PodRunning, 0, false),
		k8sPod("job-1", "pod-uid-2", "jobs", "worker-01", corev1.PodRunning, 3, false),
	)
	k := NewKubernetesWithClient(client)

	nodes, err := k.Scrape(context.Background())
	require.NoError(t, err)
	require.Len(t, nodes, 4) // cluster + 1 node + 2 pods

	cluster := nodes[0]
	assert.Equal(t, clusterNodeID, cluster.ID)
	assert.Equal(t, 2, *cluster.Pods)

	worker := nodes[1]
	assert.Equal(t, "node-uid-1", worker.ID)
	assert.Equal(t, clusterNodeID, worker.ParentID)
	assert.Equal(t, models.StatusHealthy, worker.Status)

	byName := map[string]*models.UnifiedNode{}
	for _, n := range nodes {
		byName[n.Name] = n
	}
	api := byName["api-1"]
	assert.Equal(t, "node-uid-1", api.ParentID)
	assert.Equal(t, "default", api.Namespace)
	assert.Equal(t, models.StatusHealthy, api.Status)

	job := byName["job-1"]
	assert.Equal(t, models.StatusWarning, job.Status) // restarts > 0
	assert.Equal(t, 3, *job.Restarts)
}

func TestKubernetes_NodeNotReadyIsCritical(t *testing.T) {
	client := fake.NewSimpleClientset(k8sNode("worker-01", "node-uid-1", false))
	k := NewKubernetesWithClient(client)

	nodes, err := k.Scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, models.StatusCritical, nodes[1].Status)
}

func TestKubernetes_CrashLoopBackOffIsCritical(t *testing.T) {
	client := fake.NewSimpleClientset(
		k8sNode("worker-01", "node-uid-1", true),
		k8sPod("bad-pod", "pod-uid-1", "default", "worker-01", corev1.PodRunning, 12, true),
	)
	k := NewKubernetesWithClient(client)

	nodes, err := k.Scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, models.StatusCritical, nodes[2].Status)
}

func TestKubernetes_UnscheduledPodAttachedToCluster(t *testing.T) {
	client := fake.NewSimpleClientset(
		k8sPod("pending-pod", "pod-uid-1", "default", "", corev1.PodPending, 0, false),
	)
	k := NewKubernetesWithClient(client)

	nodes, err := k.Scrape(context.Background())
	require.NoError(t, err)
	pod := nodes[1]
	assert.Equal(t, clusterNodeID, pod.ParentID)
	assert.Equal(t, models.StatusUnknown, pod.Status)
}

func TestKubernetes_HealthReflectsLastScrape(t *testing.T) {
	client := fake.NewSimpleClientset()
	k := NewKubernetesWithClient(client)

	_, err := k.Scrape(context.Background())
	require.NoError(t, err)
	assert.NoError(t, k.Health())
}
