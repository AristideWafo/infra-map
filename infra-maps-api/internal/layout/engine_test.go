package layout

import (
	"testing"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func buildTestTree() *models.UnifiedNode {
	return &models.UnifiedNode{
		ID: "root",
		Children: []*models.UnifiedNode{
			{ID: "cluster-prod", Children: []*models.UnifiedNode{
				{ID: "pod-b"}, {ID: "pod-a"}, {ID: "pod-c"},
			}},
			{ID: "vm-island-dc1", Children: []*models.UnifiedNode{
				{ID: "vm-2"}, {ID: "vm-1"},
			}},
		},
	}
}

func TestEngine_Idempotent(t *testing.T) {
	engine := NewEngine()
	t1 := buildTestTree()
	t2 := buildTestTree()

	engine.Apply(t1)
	engine.Apply(t2)

	var walk func(a, b *models.UnifiedNode)
	walk = func(a, b *models.UnifiedNode) {
		assert.Equal(t, a.X, b.X, "X of %s", a.ID)
		assert.Equal(t, a.Y, b.Y, "Y of %s", a.ID)
		assert.Equal(t, float64(0), a.Z)
		for i := range a.Children {
			walk(a.Children[i], b.Children[i])
		}
	}
	walk(t1, t2)
}

func TestEngine_OrderIndependent(t *testing.T) {
	engine := NewEngine()

	shuffled := &models.UnifiedNode{ID: "root", Children: []*models.UnifiedNode{
		{ID: "b"}, {ID: "a"}, {ID: "c"},
	}}
	ordered := &models.UnifiedNode{ID: "root", Children: []*models.UnifiedNode{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}}

	engine.Apply(shuffled)
	engine.Apply(ordered)

	// Après tri par ID, "a" est en premier dans les deux cas, même position.
	assert.Equal(t, shuffled.Children[0].ID, ordered.Children[0].ID)
	assert.Equal(t, shuffled.Children[0].X, ordered.Children[0].X)
	assert.Equal(t, shuffled.Children[2].X, ordered.Children[2].X)
}

func TestEngine_GridWraps(t *testing.T) {
	engine := NewEngine()
	tree := &models.UnifiedNode{ID: "root", Children: []*models.UnifiedNode{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"},
	}}
	engine.Apply(tree)

	// 4 colonnes : le 5e nœud passe sur la ligne suivante, colonne 0.
	assert.Equal(t, float64(0), tree.Children[4].X)
	assert.Greater(t, tree.Children[4].Y, tree.Children[0].Y)
}

func TestEngine_NilSafe(t *testing.T) {
	assert.NotPanics(t, func() { NewEngine().Apply(nil) })
}
