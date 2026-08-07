# Politique de sécurité

## Versions supportées

Le projet est déployé à partir de tags de version (`vX.Y.Z`). Seule la dernière version publiée reçoit des correctifs de sécurité ; il n'y a pas de branches de maintenance pour les versions antérieures. Un correctif est diffusé par la publication d'une nouvelle version.

| Version | Supportée |
|---|---|
| Dernière version publiée (`vX.Y.Z`) | ✅ |
| Versions antérieures | ❌ |

## Signaler une vulnérabilité

**N'ouvre pas d'issue publique pour une faille de sécurité.** Une issue est visible immédiatement par tout le monde, y compris par quelqu'un qui pourrait exploiter la faille avant qu'un correctif existe.

Utilise le signalement privé de GitHub :

👉 **[Signaler une vulnérabilité](https://github.com/Doud75/Setly/security/advisories/new)**

(Depuis le dépôt : onglet **Security** → **Report a vulnerability**.)

Ce canal est privé — seuls les mainteneurs voient le rapport — et permet de discuter du correctif, puis de publier un avis de sécurité une fois la faille corrigée.

### Ce qu'il est utile d'inclure

Plus le rapport est précis, plus le correctif arrive vite :

- Le type de vulnérabilité (injection SQL, XSS, contournement d'authentification, exposition de données inter-groupes, etc.)
- Les fichiers ou endpoints concernés
- Les étapes de reproduction, ou un proof of concept
- L'impact que tu estimes : quelles données ou quelles actions sont exposées
- Toute configuration particulière nécessaire pour reproduire

### Ce à quoi tu peux t'attendre

- **Accusé de réception sous 7 jours.**
- Une évaluation de la faille et une estimation du délai de correction.
- Une information quand le correctif est déployé.
- Un crédit dans l'avis de sécurité publié, si tu le souhaites.

Ce projet est un projet personnel maintenu bénévolement : il n'existe aucun programme de bug bounty, et les délais sont ceux d'une équipe d'une personne. Merci d'en tenir compte.

## Points d'attention connus

Quelques éléments de conception utiles à connaître si tu audites le projet :

- **Le navigateur n'appelle jamais le backend Go directement.** Toutes les requêtes passent par le serveur SvelteKit, qui injecte les en-têtes `Authorization` et `X-Band-ID` à partir de cookies gérés côté serveur. Les tokens ne sont pas exposés au JavaScript client.
- **Le cloisonnement entre groupes** repose sur l'en-tête `X-Band-ID` et le middleware d'authentification qui vérifie l'appartenance au groupe. Toute route protégée qui lirait un identifiant de groupe depuis le corps de la requête plutôt que depuis le contexte serait un bug de sécurité — les signalements sur ce point sont particulièrement bienvenus.
- **Les actions réservées aux administrateurs** sont protégées par le middleware `AdminOnly` empilé sur `JWTAuth`.

## Vérifications automatiques

Chaque Pull Request déclenche :

- [`govulncheck`](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) sur les dépendances Go
- un scan [Trivy](https://github.com/aquasecurity/trivy) du système de fichiers

Sur les PR, ces scans sont **informatifs** : ils n'échouent pas la CI, pour ne pas bloquer un contributeur sur une CVE dont il n'est pas responsable. Ils tournent en version **bloquante** une fois par semaine via [`.github/workflows/security.yml`](.github/workflows/security.yml), et les mainteneurs traitent les alertes à partir de là.
