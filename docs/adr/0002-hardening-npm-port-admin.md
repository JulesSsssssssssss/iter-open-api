# ADR 002 — Port admin Nginx Proxy Manager : rebind localhost plutôt que règle UFW

## Contexte

Un scan nmap externe du VPS a révélé que le port `81` (interface d'administration
de Nginx Proxy Manager) était accessible publiquement, alors qu'UFW avait pour
politique par défaut *deny incoming* et ne contenait aucune règle `ALLOW 81`.

Investigation : Docker gère ses propres règles `iptables` (chaîne
`DOCKER-USER` / NAT) pour tout port publié via `ports:` dans un compose file,
et ces règles sont insérées **avant** la chaîne `ufw-user-input`. Un port publié
par Docker avec `0.0.0.0:81:81` reste donc joignable depuis l'extérieur même si
UFW le refuse — un piège classique de la combinaison Docker + UFW, qui ne peut
pas se corriger en ajoutant une règle UFW supplémentaire.

## Décision

J'utilise le rebind du port sur l'interface de loopback dans le compose de
Nginx Proxy Manager, plutôt que d'essayer de compenser via UFW
(`ufw-docker` ou règles `DOCKER-USER` manuelles) :

```yaml
ports:
  - "127.0.0.1:81:81"   # au lieu de "81:81" (= 0.0.0.0:81:81)
```

Docker ne publie alors plus ce port que sur `127.0.0.1` : uniquement
accessible depuis le VPS lui-même, inatteignable depuis l'extérieur quelle que
soit la configuration UFW.

## Conséquences

+ L'admin NPM reste utilisable en HTTPS via son propre reverse proxy
  (`proxy.jules.ruberti.mds-nantes.fr`), aucune perte de fonctionnalité.
+ Solution garantie au niveau réseau (binding d'interface), pas dépendante
  d'une règle de pare-feu applicative qui pourrait être oubliée ou contournée.
+ Validé par un nouveau scan nmap externe (1000+ ports) : seuls 22, 80, 443
  restent ouverts.
- Le mot de passe admin NPM a dû être réinitialisé par précaution (exposition
  passée), ce qui suppose de reconfigurer l'accès sur tous les postes admin.
- Réflexe à reproduire manuellement pour chaque futur service Docker exposant
  un port d'administration sur ce VPS — rien ne l'impose automatiquement.
