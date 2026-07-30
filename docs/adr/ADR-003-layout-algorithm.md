# ADR-003 — Algorithme de Layout Déterministe

**Date :** 2026-07-21  
**Statut :** Accepté  
**Décideur :** Aristide Wafo

---

## Contexte

InfraMaps doit afficher les nœuds d'infrastructure dans un espace 2D (Grid) ou 3D (Maps). Chaque nœud a des coordonnées X, Y, Z dans le modèle de données.

**Le problème :** Si les positions sont recalculées à chaque refresh (30s), les nœuds "sautent" à l'écran — expérience utilisateur catastrophique pour un outil d'ops en production.

**La contrainte :** Les positions doivent être stables entre les refreshes tant que la topologie ne change pas.

---

## Décision

**Layout Engine déterministe basé sur l'index d'ordre des enfants dans leur parent.**

**Grid 2D :**
1. Trier les enfants d'un nœud parent par ID (tri alphabétique, stable)
2. Calculer `cols = ceil(sqrt(len(children)))`
3. Position[i] : `X = parentX + (i % cols) * (nodeWidth + gap)`, `Y = parentY + (i / cols) * (nodeHeight + gap)`
4. Appliquer récursivement

**3D Maps (Phase 4) :**
- `Y = hierarchyLevel * 150` (root=0, cluster=150, node=300, pod=450)
- `X = FNV32(nodeID) % parentWidth + parentOffsetX`
- `Z = (FNV32(nodeID) >> 16) % parentDepth + parentOffsetZ`

**Garantie :** Même topologie en entrée → mêmes positions en sortie. Testé unitairement (test d'idempotence).

**Mise à jour des positions :** Si un nouveau nœud apparaît, il prend la prochaine position disponible dans la grille. Les nœuds existants ne bougent pas.

---

## Conséquences

**Positives :**
- Positions stables → pas de "jump" à l'écran entre les refreshes
- Algorithme simple à comprendre et à déboguer
- Pas de dépendance externe (pas de bibliothèque de layout)
- Résultat identique entre deux runs (testable)
- Très rapide (O(n) où n = nombre de nœuds)

**Négatives :**
- Pas d'optimisation visuelle : des nœuds proches fonctionnellement peuvent être éloignés spatialement
- Pour très grands clusters (>100 pods par node), la grille peut devenir dense
- Pas de "beau" layout automatique comme force-directed

**Neutres :**
- Les zones (clusters, islands) s'étendent automatiquement pour contenir leurs enfants
- L'utilisateur peut naviguer drill-down pour zoomer dans une zone dense

---

## Alternatives Considérées

### Alternative A — Force-directed layout (d3-force)

Algorithme physique qui place les nœuds connectés proches les uns des autres.

**Rejeté parce que :**
- Non déterministe sans seed (positions différentes à chaque run)
- Avec seed, le calcul dépend de l'ordre d'initialisation → fragile
- Coûteux en CPU pour grandes infrastructures (itérations physiques)
- Weave Scope a essayé cette approche et c'était une des raisons de sa lourdeur
- Nodes connectés dans K8s = presque tous (services → pods) → le graphe de forces ne converge pas bien

### Alternative B — Position manuelle par l'utilisateur

L'utilisateur drag-and-drop les nœuds et les positions sont sauvegardées.

**Rejeté parce que :**
- Nécessite une base de données (violation ADR-002)
- Pour de grandes infrastructures (50+ pods), configurer manuellement est prohibitif
- Les pods sont éphémères — ils disparaissent et réapparaissent, rendant les positions manuelles obsolètes

### Alternative C — Hash déterministe du nodeID pour la 2D aussi

Utiliser FNV32(nodeID) pour la 2D comme pour la 3D.

**Rejeté pour la 2D parce que :**
- Le hash peut créer des collisions de position (deux nœuds au même endroit)
- La grille indexée est plus prévisible et plus facile à comprendre
- La grille naturelle respecte les zones parent/enfant

**Retenu pour la 3D (Phase 4) :** Le hash sur le plan XZ est adapté à l'espace 3D ouvert où les collisions sont moins critiques visuellement.
