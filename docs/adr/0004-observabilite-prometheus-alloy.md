# ADR 004 — Observabilité : exporter Prometheus natif + réutilisation d'Alloy/cAdvisor

## Contexte

L'API exposait déjà une route `/metrics`, mais basée sur le monitor intégré de
`gofiber/contrib` : elle renvoie du JSON pensé pour un dashboard fiber interne,
pas un format exploitable par un scraper Prometheus.

Le VPS fait par ailleurs déjà tourner une stack d'observabilité complète pour
d'autres projets : Grafana Alloy (avec un `prometheus.exporter.unix` pour
l'hôte) et cAdvisor, qui remontent vers Grafana Cloud via
`prometheus.remote_write`. Cette stack applique déjà un filtrage (`keep`)
serré sur les métriques cAdvisor pour rester sous la limite de débit de
Grafana Cloud (75 req/s). Deux options se présentaient : déployer une stack
Prometheus/Grafana dédiée à ce projet, ou brancher l'API sur l'infrastructure
d'observabilité déjà en place.

## Décision

J'utilise `prometheus/client_golang` + `promhttp.Handler()` (monté via le
middleware `adaptor` de fiber v3) à la place du monitor JSON, avec deux
métriques applicatives (`http_requests_total`, `http_request_duration_seconds`)
en plus des métriques runtime Go par défaut. Côté infra, pas de nouvelle stack :
le conteneur `alloy` existant est rattaché à `public_network` pour résoudre
`open_dpp_api` par son nom, avec un bloc `prometheus.scrape` dédié dans
`config.alloy` et son propre `prometheus.relabel` qui ne garde que les
métriques utiles (même logique de filtrage que cAdvisor).

## Conséquences

+ Aucune nouvelle stack à maintenir, sécuriser ou monitorer elle-même :
  l'observabilité de l'API bénéficie directement du travail déjà fait sur
  Alloy (durcissement, remote_write, gestion du quota Grafana Cloud).
+ Le dashboard Grafana ([`observability/dashboard-vps.json`](../../observability/dashboard-vps.json))
  est versionné en JSON importable plutôt que construit à la main dans l'UI.
+ Format Prometheus standard : compatible avec n'importe quel outil qui sait
  scraper du texte Prometheus, pas seulement Alloy.
- Toute nouvelle métrique applicative doit aussi être ajoutée à la regex
  `keep` de `prometheus.relabel.open_dpp_api` dans `config.alloy`, sinon elle
  est scrapée par Alloy mais jamais transmise à Grafana Cloud — dette à
  surveiller à chaque évolution de l'API.
- Le rattachement d'`alloy` à `public_network` élargit sa surface réseau (il
  peut désormais résoudre tout service présent sur ce réseau partagé).
