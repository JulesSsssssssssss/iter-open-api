# ADR 001 — Segmentation réseau Docker (interne vs public)

## Contexte

La stack comprend trois services : la base PostgreSQL (`open-dpp-db`), l'API Go
(`open-dpp-api`) et le Swagger UI. Seuls l'API et le Swagger UI doivent être
joignables depuis l'extérieur (via Nginx Proxy Manager) ; la base de données ne
doit jamais être accessible autrement que par l'API.

Dans une première version de `compose.yml`, le service `open-dpp-db` n'avait
aucune clé `networks:` définie, ce qui le plaçait sur le réseau par défaut du
projet — techniquement joignable par l'API, mais sans qu'aucune isolation ne
soit explicite ni garantie dans le temps (un service ajouté par erreur sur ce
même réseau par défaut aurait pu atteindre la DB).

## Décision

J'utilise deux réseaux Docker distincts, déclarés explicitement pour chaque
service, plutôt qu'un réseau par défaut partagé :

- `internal_network` (bridge interne) : `open-dpp-db` et `open-dpp-api`. C'est
  le seul réseau sur lequel la base écoute.
- `public_network` (partagé avec Nginx Proxy Manager) : `open-dpp-api` et
  `swagger-ui` uniquement. La base n'y est jamais attachée.

L'API sert ainsi de frontière explicite entre le réseau public et la base de
données.

## Conséquences

+ La DB est injoignable depuis l'extérieur du réseau interne, y compris par
  d'autres conteneurs du VPS (vérifié en local : une connexion directe à
  `open-dpp-db:5432` échoue depuis un conteneur placé sur `public_network`).
+ Isolation garantie dans le temps : tout nouveau service doit choisir
  explicitement son réseau, pas d'attachement implicite qui romprait
  l'isolation.
+ Le réseau `public_network` a pu être réutilisé tel quel pour l'observabilité
  (voir [ADR 004](0004-observabilite-prometheus-alloy.md)), sans exposer la DB.
- Deux réseaux à déclarer et à retenir pour chaque nouveau service ajouté à la
  stack, contre un seul réseau par défaut si tout avait été laissé implicite.
