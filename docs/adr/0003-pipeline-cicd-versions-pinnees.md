# ADR 003 — Versions pinnées et scan bloquant dans le pipeline CI/CD

## Contexte

Le pipeline (`.github/workflows/go.yml`) build et déploie automatiquement sur
le VPS à chaque push sur `main`. Deux catégories de problèmes sont apparues
pendant sa mise en place :

1. **Dérive de version non maîtrisée.** `golangci-lint-action@v6` retélécharge
   un binaire figé sur go1.24, incompatible avec le go1.26.3 du projet. Le
   script d'installation `master` de golangci-lint échoue ensuite en
   vérification de checksum contre une release v2, car `master` et le binaire
   ne sont plus alignés. `aquasecurity/trivy-action@0.28.0` référence un tag
   qui n'existe pas (`v` manquant).
2. **Risque de déployer une image vulnérable.** Rien n'empêchait de pousser en
   prod une image contenant une CVE critique connue.

## Décision

J'utilise des versions pinnées à un tag exact partout où une dérive silencieuse
est possible, et un scan de vulnérabilités bloquant avant tout déploiement :

- `golangci-lint` installé depuis son script officiel pinné à `v2.12.2` (même
  tag que le binaire téléchargé, pour que la vérification de checksum
  réussisse) ; `aquasecurity/trivy-action` pinné à `v0.36.0`.
- Images Docker de base pinnées à une version précise
  (`golang:1.26.3-alpine3.23`, `alpine:3.23.5`) plutôt qu'à un tag flottant
  (`alpine`, `golang:alpine`).
- Job `trivy-scan` **bloquant** (`exit-code: '1'`) sur toute vulnérabilité
  `CRITICAL`, exécuté après le build de l'image et avant le déploiement — le
  job `deploy` en dépend (`needs:`) et ne s'exécute pas si Trivy échoue.
- Images poussées sur GHCR taguées à la fois `latest` et
  `${{ github.sha }}`, pour pouvoir identifier et rollback vers une version
  précise (voir [runbook pipeline en échec](../runbooks/pipeline-echec.md)).

## Conséquences

+ Pipeline reproductible : un `go.mod` à go1.26.3 fonctionne avec un lint qui
  comprend cette version, sans dépendre de ce que l'action tierce décide de
  retélécharger un jour donné.
+ Une vulnérabilité CRITICAL détectée par Trivy bloque le déploiement — le
  correctif doit être appliqué avant toute mise en prod, pas après.
+ Chaque image en prod est traçable jusqu'au commit exact via le tag SHA,
  ce qui rend le rollback possible.
- Chaque montée de version majeure (Go, golangci-lint, Trivy) est un
  changement explicite à faire dans le workflow, pas automatique.
