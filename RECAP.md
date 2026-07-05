# Récap — Axes 1, 2 et 3 (Docker, CI/CD, Sécurité VPS)

## Axe 1 — Dockerfiles & Compose

**Bugs corrigés :**
- `compose.yml` : mot de passe PostgreSQL en clair (`***REMOVED-LEAKED-SECRET***`) → déplacé vers `.env` via `env_file`.
- `compose.yml` : le service `open-dpp-db` n'avait aucune clé `networks:`, donc n'était joignable ni par l'API ni par personne — ajouté sur `internal_network`.
- `open-api/Dockerfile` : images non pinnées (`golang:alpine`, `alpine`) → `golang:1.26.3-alpine3.23` et `alpine:3.23.5`.
- `db/init/check.sh` : syntaxe bash invalide (`if` non fermé, fonction vide) + pas exécutable → empêchait la DB de démarrer `healthy`. Réécrit et `chmod +x`.
- `.gitignore` : ajout de `.env` pour ne jamais committer de secret réel.

**Validé** avec un vrai `docker compose up --build` en local : DB `healthy`, API répond 200 sur `/health` et `/api`, DB injoignable depuis l'extérieur du réseau interne.

## Axe 2 — Pipeline CI/CD (`.github/workflows/go.yml`)

**Ajouté :**
- Job `lint` (golangci-lint).
- Vrais tests unitaires (`main_test.go`, refactor `setupApp()` dans `main.go` pour les rendre testables — avant, `go test` ne testait rien).
- Tags d'image `latest` + `${{ github.sha }}` au lieu de `latest` seul.
- Job `trivy-scan` bloquant sur vulnérabilités `CRITICAL`.
- Job `deploy` : copie `compose.prod.yml` sur le VPS, login GHCR éphémère (token du run, rien de stocké côté serveur), `pull` + `up -d`.

**Bugs rencontrés et corrigés en cours de route** (tous vérifiés en conditions réelles sur GitHub Actions) :
1. `golangci-lint-action@v6` retélécharge un vieux binaire (go1.24) incompatible avec go1.26.3 → installation manuelle du binaire officiel, pinnée à `v2.12.2`.
2. Le script d'install `master` de golangci-lint échoue en vérification de checksum contre un release v2 → script pinné au même tag que le binaire.
3. `aquasecurity/trivy-action@0.28.0` → tag inexistant (il manquait le `v`) → `v0.36.0`.
4. `compose.prod.yml` référençait une image d'un autre dépôt (`ghcr.io/marguillat/...`) → corrigé vers `ghcr.io/julesssssssssssss/iter-open-api:latest` (avec une coquille de comptage de caractères corrigée au passage : 15 vs 17 `s`).
5. Healthcheck `wget http://localhost:7000/health` échouait car `localhost` résout d'abord en IPv6 (`::1`) dans le conteneur, alors que l'API n'écoute qu'en IPv4 → remplacé par `127.0.0.1`.

**Résultat :** pipeline complet vert — Lint → Tests → Snyk → SonarQube → Build/Push GHCR → Trivy → Deploy VPS. Conteneurs vérifiés `healthy` et fonctionnels sur le VPS après déploiement réel.

**Secrets GitHub mis en place :** `SSH_HOST`, `SSH_USER`, `SSH_PORT`, `SSH_PRIVATE_KEY`, `SNYK_TOKEN`, `SONAR_TOKEN`, `SONAR_HOST_URL`.

## Axe 3 — Sécurité VPS

**Déjà en place (vérifié, rien à changer) :**
- SSH durci : `PermitRootLogin no`, `PasswordAuthentication no`, clé uniquement.
- `fail2ban` actif avec une jail `sshd`.
- UFW actif, politique par défaut *deny incoming*.

