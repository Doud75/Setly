# Changelog

Toutes les modifications notables de ce projet sont consignées dans ce fichier.

Le format s'appuie sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/),
et le projet suit le [versionnage sémantique](https://semver.org/lang/fr/).

Les catégories utilisées sont : `Ajouté`, `Modifié`, `Déprécié`, `Retiré`, `Corrigé`, `Sécurité`.

> Ce changelog démarre à l'ouverture du projet en open source. L'historique antérieur
> n'a pas été reconstitué : il reste consultable via `git log`.

## [Non publié]

## [0.1.0] - 2026-08-06

Première release publique, à l'ouverture du projet en open source.

### Ajouté

- Licence MIT (`LICENSE`).
- Guide de contribution (`CONTRIBUTING.md`) : workflow de fork, convention de commits, exécution des tests, migrations.
- Code de conduite (`CODE_OF_CONDUCT.md`), basé sur le Contributor Covenant 2.1.
- Politique de sécurité (`SECURITY.md`) et procédure de signalement privé des vulnérabilités.
- Vue d'ensemble de l'architecture (`ARCHITECTURE.md`).
- Templates GitHub : formulaires d'issue (bug, fonctionnalité) et template de Pull Request.
- Workflow `security.yml` : `govulncheck` et Trivy en version bloquante, une fois par semaine.
- Cible `make test-backend-integration` : tests backend exécutés contre une vraie base
  PostgreSQL, avec un test de concurrence sur l'ajout d'éléments à une setlist.

### Modifié

- Le job de sécurité de la CI devient informatif sur les Pull Requests, pour ne pas
  faire échouer la contribution d'un tiers à cause d'une CVE dans une dépendance.
- `.github/CODEOWNERS` : ajout du motif de chemin manquant, sans lequel la règle était
  ignorée par GitHub et aucune review n'était demandée automatiquement.
- L'action `trivy-action` est épinglée sur une version au lieu de suivre `master`.
- La CI annule désormais les exécutions rendues obsolètes par un nouveau push.

### Corrigé

- L'ajout d'un morceau à une setlist n'était pas transactionnel : la position était calculée
  puis insérée en deux requêtes distinctes. Deux membres ajoutant un morceau à la même setlist
  au même moment obtenaient la même position, la seconde insertion violait la contrainte
  `unique_position_in_setlist` et l'utilisateur recevait une erreur 500 ([#80](https://github.com/Doud75/Setly/issues/80)).

[Non publié]: https://github.com/Doud75/Setly/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Doud75/Setly/releases/tag/v0.1.0
