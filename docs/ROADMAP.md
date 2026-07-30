# Roadmap — InfraMaps

> **Règle de la roadmap :** chaque phase est fonctionnelle et utile **indépendamment** des phases suivantes.  
> On ne commence pas une phase si la précédente n'est pas stable et testée.  
> Les dates sont des estimations — la condition de sortie prime sur le calendrier.
>
> **Règle non négociable ajoutée (2026-07-29) :** aucune tâche de code n'est considérée "faite" sans test unitaire qui la couvre, et aucun merge sur `main` n'arrive en production sans passer par la CI + le pipeline de release automatisé (voir Phase 0.5). Ce n'est plus une tâche de Phase 2 — c'est un prérequis transverse à toutes les phases.

---

## ⚠️ Constat de réalité (2026-07-29)

Cette roadmap déclarait la Phase 0 "✅ Done" et documentait déjà toute l'arborescence (`internal/scraper`, `internal/layout`, Zustand store, hooks, etc.) dans `BACKEND_GUIDE.md` / `FRONTEND_GUIDE.md`. En auditant le code réel :

- `infra-maps-api` : un seul fichier (`cmd/server/main.go`), un handler `/health` qui répond `{"status":"ok"}`. Tous les dossiers `internal/{api,scraper,layout,resolver,cache,models}` sont **vides**. Aucun scraper, aucun cache, aucun layout engine n'existe.
- `infra-maps-ui` : le boilerplate Vite par défaut (compteur + logos React/Vite). Aucun `store/`, `hooks/`, `views/`, `components/` — la structure documentée dans `FRONTEND_GUIDE.md` n'a pas encore de code derrière.
- **Zéro test** dans tout le repo (`0` fichier `*_test.go`, `0` fichier `*.test.tsx`).
- **Zéro CI/CD** — aucun `.github/workflows`, aucune release GitHub, pas de convention de commit.

Rien de grave — c'est normal à ce stade — mais la doc ne doit pas raconter une histoire en avance sur le code : ça trompe la priorisation. **Phase 0 est donc rétrogradée à "🔄 En cours"**, et une Phase 0.5 est insérée pour poser les fondations d'automatisation et de tests avant d'attaquer les vrais scrapers, comme demandé.

---

## Vue d'ensemble

```
Phase 0   ✅  MVP Grid 2D + API Go (arbre mocké de bout en bout, testé)
Phase 0.5 ✅  Fondations CI/CD + tests unitaires + releases GitHub automatiques (reste : protection de branche — manuel)
Phase 1   🔄  Scrapers réels — code livré, validation cluster réel restante
Phase 2   🔄  Utilisabilité quotidienne — P0+P1+P2 livrés (Loki, WS alertes, Helm, breadcrumb, search)
Phase 3   🔄  Robustesse production — stale-serving, timeouts, URL state, Échap livrés
Phase 4   ⏳  Vue 3D Maps — Three.js [CONDITIONNEL]
Phase 5   ⏳  Actions production [PRUDENCE — décision future]
```

---

## Phase 0 — ✅ MVP Fonctionnel (clos le 2026-07-30)

**Livré :**
- Backend complet sur l'architecture cible : `models` (UnifiedNode + calcul/propagation de statut), `cache` (TTL mémoire), `layout` (grid déterministe, tri par ID), `scraper` (interface + `MockPrometheus` + orchestrateur parallèle avec timeout et mode dégradé), handlers `tree/nodes/tags/connections/health` conformes à `API_CONTRACT.md` (filtres namespace/status/tag/source/maxDepth, codes `ERR_*`, `Cache-Control: max-age=25`)
- Frontend : store Zustand, hook `useInfraTree` (polling 30s), GridView (zones récursives + NodeCard), SidePanel (métriques instantanées), design system appliqué via `variables.css` (zéro couleur hardcodée)
- Tests : Go 80% de couverture sur `internal/` (`-race` vert), 20 tests Vitest/RTL, lint + build verts

**Condition de sortie validée :** l'interface tourne localement, affiche l'arbre mocké via l'API Go (cluster + VMs, statuts propagés, panel de détails), chaque module a ses tests.

**Note Phase 1 :** remplacer `MockPrometheus` par le vrai scraper = implémenter la même interface, zéro changement ailleurs.

---

## Phase 0.5 — 🔄 Fondations CI/CD, Tests & Releases Automatiques