**Fait :**
- SonarQube (déjà présent sur le VPS mais jamais démarré) : lancé, mot de passe admin réinitialisé, token API généré pour la CI.
- Nginx Proxy Manager (déjà en place pour d'autres services) : ajout de 2 nouveaux hosts avec certificats Let's Encrypt automatiques :
  - `api.jules.ruberti.mds-nantes.fr` → `open_dpp_api:7000`
  - `swagger.jules.ruberti.mds-nantes.fr` → `iter_swagger_ui:8080`
  - HTTPS forcé, HSTS, HTTP/2, headers de sécurité (`X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy`), redirection HTTP → HTTPS.
- **Bug de sécurité trouvé et corrigé :** le port `81` (panel admin NPM) était exposé publiquement dans UFW. Supprimer la règle UFW ne suffisait pas — Docker gère iptables directement et contourne les règles UFW pour les ports qu'il publie (piège classique Docker+UFW). Fix réel : rebind du port dans le compose de NPM sur `127.0.0.1:81:81` au lieu de `0.0.0.0:81`. L'admin NPM reste accessible en HTTPS via `proxy.jules.ruberti.mds-nantes.fr`.
- Mot de passe admin NPM réinitialisé au passage (les proxy hosts/certificats existants — vault, portfolio, sonarqube, claude — n'ont pas été affectés, stockés dans des tables séparées).

**Validé par scan nmap externe** (1000+ ports scannés) : seuls 22, 80, 443 sont ouverts. DB, port admin NPM, et tout le reste sont invisibles depuis l'extérieur.

## Axe 4 — Observabilité

**Constat de départ :** `/metrics` existait déjà mais renvoyait le JSON interne du monitor `gofiber/contrib` (dashboard UI de fiber), pas un format exploitable par Prometheus/Alloy.

**Fait :**
- `open-api/main.go` : remplacement par `prometheus/client_golang` + `promhttp.Handler()` (monté via le middleware `adaptor` de fiber v3) → `/metrics` répond maintenant en texte Prometheus.
- Ajout de métriques applicatives : `http_requests_total{method,path,status}` (compteur) et `http_request_duration_seconds{method,path}` (histogramme), plus les métriques runtime Go par défaut (`go_goroutines`, `process_cpu_seconds_total`, `process_resident_memory_bytes`…).
- Réutilisation de la stack déjà en place sur le VPS (`/home/ubuntu/apps/grafana`, Alloy + cAdvisor → Grafana Cloud) plutôt que d'en reconstruire une :
  - `compose.yml` de la stack Alloy : ajout de `public_network` au service `alloy` (en plus de son réseau par défaut) pour qu'il puisse résoudre `open_dpp_api` par son nom de conteneur.
  - `config.alloy` : nouveau bloc `prometheus.scrape "open_dpp_api"` (cible `open_dpp_api:7000`, intervalle 30s) + `prometheus.relabel "open_dpp_api"` qui ne garde que les métriques utiles (`http_requests_total`, `http_request_duration_seconds_*`, `process_cpu_seconds_total`, `process_resident_memory_bytes`, `go_goroutines`) pour rester sous la limite de débit Grafana Cloud (75 req/s), même logique que le filtrage déjà appliqué à cAdvisor.
- **Validé en conditions réelles** : après redéploiement de l'API (pipeline CI/CD) et redémarrage d'Alloy, le composant `prometheus.scrape.open_dpp_api` de l'API de debug d'Alloy (`:12345/api/v0/web/components`) est passé de `health: down` (l'ancien `/metrics` renvoyait du JSON, imparsable par Alloy) à `health: up`.
- Dashboard Grafana livré en JSON prêt à importer : [`observability/dashboard-vps.json`](observability/dashboard-vps.json) — reprend les panneaux CPU/RAM/réseau/disque par conteneur du TP (variable `$container`, cartouches de synthèse) et ajoute une section API (requêtes/s par route et code HTTP, latence p50/p95/p99, taux d'erreurs 5xx, CPU/RAM/goroutines du process). À l'import, Grafana demande de mapper la datasource `${DS_PROMETHEUS}` vers le Prometheus Grafana Cloud existant.

## Ce qu'il reste (axe 5, + bonus axe 3)

- **Axe 5 — Documentation & ADR** : pas commencé. Le README racine est vide, aucun ADR, aucun runbook.
- **Bonus axe 3** : audit Mozilla Observatory, CSP header, rotation des secrets documentée, backup off-site.