# Changelog

## [0.5.0](https://github.com/AristideWafo/infra-map/compare/infra-maps-api-v0.4.0...infra-maps-api-v0.5.0) (2026-07-30)


### Features

* **infra-maps-api:** per-scraper timeout override ([#20](https://github.com/AristideWafo/infra-map/issues/20)) ([bba50ec](https://github.com/AristideWafo/infra-map/commit/bba50ec6c8888b6bbcb35875cb599f9250c76cbe))
* **infra-maps-api:** resolve Ingress → Service connections ([#22](https://github.com/AristideWafo/infra-map/issues/22)) ([90b8521](https://github.com/AristideWafo/infra-map/commit/90b8521f992f730d6fb00e209870fe7a84b201c9))

## [0.4.0](https://github.com/AristideWafo/infra-map/compare/infra-maps-api-v0.3.0...infra-maps-api-v0.4.0) (2026-07-30)


### Features

* multi-mode scraper tests + stale-data visual signal on 200 responses ([#13](https://github.com/AristideWafo/infra-map/issues/13)) ([e38409f](https://github.com/AristideWafo/infra-map/commit/e38409f40d5bdbc2865e3bbab7d42abf6fbf5124))

## [0.3.0](https://github.com/AristideWafo/infra-map/compare/infra-maps-api-v0.2.0...infra-maps-api-v0.3.0) (2026-07-30)


### Features

* graceful degradation — serve stale tree, scraper timeout env, stale banner ([4988562](https://github.com/AristideWafo/infra-map/commit/49885627d7d180824b861f31436168e3c6b7fb9c))
* **infra-maps-api:** enrich K8s pods with CPU/memory from Prometheus cadvisor ([#8](https://github.com/AristideWafo/infra-map/issues/8)) ([a5644a4](https://github.com/AristideWafo/infra-map/commit/a5644a4cb7b23d56877ba2acf2b23b818e1a0343))
* **infra-maps-api:** pagination on /connections per contract conventions ([#9](https://github.com/AristideWafo/infra-map/issues/9)) ([2a9c0a7](https://github.com/AristideWafo/infra-map/commit/2a9c0a73392d7637ad53162c9f65769c1c82c03f))
* **infra-maps-ui:** side panel tabs (metrics/logs/details), WS alerts with pulse ([cacc59d](https://github.com/AristideWafo/infra-map/commit/cacc59d23e754657cffa7daaacb29e5f0b17a1aa))

## [0.2.0](https://github.com/AristideWafo/infra-map/compare/infra-maps-api-v0.1.0...infra-maps-api-v0.2.0) (2026-07-30)


### Features

* **infra-maps-api:** connection resolver — Service to Pod topology via Endpoints ([03c988c](https://github.com/AristideWafo/infra-map/commit/03c988cb30d027e99bc6616f396bd83357918886))
* **infra-maps-api:** Docker scraper — standalone containers + compose groups ([f1eb1ae](https://github.com/AristideWafo/infra-map/commit/f1eb1aef92bf411fdcd36e23d655bbd31d75c7f4))
* **infra-maps-api:** Kubernetes scraper — cluster/nodes/pods topology via client-go ([df3e41f](https://github.com/AristideWafo/infra-map/commit/df3e41f220fd5c05288e2005382213c2645b6cfc))
* **infra-maps-api:** metrics history proxy + Loki logs proxy ([5049975](https://github.com/AristideWafo/infra-map/commit/5049975ee5d7fd608d9a0ca073fef7ec76bd934e))
* **infra-maps-api:** phase 0 — mock Prometheus scraper, deterministic layout, tree API ([c3b1752](https://github.com/AristideWafo/infra-map/commit/c3b1752b8712e992bd2a8353635efede7142496e))
* **infra-maps-api:** real Prometheus scraper — VM discovery via targets + PromQL metrics ([de2d1c8](https://github.com/AristideWafo/infra-map/commit/de2d1c8d3eb1abc4cd87bff3c961236e575fc304))
* **infra-maps-api:** WebSocket alerts — hub, watcher, /ws/alerts route ([5048cd8](https://github.com/AristideWafo/infra-map/commit/5048cd8111b766c992a392a21c0b8514b131bf50))
