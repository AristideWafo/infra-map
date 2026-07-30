package scraper

import (
	"context"

	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// Scraper est le contrat que toute source de données doit respecter.
// Chaque implémentation doit être testable sans vraie connexion réseau.
type Scraper interface {
	// Name retourne l'identifiant de la source (logs et health checks).
	Name() string

	// Scrape collecte les données et retourne des UnifiedNodes normalisés.
	// Doit respecter le context (timeout, annulation).
	Scrape(ctx context.Context) ([]*models.UnifiedNode, error)

	// Health retourne nil si la source est accessible, une erreur sinon.
	Health() error

	// Connections retourne les connexions réseau découvertes (peut retourner nil).
	Connections(ctx context.Context) ([]*models.Connection, error)
}

// Noop est une implémentation vide pour les tests et les sources désactivées.
type Noop struct{ name string }

func NewNoop(name string) *Noop { return &Noop{name: name} }

func (n *Noop) Name() string { return n.name }
func (n *Noop) Scrape(ctx context.Context) ([]*models.UnifiedNode, error) {
	return nil, nil
}
func (n *Noop) Health() error { return nil }
func (n *Noop) Connections(ctx context.Context) ([]*models.Connection, error) {
	return nil, nil
}
