# Design System — InfraMaps

> Toutes les décisions visuelles sont ici. Utiliser les variables CSS définies dans ce document.
> Ne jamais hardcoder de couleur, de taille ou de typographie directement dans les composants.

---

## Principes Fondamentaux

| Principe | Application |
|----------|-------------|
| **Dark mode uniquement** | Les outils ops se lisent 8h/jour. Fond sombre = fatigue réduite. |
| **Couleur = signal** | La couleur porte de l'information (statut). Jamais décorative. |
| **Densité maximale** | Afficher le maximum d'info dans l'espace disponible. |
| **Accessibilité** | Statut communiqué par couleur + icône + texte. Jamais couleur seule. |
| **Stabilité visuelle** | Les nœuds ne bougent pas entre les refreshes. Positions fixes. |

---

## Palette de Couleurs

### Fond et surfaces

```css
:root {
  /* Fonds */
  --bg-canvas:    #0d1117; /* Fond du canvas principal */
  --bg-panel:     #161b22; /* Panneaux, cards, side panel */
  --bg-elevated:  #1c2128; /* Hover states, menus déroulants */
  --bg-input:     #0d1117; /* Champs de formulaire */

  /* Bordures */
  --border-default: #30363d;
  --border-muted:   #21262d;
  --border-strong:  #484f58;
}
```

### Statuts — Règle absolue de cohérence

```css
:root {
  /* Healthy */
  --status-healthy:        #3fb950; /* Texte, bordure, icône */
  --status-healthy-bg:     #0d2614; /* Background badge */
  --status-healthy-glow:   rgba(63, 185, 80, 0.2);

  /* Warning */
  --status-warning:        #d29922; /* Texte, bordure, icône */
  --status-warning-bg:     #271d00; /* Background badge */
  --status-warning-glow:   rgba(210, 153, 34, 0.2);

  /* Critical */
  --status-critical:       #f85149; /* Texte, bordure, icône */
  --status-critical-bg:    #2d0a0a; /* Background badge */
  --status-critical-glow:  rgba(248, 81, 73, 0.2);

  /* Unknown */
  --status-unknown:        #6e7681; /* Texte, bordure, icône */
  --status-unknown-bg:     #1c1f24; /* Background badge */
  --status-unknown-glow:   rgba(110, 118, 129, 0.1);
}
```

### Couleurs interactives

```css
:root {
  --accent:          #1f6feb; /* Actions primaires, liens, sélection */
  --accent-hover:    #388bfd;
  --accent-subtle:   rgba(31, 111, 235, 0.1); /* Background hover sur éléments */

  --connection-default:  #30363d;
  --connection-hover:    #1f6feb;
  --connection-selected: #58a6ff;
  --connection-error:    #f85149;
}
```

### Texte

```css
:root {
  --text-primary:   #e6edf3; /* Noms, valeurs importantes */
  --text-secondary: #7d8590; /* Labels, descriptions */
  --text-disabled:  #484f58; /* Éléments inactifs */
  --text-inverse:   #0d1117; /* Texte sur fond clair (badges colorés) */
}
```

---

## Typographie

### Familles de polices

```css
:root {
  /* Interface générale — labels, boutons, textes */
  --font-ui: 'Inter', system-ui, -apple-system, sans-serif;

  /* Métriques, IDs, valeurs techniques */
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
}
```

**Chargement :**
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

Ou via npm : `@fontsource/inter` + `@fontsource/jetbrains-mono`.

### Échelle typographique

```css
:root {
  --text-xs:   10px; /* IDs tronqués, labels secondaires */
  --text-sm:   11px; /* Métriques dans les nœuds compacts */
  --text-base: 13px; /* Texte courant, noms de nœuds */
  --text-md:   15px; /* Titres de panneaux */
  --text-lg:   18px; /* Titre du side panel */
  --text-xl:   24px; /* En-têtes de zones (cluster, island) */
}
```

### Règle d'usage : UI vs Mono

| Utilisation | Police | Exemple |
|-------------|--------|---------|
| Noms de nœuds | Inter | `worker-01` |
| Labels, boutons | Inter | `Filtrer`, `Namespace` |
| Valeurs CPU/RAM | JetBrains Mono | `42.5%`, `8.2 GB` |
| IDs techniques | JetBrains Mono | `550e8400-e29b-41d4` |
| Logs | JetBrains Mono | `2026-07-21 ERROR connection refused` |
| Timestamps | JetBrains Mono | `10:42:05` |

---

## Tailles des Nœuds (Grid 2D)

```css
:root {
  /* Nœuds feuilles */
  --node-pod-w:        80px;
  --node-pod-h:        60px;
  --node-container-w:  80px; /* Même taille que pod */
  --node-container-h:  60px;
  --node-vm-w:        120px;
  --node-vm-h:         88px;

  /* Zones (taille minimale — s'étendent pour contenir leurs enfants) */
  --zone-node-min-w:  280px;
  --zone-node-min-h:  200px;
  --zone-cluster-min-w: 640px;
  --zone-island-min-w:  480px;

  /* Espacement interne */
  --node-padding:       16px; /* Padding dans les cartes */
  --zone-padding:       24px; /* Padding dans les zones */
  --grid-gap:           12px; /* Gap entre nœuds dans une zone */

  /* Bordures */
  --node-border-radius:  6px;
  --zone-border-radius: 10px;
  --border-width:        1px;
  --border-width-active: 2px;
}
```