**Estimé :** 3–5 jours — **bloquant avant Phase 1**, à la demande explicite du projet : tout doit être automatique, y compris les releases GitHub.

**Objectif :** Aucun code n'est mergé sans passer des tests automatisés ; chaque merge sur `main` produit, sans intervention manuelle, une version, un changelog et une GitHub Release.

### Fait (2026-07-29)

| Tâche | Statut |
|-------|--------|
| Handler `/health` extrait dans `internal/api/handlers`, testable, testé (testify) | ✅ |
| Vitest + React Testing Library installés, config `vite.config.ts`, test de fumée `App.test.tsx` | ✅ |
| `.github/workflows/backend-ci.yml` — `go vet`, `go test -race`, seuil de couverture 60% sur `internal/`, `go build` | ✅ |
| `.github/workflows/frontend-ci.yml` — `oxlint`, `vitest run`, `vite build`, path-filtré (ne tourne que si `infra-maps-ui/**` change) | ✅ |
| `release-please` configuré en mode monorepo (`infra-maps-api` = release-type `go`, `infra-maps-ui` = release-type `node`), manifest séparé par package | ✅ |
| `.github/workflows/release-please.yml` — ouvre la PR de release et l'auto-merge dès que les checks CI passent (`autorelease: pending`) | ✅ |
| `CONTRIBUTING.md` — convention Conventional Commits (obligatoire pour que release-please fonctionne) | ✅ |

### Restant

| Tâche | Priorité | Notes |
|-------|----------|-------|
| Activer la protection de branche `prod` : check `Lint, test, build` (backend + frontend) requis, `Allow auto-merge` sur le repo | 🔴 P0 | **Seule étape manuelle restante** — `gh auth login` puis `gh api -X PUT repos/AristideWafo/infra-map/branches/prod/protection ...` (ou via Settings GitHub) |

