# ADR-002 — Pas de Base de Données

**Date :** 2026-07-21  
**Statut :** Accepté  
**Décideur :** Aristide Wafo

---

## Contexte

InfraMaps doit être déployable chez des clients en 10 minutes via `helm install`. Chaque composant supplémentaire est un frein au déploiement et une charge de maintenance pour le client.

La question : faut-il une base de données pour stocker l'état de l'infrastructure ?

---

## Décision

**InfraMaps n'a pas de base de données. Jamais.**

- **Cache en mémoire (Go map + TTL 30s)** pour les données de visualisation courantes.
- **Prometheus** est la base de données des métriques historiques — elle est déjà en place chez tous les clients utilisant InfraMaps.
- **localStorage du navigateur** pour les préférences utilisateur (zoom, nœud sélectionné, mode Grid/3D).

---

## Conséquences

**Positives :**
- Déploiement en un seul composant : juste le pod `infra-maps`
- Pas de backup à configurer, pas de migration de schéma, pas de tuning
- Redémarrage propre : au pire 30 secondes sans données (le temps d'un cycle de scraping)
- Empreinte mémoire faible : ~50-100MB pour le cache Go

**Négatives :**
- Redémarrage = perte du cache (30s sans données)
- Pas d'historique InfraMaps propre (dépendance Prometheus pour l'historique)
- Pas de HA possible sans Redis (hors scope)
- Les préférences utilisateur sont par navigateur, pas partagées

**Neutres :**
- L'historique des métriques (CPU, RAM sur 1h, 24h) est correctement couvert par Prometheus
- Le TTL de 30s correspond exactement à l'intervalle de scraping

---

## Alternatives Considérées

### Alternative A — SQLite embarqué

Base de données légère, sans serveur, dans le binaire Go.

**Rejeté parce que :** Ajoute CGO (compilation plus complexe), poids dans l'image Docker, migrations de schéma, backup. Pour quel gain ? L'historique est déjà dans Prometheus. Le cache en mémoire suffit pour les 30s de présent.

### Alternative B — Redis comme cache externe

Cache distribué, permettrait la HA (plusieurs réplicas).

**Rejeté pour le MVP parce que :** HA n'est pas un requirement. Un seul pod InfraMaps suffit. Redis = composant supplémentaire à déployer, configurer, maintenir. Violation du principe KISS.

### Alternative C — etcd ou BoltDB

Base de données clé-valeur embarquée.

**Rejeté pour les mêmes raisons que SQLite.** L'historique n'est pas un besoin que Prometheus ne couvre pas déjà.
