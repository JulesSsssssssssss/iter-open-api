# Alertes Grafana Cloud

L'API et les conteneurs remontent déjà vers Grafana Cloud (voir
[ADR 004](adr/0004-observabilite-prometheus-alloy.md)), mais aucune règle
d'alerte n'existe encore par défaut : Grafana ne fait qu'afficher les
métriques tant qu'on ne configure pas explicitement des règles. Ce document
donne les 3 alertes minimales à créer (une fois, ~5 min), avec la requête
PromQL et le seuil exacts.

## Pourquoi 3 alertes précisément

- **Service down** : le signal le plus basique et le plus actionnable —
  personne ne veut apprendre une panne par un utilisateur.
- **Taux d'erreur élevé** : détecte une régression applicative (bug déployé)
  avant qu'elle ne devienne une panne complète.
- **Disque plein** : cause silencieuse d'incident en cascade (DB qui ne peut
  plus écrire, logs qui remplissent le disque) — invisible sans alerte
  dédiée, contrairement à un crash immédiat.

## 1. Service API down

- **Menu** : ☰ → Alerting → Alert rules → **New alert rule**.
- **Requête** (datasource Prometheus Grafana Cloud) :
  ```promql
  up{job="open-dpp-api"} == 0
  ```
- **Condition** : `IS BELOW 1` (ou `WHEN last() IS BELOW 1`), évaluée toutes
  les `1m`, **for** `1m` avant de déclencher (évite les faux positifs sur un
  redémarrage normal de quelques secondes).
- **Nom** : `APIDown`. **Résumé** : "L'API open-dpp-api ne répond plus depuis
  1 minute." **Lien runbook** : coller l'URL du
  [runbook API down](runbooks/api-down.md) dans le champ "Runbook URL" de
  l'annotation.

## 2. Taux d'erreur élevé (5xx)

- **Requête** :
  ```promql
  sum(rate(http_requests_total{job="open-dpp-api", status=~"5.."}[5m]))
  /
  sum(rate(http_requests_total{job="open-dpp-api"}[5m]))
  ```
- **Condition** : `IS ABOVE 0.05` (plus de 5 % de réponses 5xx sur 5 minutes),
  **for** `5m`.
- **Nom** : `APIHighErrorRate`. **Runbook** :
  [pipeline en échec](runbooks/pipeline-echec.md) si l'erreur vient d'un
  déploiement récent, sinon investiguer les logs applicatifs directement.

## 3. Disque plein sur le VPS

- **Requête** (métriques `node_exporter` exposées par Alloy) :
  ```promql
  node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}
  ```
- **Condition** : `IS BELOW 0.1` (moins de 10 % d'espace disque libre sur `/`),
  **for** `10m`.
- **Nom** : `DiskAlmostFull`. **Résumé** : "Moins de 10 % d'espace disque
  libre sur le VPS."

## Contact point

Pour un rendu individuel sans webhook Slack/Discord déjà en place, le plus
rapide est le contact point **Email** intégré à Grafana Cloud (aucune
configuration supplémentaire) : ☰ → Alerting → Contact points → vérifier
qu'un contact point email existe et est bien assigné à la policy de
notification par défaut. Une intégration Discord/Slack peut être ajoutée
plus tard sans changer les règles d'alerte elles-mêmes.

## Vérifier qu'une alerte fonctionne

Le plus simple : couper temporairement l'API (`docker stop open_dpp_api` sur
le VPS) et observer que `APIDown` passe à l'état **Firing** dans Grafana en
moins de 2 minutes, puis revient à **Normal** après `docker start
open_dpp_api`.
