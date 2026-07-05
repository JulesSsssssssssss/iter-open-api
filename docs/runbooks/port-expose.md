# Runbook : Port exposé publiquement par erreur

## Détection

- Scan nmap externe périodique (`nmap -Pn <ip-vps>` depuis une machine hors du
  VPS) qui révèle un port ouvert non attendu (seuls 22, 80, 443 doivent
  répondre).
- Alerte manuelle : un service admin (NPM, SonarQube, Grafana Alloy debug UI
  sur `12345`...) accessible directement par IP:port depuis l'extérieur.

## Diagnostic

Sur le VPS (`ssh VPS_MDS`) :

1. `sudo ufw status numbered` → le port apparaît-il dans les règles UFW
   `ALLOW` ? S'il n'y est pas alors qu'il répond quand même de l'extérieur,
   c'est le symptôme du piège Docker + UFW (voir
   [ADR 002](../adr/0002-hardening-npm-port-admin.md)).
2. `docker ps --format '{{.Names}}\t{{.Ports}}'` → identifier quel conteneur
   publie ce port, et sur quelle interface (`0.0.0.0:PORT->...` = public,
   `127.0.0.1:PORT->...` = local uniquement).
3. `docker inspect <container> --format '{{json .NetworkSettings.Ports}}'`
   pour confirmer le binding exact.

## Mitigation

- **Le port n'a pas besoin d'être public** (cas le plus fréquent : interface
  d'admin, UI de debug) → éditer le `compose.yml` du service concerné pour
  rebinder sur loopback :

  ```yaml
  ports:
    - "127.0.0.1:PORT:PORT"   # au lieu de "PORT:PORT"
  ```

  puis `docker compose up -d` pour recréer le conteneur avec le nouveau
  binding. Ne pas essayer de corriger uniquement via une règle UFW
  supplémentaire : elle n'a aucun effet sur un port publié par Docker.
- **Le port doit être accessible mais seulement pour un usage précis** (ex.
  accès admin depuis un poste fixe) → passer par le reverse proxy (Nginx
  Proxy Manager) avec restriction d'accès (liste d'IP, HTTPS + auth), plutôt
  que de publier le port directement.
- **Vérification finale** → relancer un scan nmap externe et confirmer que
  seuls 22, 80, 443 (et éventuellement un port SSH custom) répondent.
- Si le service exposé avait un accès non authentifié (cas NPM), réinitialiser
  son mot de passe admin par précaution.
