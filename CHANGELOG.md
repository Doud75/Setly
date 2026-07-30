# Changelog

Toutes les modifications notables de ce projet sont consignées dans ce fichier.

Le format s'appuie sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/),
et le projet suit le [versionnage sémantique](https://semver.org/lang/fr/).

Les catégories utilisées sont : `Ajouté`, `Modifié`, `Déprécié`, `Retiré`, `Corrigé`, `Sécurité`.

> Ce changelog démarre à l'ouverture du projet en open source. L'historique antérieur
> n'a pas été reconstitué : il reste consultable via `git log`.

## [Non publié]

### Ajouté

- Licence MIT (`LICENSE`).
- Guide de contribution (`CONTRIBUTING.md`) : workflow de fork, convention de commits, exécution des tests, migrations.
- Code de conduite (`CODE_OF_CONDUCT.md`), basé sur le Contributor Covenant 2.1.
- Politique de sécurité (`SECURITY.md`) et procédure de signalement privé des vulnérabilités.
- Vue d'ensemble de l'architecture (`ARCHITECTURE.md`).
- Templates GitHub : formulaires d'issue (bug, fonctionnalité) et template de Pull Request.
- Workflow `security.yml` : `govulncheck` et Trivy en version bloquante, une fois par semaine.

### Modifié

- Le job de sécurité de la CI devient informatif sur les Pull Requests, pour ne pas
  faire échouer la contribution d'un tiers à cause d'une CVE dans une dépendance.
- `.github/CODEOWNERS` : ajout du motif de chemin manquant, sans lequel la règle était
  ignorée par GitHub et aucune review n'était demandée automatiquement.
- L'action `trivy-action` est épinglée sur une version au lieu de suivre `master`.
- La CI annule désormais les exécutions rendues obsolètes par un nouveau push.