Fait depuis (2026-07-30) : Dockerfile multi-stage distroless ✅, job GHCR sur release ✅ (tag image = tag release), Conventional Commits adoptés sur tous les commits ✅, workflows repointés sur `prod` (ils visaient `main` qui n'existe pas) ✅. Les PRs release-please s'ouvrent bien à chaque push.

### Condition de sortie ✅

> Je merge une PR avec un commit `feat(infra-maps-api): ...` sur `main` → une PR de release s'ouvre automatiquement, se merge seule dès que la CI est verte, et une GitHub Release versionnée apparaît sans que j'aie tapé une seule commande `git tag`.

---

## Phase 1 — 🔄 Scrapers Réels (code livré le 2026-07-30 — validation sur cluster réel restante)

**Objectif :** InfraMaps affiche les vraies données de mon infrastructure.

**Livré :**
- Scraper Prometheus réel : discovery VMs via `/api/v1/targets`, CPU/RAM/Disk en PromQL, islands par label `datacenter`, target down → critical, métriques absentes → unknown. Testé avec fake `promAPI`.
- Scraper Kubernetes (client-go) : cluster → nodes → pods, CrashLoopBackOff → critical, restarts → warning, pods non schedulés rattachés au cluster. Testé avec fake clientset.
- Scraper Docker : containers + groupes compose, mapping état → statut. Testé avec fake API.
- Connection Resolver : Service → Pod via Endpoints (`internal/resolver`).
- Config env complète (`internal/config`), fallback mock si `PROMETHEUS_URL` vide, sources désactivables.
- Front : FilterBar namespace/statut/tag branchée (filtrage serveur), ConnectionSVG (lignes pointillées), hooks connections/tags.
- `docker-compose.dev.yml` (Prometheus + node_exporter + api).

**Restant pour clore :** brancher sur un cluster réel (ou kind) et valider la condition de sortie ; métriques CPU/RAM des pods K8s via PromQL `kube_state_metrics` (aujourd'hui topologie + statuts seulement).

### Backend

| Tâche | Priorité | Complexité |
|-------|----------|-----------|
| Prometheus scraper réel (PromQL CPU/RAM/Disk) | 🔴 P0 | Moyenne |
| VM scraper (Prometheus `/api/v1/targets` + node_exporter) | 🔴 P0 | Moyenne |
| K8s API scraper — nodes, pods, namespaces (client-go) | 🔴 P0 | Haute |
| Docker scraper — containers standalone + réseaux | 🟡 P1 | Moyenne |
| Connection Resolver (K8s Services + Endpoints + Ingress) | 🔴 P0 | Haute |
| Layout Engine — positions déterministes (Grid 2D) | 🔴 P0 | Haute |
| Health endpoint réel (`/api/v1/health`) | 🟡 P1 | Faible |

### Frontend

| Tâche | Priorité | Complexité |
|-------|----------|-----------|
| ConnectionSVG.tsx — lignes SVG entre nœuds | 🟡 P1 | Moyenne |
| StatusBadge avec code couleur depuis design system | 🔴 P0 | Faible |
| Filtre namespace fonctionnel | 🟡 P1 | Faible |
| Skeleton loading states | 🟡 P1 | Faible |

### Tests

| Tâche |
|-------|
| Tests unitaires PrometheusScraper (mock Prometheus client) |
| Tests unitaires K8sScraper (mock client-go) |
| Tests unitaires DockerScraper (mock Docker client) |
| Test Layout Engine idempotence |
| Test Connection Resolver avec fixtures K8s |
| docker-compose.dev.yml fonctionnel |

### Condition de sortie ✅

> Je branche InfraMaps sur mon cluster (ou kind local) et je vois **mes pods réels** avec leurs métriques CPU/mémoire réelles et leurs connexions Services/Pods.

---

## Phase 2 — 🔄 Utilisabilité Quotidienne (P0 livrés le 2026-07-30)

**Objectif :** InfraMaps devient agréable à utiliser tous les jours.

**Livré :**
- Backend : proxy métriques historiques `/nodes/:id/metrics` (range queries PromQL, mock synthétique en dev), proxy Loki `/nodes/:id/logs` (résolution UID→nom de pod, `ERR_LOKI_DISABLED` si non configuré), WebSocket `/ws/alerts` (hub + watcher diff fired/resolved + replay de l'état aux nouveaux clients), config env complète, CORS configurable.
- Front : SidePanel à onglets Métriques (Recharts) | Logs (LogViewer, filtre niveau, auto-scroll) | Détails ; alertes temps réel via WS avec reconnexion backoff (1s→30s) et pulse CSS sur les nœuds en alerte.
- Packaging : chart Helm complet (RBAC moindre privilège, probes, non-root, limits, tag image = appVersion), `helm lint` + `template` verts.

**Restant :** breadcrumb (P1), search globale (P2), condition de sortie temporelle (2 semaines d'usage quotidien).

### Backend

| Tâche | Priorité |
|-------|----------|
| Loki proxy réel (logs on-demand par nodeId) | 🔴 P0 |
| WebSocket alertes (Alertmanager webhook → push WS) | 🔴 P0 |
| Config complète via env vars (toutes les vars de DEPLOYMENT.md) | 🔴 P0 |
| CORS configurable | 🟡 P1 |

### Frontend

| Tâche | Priorité |
|-------|----------|
| Side Panel métriques historiques (Recharts, proxy Prometheus) | 🔴 P0 |
| Side Panel logs (LogViewer, auto-scroll, filtre level) | 🔴 P0 |
| Pulse animation CSS sur les nœuds en alerte | 🔴 P0 |
| Breadcrumb navigation (root → cluster → node → pod) | 🟡 P1 |
| Filtre tag + statut dans la Toolbar | 🟡 P1 |
| Search globale (nom de pod, namespace) | 🟠 P2 |

### Packaging

| Tâche | Priorité |
|-------|----------|
| Helm chart fonctionnel (deploy namespace, RBAC, values.yaml) | 🔴 P0 |
| CI/CD GitHub Actions (test + build + push image) | 🟡 P1 |

### Condition de sortie ✅

> Je déploie InfraMaps sur mon cluster perso via `helm install` en moins de 10 minutes **et** je l'utilise au quotidien pendant 2 semaines sans friction notable.

---

## Phase 3 — Robustesse Production

**Estimé :** 3 semaines (après Phase 2 stable)

**Objectif :** Prêt à déployer chez des clients sans surveillance.

### Backend

| Tâche | Priorité |
|-------|----------|
| Graceful degradation (scraper down → données périmées affichées) | 🔴 P0 |
| Reconnexion automatique WebSocket (backoff exponentiel) | 🔴 P0 |
| Pagination API (`?page=N&limit=50`) pour clusters > 100 pods | 🔴 P0 |
| Mode sans K8s — Docker Compose uniquement | 🔴 P0 |
| Mode sans Docker — K8s uniquement | 🔴 P0 |
| Mode VM only — Prometheus targets uniquement | 🔴 P0 |
| Timeout configurable sur chaque scraper | 🟡 P1 |

### Frontend

| Tâche | Priorité |
|-------|----------|
| Lazy loading sous-arbres (cliquer pour charger les enfants) | 🔴 P0 |
| État erreur source (scraper down — données périmées signalées) | 🔴 P0 |
| État vide (cluster/island sans données) | 🔴 P0 |
| URL state (lien partageable vers un nœud : `?node=pod-api-1`) | 🟡 P1 |
| Keyboard navigation (Tab, Entrée, Échap) | 🟠 P2 |

### Condition de sortie ✅

> InfraMaps tourne en production (ou staging client) pendant **1 mois sans incident** — scraper K8s qui redémarre, Loki indisponible → InfraMaps continue d'afficher les données disponibles sans crash.

---

## Phase 4 — Vue 3D Maps [CONDITIONNEL]

**Estimé :** 4–6 semaines (après Phase 3 stable)

**Gate d'entrée — les deux conditions doivent être remplies :**
1. Phase 3 est stable depuis au moins 1 mois
2. La vue Grid 2D est utilisée régulièrement ET j'ai identifié un cas où la 3D apporterait une valeur que la 2D ne couvre pas

**Si les deux conditions ne sont pas remplies : ne pas lancer Phase 4.**

### Ce que ça apporte

- Effet "wow" pour les démos clients (différenciation commerciale)
- Mémorisation spatiale de l'architecture ("le gros cube rouge à gauche = cluster prod")
- Diagnostic visuel immédiat pour grandes infrastructures

### Périmètre

| Tâche | Priorité |
|-------|----------|
| React Three Fiber installation et setup de base | 🔴 P0 |
| Layout Engine extension — positions 3D (Y = niveau, XZ = hash) | 🔴 P0 |
| Rendu hiérarchique (plateau cluster, cubes nodes, pods = fenêtres) | 🔴 P0 |
| LOD (Level of Detail) — métriques visibles selon zoom | 🟡 P1 |
| Pulse animation Three.js sur nœuds en alerte | 🟡 P1 |
| Toggle Grid ↔ 3D sans rechargement des données | 🔴 P0 |
| Contrôles caméra (orbit, zoom, pan) | 🔴 P0 |
| Performance (> 200 pods sans frame drop) | 🟡 P1 |

**Règle de régression :** La vue Grid 2D doit rester fonctionnelle. Aucune fonctionnalité de la Phase 3 ne peut régresser.

---

## Phase 5 — Actions Production [PRUDENCE]

> **Cette phase n'a pas de timeline.** Elle sera décidée après usage réel en production ET analyse sécurité.

### Ce qui est étudié (pas décidé)

| Action | Complexité | Risque | Décision |
|--------|-----------|--------|----------|
| Restart pod | Moyenne | Élevé (K8s write access) | À évaluer |
| Scale deployment | Haute | Très élevé | À évaluer |
| Exec shell | — | **Inacceptable** | ❌ Hors scope permanent |

### Prérequis si Phase 5 lancée

- Audit log de chaque action (qui, quand, quoi)
- RBAC granulaire (qui peut restart quoi)
- Confirmation UI avec délai (bouton rouge + "Confirmez dans 5s")
- Tests E2E complets
- Revue sécurité externe

---

## Skills Map par Phase

> Corrigé (2026-07-29) : la liste précédente référençait des skills qui n'existent pas dans cette installation (`go-api-builder`, `helm-chart-builder`, `frontend-design`, `tech-doc-writing`). Voici les skills réellement disponibles et pertinents.

| Phase | Skill disponible | Usage |
|-------|-------------------|-------|
| **Phase 0.5** | `cicd-pipeline-builder` | Pipeline GitHub Actions, release automation — déjà utilisé pour poser les fondations ci-dessus |
| **Phase 0.5 / 2** | `docker-builder` | Dockerfile multi-stage `infra-maps-api`, image slim non-root |
| **Phase 1 / 3** | `sre-observability-expert` | Monitorer InfraMaps lui-même (il consomme Prometheus, il doit aussi être observable) |
| **Phase 2** | `grafana-dashboard-builder` | Si un dashboard Grafana est utile pour suivre la santé d'InfraMaps en prod |
| **Toutes** | `run` | Lancer et vérifier l'app dans un vrai navigateur avant de clore une tâche front |
| **Toutes** | `simplify` | Passe de nettoyage après une feature, avant merge |
