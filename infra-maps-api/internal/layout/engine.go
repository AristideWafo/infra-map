package layout

import (
	"sort"

	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// Engine calcule des positions Grid 2D déterministes (ADR-003).
// Même arbre en entrée → mêmes positions en sortie, toujours.
// Les positions sont relatives au parent ; le front les compose.
type Engine struct {
	cellW   float64
	cellH   float64
	gap     float64
	columns int
}

func NewEngine() *Engine {
	return &Engine{cellW: 160, cellH: 120, gap: 24, columns: 4}
}

// Apply positionne chaque niveau de l'arbre sur une grille.
// Les enfants sont triés par ID avant placement : l'ordre d'arrivée
// des scrapers n'influence jamais les positions.
func (e *Engine) Apply(root *models.UnifiedNode) {
	if root == nil {
		return
	}
	root.X, root.Y, root.Z = 0, 0, 0
	e.placeChildren(root)
}

func (e *Engine) placeChildren(n *models.UnifiedNode) {
	children := n.Children
	sort.Slice(children, func(i, j int) bool { return children[i].ID < children[j].ID })

	for i, c := range children {
		col := i % e.columns
		row := i / e.columns
		c.X = float64(col) * (e.cellW + e.gap)
		c.Y = float64(row) * (e.cellH + e.gap)
		c.Z = 0
		e.placeChildren(c)
	}
}