---

## États des Nœuds

### 1. État normal

```
┌─────────────────────────┐ ← border: --border-default (1px)
│ ■ api-frontend-1        │ ← statut icon + nom (Inter, --text-base)
├─────────────────────────┤
│  CPU  42.5%   MEM 256MB │ ← valeurs (JetBrains Mono, --text-sm)
└─────────────────────────┘
  background: --bg-panel
  border-left: 3px solid --status-healthy
```

### 2. Hover

```css
.node:hover {
  border-color: var(--accent);
  background: var(--bg-elevated);
  cursor: pointer;
  box-shadow: 0 0 0 1px var(--accent-subtle);
}
```

### 3. Sélectionné (click)

```css
.node.selected {
  border-color: var(--accent);
  border-width: var(--border-width-active);
  box-shadow: 0 0 0 3px var(--accent-subtle);
}
```

### 4. En alerte (pulse animation)

```css
@keyframes alert-pulse {
  0%, 100% { box-shadow: 0 0 0 0 var(--status-critical-glow); }
  50%       { box-shadow: 0 0 0 8px transparent; }
}

.node.alerting {
  animation: alert-pulse 2s ease-in-out infinite;
  border-color: var(--status-critical);
}
```

### 5. Chargement (skeleton)

```css
@keyframes skeleton-shimmer {
  0%   { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

.node.loading {
  background: linear-gradient(90deg, var(--bg-panel) 25%, var(--bg-elevated) 50%, var(--bg-panel) 75%);
  background-size: 200%;
  animation: skeleton-shimmer 1.5s infinite;
}
```

### 6. État vide (source indisponible)

Texte centré dans la zone : `Aucune donnée — vérifier [source]`.  
Icône : ⚠ en `--status-unknown`. Pas de message d'erreur technique.

### 7. Erreur source (scraper down)

Badge dans le coin supérieur droit de la zone : `Source indisponible`.  
Les données périmées restent affichées avec opacité réduite (0.5).

---

## Icônes de Statut (Accessibilité Daltoniens)

```
Healthy  → ✓ (checkmark vert) + texte "Healthy" masqué visuellement (aria-label)
Warning  → ⚠ (triangle ambre)  + texte "Warning"
Critical → ✕ (croix rouge)     + texte "Critical"
Unknown  → ? (point gris)      + texte "Unknown"
```

Utiliser `aria-label` sur chaque badge de statut.

---

## Connexions Réseau (SVG)

```css
:root {
  --connection-width:        1.5px;
  --connection-width-hover:  2.5px;
  --connection-opacity:      0.6;
  --connection-opacity-hover: 1;
}
```

**Rendu SVG :**
```
- Lignes droites (pas de courbes — plus lisible en grid dense)
- Flèche remplie à la destination (triangle 8×6px)
- stroke-dasharray: 6 4 pour les connexions inférées (topologie logique)
- stroke-dasharray: none pour les connexions avec trafic réel (service mesh)
```

---

## Side Panel

```css
:root {
  --panel-width:          360px;
  --panel-header-height:   56px;
  --panel-tab-height:      40px;
}
```

**Onglets :** Métriques | Logs | Détails  
**Animation d'ouverture :** `transform: translateX(360px → 0)`, `transition: 200ms ease-out`  
**Fermeture :** clic sur le fond ou sur ✕

---

## Toolbar (barre supérieure)

```css
:root {
  --toolbar-height: 48px;
  --toolbar-bg:     var(--bg-panel);
  --toolbar-border: 1px solid var(--border-default);
}
```

**Contenu de gauche à droite :**
`[Logo InfraMaps]  [Breadcrumb]  [spacer]  [Search]  [Filtres]  [Toggle Grid/3D]  [Refresh]`

---

## Animations — Règles Globales

| Principe | Valeur |
|----------|--------|
| Durée max (données) | 150ms — les données changent vite |
| Durée (UI transitions) | 200ms — ouverture panel, menus |
| Durée (alertes pulse) | 2s — attention sans distraction |
| Easing standard | `ease-out` |
| Désactiver si `prefers-reduced-motion` | Oui — respecter la préférence système |

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## CSS Variables — Fichier Global

> Créer `src/styles/variables.css` et l'importer dans `main.tsx`.

```css
/* src/styles/variables.css */
:root {
  /* Coller ici toutes les variables définies dans ce document */
  /* Fonds */
  --bg-canvas: #0d1117;
  --bg-panel: #161b22;
  --bg-elevated: #1c2128;
  /* Statuts */
  --status-healthy: #3fb950;
  --status-warning: #d29922;
  --status-critical: #f85149;
  --status-unknown: #6e7681;
  /* ... etc. */
}
```

**Règle stricte :** Toute valeur de couleur, taille ou police dans un composant doit être une `var(--xxx)`. Zéro valeur hardcodée.
