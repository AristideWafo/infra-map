# ADR-001 — Architecture Multi-Source Optionnelle

**Date :** 2026-07-21  
**Statut :** Accepté  
**Décideur :** Aristide Wafo

---

## Contexte

InfraMaps visualise l'infrastructure. L'infrastructure des clients est hétérogène :
- Certains clients ont K8s + Docker + Prometheus + Loki (stack complète)
- D'autres ont uniquement des VMs avec Prometheus/node_exporter, sans containers
- D'autres ont Docker Compose mais pas K8s
- La majorité a une combinaison de tout ça

La documentation initiale assumait implicitement que K8s était toujours présent. Ce n'est pas le cas.

---

## Décision

**Prometheus est la seule source obligatoire. Toutes les autres sources (K8s API, Docker socket, Loki) sont optionnelles et indépendamment activables via configuration.**

Le backend démarre avec uniquement Prometheus. Si K8s est activé ET accessible, il scrape la topologie K8s. Si Docker est activé ET accessible, il scrape les containers Docker. Si Loki est configuré, il proxifie les logs.

**Découverte des VMs sans K8s :** L'API Prometheus `/api/v1/targets` liste tous les hosts scrapés. Les targets avec `job="node"` correspondent aux VMs avec `node_exporter`. Leurs labels Prometheus permettent de les grouper en îlots logiques.

**Comportement en cas de source indisponible :** Le backend ne crash pas. Il affiche les données des sources disponibles, marque les autres comme `disabled` dans le health endpoint, et continue à fonctionner en mode dégradé.

---

## Conséquences

**Positives :**
- InfraMaps déployable sur n'importe quelle infrastructure avec Prometheus
- Pas de dépendance à K8s pour les clients Docker-only ou VM-only
- Mode dégradé gracieux : perte d'une source ≠ perte du service
- Moins de prérequis à expliquer aux clients

**Négatives :**
- Plus de code de configuration à gérer
- Tests plus complexes (toutes les combinaisons de sources)
- La topologie est incomplète sans K8s (pas de relations pod/service)

**Neutres :**
- Le Scraper interface pattern absorbe naturellement cette complexité
- Un scraper `Noop` satisfait l'interface pour les sources désactivées

---

## Alternatives Considérées

### Alternative A — K8s obligatoire

Tous les clients doivent avoir K8s. InfraMaps devient un outil K8s-only.

**Rejeté parce que :** Trop restrictif. Un grand nombre de clients ont des VMs ou du Docker Compose sans K8s. Ça exclut une partie importante du marché cible (SOC, PME, infra legacy).

### Alternative B — Agents propriétaires

Déployer un agent InfraMaps sur chaque host pour la découverte.

**Rejeté parce que :** Violation du principe KISS fondamental du projet. Ajouter un agent = ajouter de la maintenance chez chaque client. L'avantage clé d'InfraMaps est de n'utiliser que l'infrastructure déjà en place.

### Alternative C — Service mesh obligatoire

Requérir Istio ou Linkerd pour la découverte des connexions réseau.

**Rejeté parce que :** Très peu de clients ont un service mesh. Ce serait une dépendance bloquante pour la majorité des cas d'usage.
