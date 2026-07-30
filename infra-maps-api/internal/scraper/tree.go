package scraper

import (
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// BuildTree assemble une liste plate de nœuds (reliés par ParentID) en arbre
// sous une racine synthétique, puis propage les statuts vers les parents.
// Les nœuds orphelins (ParentID inconnu) sont rattachés à la racine plutôt
// que perdus — un scraper partiel ne doit jamais faire disparaître des nœuds.
func BuildTree(nodes []*models.UnifiedNode) *models.UnifiedNode {
	root := &models.UnifiedNode{
		ID:       "root",
		Name:     "Infrastructure",
		Type:     models.NodeTypeRoot,
		Status:   models.StatusHealthy,
		Source:   "internal",
		LastSeen: time.Now(),
	}

	byID := make(map[string]*models.UnifiedNode, len(nodes))
	for _, n := range nodes {
		byID[n.ID] = n
	}

	for _, n := range nodes {
		if n.ParentID == "" {
			root.Children = append(root.Children, n)
			continue
		}
		parent, ok := byID[n.ParentID]
		if !ok {
			root.Children = append(root.Children, n)
			continue
		}
		parent.Children = append(parent.Children, n)
	}

	models.PropagateStatus(root)
	return root
}
