# Runbook : API `open-dpp-api` down

## Détection

- Healthcheck Docker rouge : `docker ps` affiche `open_dpp_api` en
  `(unhealthy)` ou le conteneur n'apparaît plus du tout.
- `curl -sf https://api.jules.ruberti.mds-nantes.fr/health` ne répond pas
  `200 OK` (timeout, 502/503 via Nginx Proxy Manager, ou connexion refusée).
- Panneau **"Requêtes / s par route et code HTTP"** ou **"Taux d'erreurs 5xx"**
  du dashboard Grafana (`observability/dashboard-vps.json`) qui chute à zéro ou
  monte en erreurs.

## Diagnostic

Sur le VPS (`ssh VPS_MDS`) :

1. `docker ps -a --filter name=open_dpp_api` → le conteneur est-il arrêté,
   en boucle de redémarrage, ou juste `unhealthy` mais toujours `Up` ?
2. `docker logs --tail 100 open_dpp_api` → erreur applicative au démarrage
   (mauvais port, panic Go) ou simplement lent à répondre ?
3. `docker inspect open_dpp_api --format '{{json .State.Health}}'` → historique
   des derniers checks de healthcheck et leur sortie exacte.
4. `docker stats open_dpp_api --no-stream` → le conteneur est-il en limite de
   CPU/RAM (auquel cas le dashboard Grafana section "Conteneurs" le confirmera
   aussi) ?
5. Si le conteneur tourne mais ne répond pas : `docker exec -it open_dpp_api sh`
   puis `wget -qO- http://127.0.0.1:7000/health` depuis l'intérieur, pour
   distinguer un problème réseau (Nginx Proxy Manager / `public_network`) d'un
   problème applicatif.

## Mitigation

- **Conteneur crashé / boucle de redémarrage** → lire la cause dans les logs
  (étape 2), corriger dans le code, puis redéployer via un nouveau push sur
  `main` (le pipeline CI/CD rebuild et redéploie automatiquement).
- **Dernier déploiement fautif, besoin d'un retour arrière rapide** → suivre
  la procédure de [rollback](../README.md#rollback) vers le tag
  `${{ github.sha }}` précédent, connu comme fonctionnel.
- **Conteneur simplement arrêté (pas de crash)** →
  `docker compose -f compose.prod.yml up -d` depuis `~/iter-open-api` sur le
  VPS pour le relancer.
- **Ressources saturées (CPU/RAM)** → vérifier les autres conteneurs du VPS
  via le dashboard Grafana (section "Conteneurs (cAdvisor)") pour identifier
  un voisin bruyant, avant d'envisager d'augmenter les ressources du VPS.
