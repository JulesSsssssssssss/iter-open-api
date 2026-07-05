# Runbook : Pipeline CI/CD en échec

## Détection

- Badge/statut rouge sur GitHub Actions (`gh run list --repo <org>/iter-open-api`
  ou onglet **Actions** du repo).
- Un push sur `main` ne se retrouve pas déployé sur le VPS après quelques
  minutes alors qu'il n'y a pas eu d'erreur visible côté développeur.

## Diagnostic

1. `gh run list --repo <org>/iter-open-api --limit 5` → quel run a échoué,
   sur quel commit ?
2. `gh run view <run-id> --repo <org>/iter-open-api` → quel **job** a échoué
   (`lint`, `security`, `tests`, `quality`, `build-and-push`, `trivy-scan`,
   `deploy`) ?
3. Selon le job en échec :
   - `lint` → `golangci-lint` a trouvé une erreur de style/qualité : lire le
     détail dans les logs du job, corriger dans le code.
   - `tests` → un test unitaire (`main_test.go`) échoue : reproduire en local
     avec `cd open-api && go vet ./... && go test -v ./...`.
   - `security` (Snyk) ou `quality` (SonarQube) → problème de configuration
     du token (`SNYK_TOKEN`, `SONAR_TOKEN`, `SONAR_HOST_URL` dans les secrets
     GitHub) ou vulnérabilité/dette technique réelle à traiter.
   - `trivy-scan` → une vulnérabilité `CRITICAL` a été trouvée dans l'image :
     lire le rapport dans les logs du job pour identifier le paquet concerné.
   - `deploy` → généralement un problème de connexion SSH (secrets `SSH_HOST`,
     `SSH_USER`, `SSH_PORT`, `SSH_PRIVATE_KEY`) ou le VPS est injoignable.

## Mitigation

- **`lint` / `tests`** → corriger le code en local, vérifier avec les mêmes
  commandes que la CI avant de repush.
- **`trivy-scan` (CRITICAL)** → mettre à jour la dépendance ou l'image de base
  concernée (`go.mod`/`go.sum` via `go get -u`, ou tag `golang`/`alpine` plus
  récent dans le `Dockerfile`), puis repush. Le déploiement reste bloqué tant
  que la CVE n'est pas corrigée (comportement voulu, voir
  [ADR 003](../adr/0003-pipeline-cicd-versions-pinnees.md)).
- **`deploy` (connexion SSH)** → vérifier que le VPS répond
  (`ssh VPS_MDS "echo ok"`), que les secrets GitHub n'ont pas expiré/changé, et
  que le port SSH custom (secret `SSH_PORT`) est toujours correct.
- **Dans tous les cas** → une fois le correctif poussé, le pipeline repart
  automatiquement sur `main` ; pas de ré-exécution manuelle nécessaire sauf cas
  de flakiness ponctuelle (`gh run rerun <run-id>`).
