# Setly 🎸

[![CI](https://github.com/Doud75/Setly/actions/workflows/ci.yml/badge.svg)](https://github.com/Doud75/Setly/actions/workflows/ci.yml)
[![Licence : MIT](https://img.shields.io/badge/licence-MIT-blue.svg)](LICENSE)
[![PRs bienvenues](https://img.shields.io/badge/PRs-bienvenues-brightgreen.svg)](CONTRIBUTING.md)

Bienvenue sur **Setly** ! Cette application est une *Progressive Web App* conçue pour aider les groupes de musique à créer, gérer et partager leurs setlists.

📖 [Contribuer](CONTRIBUTING.md) · [Architecture](ARCHITECTURE.md) · [Sécurité](SECURITY.md) · [Changelog](CHANGELOG.md)

## ✨ Fonctionnalités

*   **Gestion de Bibliothèque** : Ajoutez et organisez toutes vos chansons et interludes.
*   **Création de Setlists** : Créez des setlists personnalisées par simple glisser-déposer.
*   **Calcul de Durée** : Calcule automatiquement la durée totale de vos concerts.
*   **Mode PWA** : Installez l'application sur votre téléphone pour un accès rapide.
*   **Export PDF** : Générez un PDF propre de vos setlists pour l'impression ou le partage.
*   **Collaboratif** : Partagez l'accès à votre groupe pour que tout le monde soit synchronisé.

## 🛠️ Stack Technique

Ce projet est un monorepo qui combine un backend en Go avec un frontend moderne en SvelteKit.

*   **Frontend** (`/frontend`):
    *   **Framework** : [SvelteKit](https://kit.svelte.dev/) (avec Svelte 5 & Runes)
    *   **Langage** : TypeScript
    *   **Styling** : [Tailwind CSS](https://tailwindcss.com/)
    *   **PWA** : `@vite-pwa/sveltekit`
    *   **Tests** : Vitest (Unitaires) & Playwright (End-to-End)

*   **Backend** (`/backend`):
    *   **Langage** : [Go](https://go.dev/) (API REST native sans framework)
    *   **Base de données** : PostgreSQL
    *   **Authentification** : JWT (JSON Web Tokens)
    *   **Migrations** : `golang-migrate`

*   **Environnement** :
    *   **Orchestration** : [Docker](https://www.docker.com/) & Docker Compose
    *   **Automatisation** : `Makefile` pour des commandes simplifiées
    *   **Intégration Continue** : GitHub Actions

## 🔐 Authentification & rafraîchissement des tokens

> ⚠️ **Contrainte de scalabilité du frontend.**
> Le rafraîchissement des JWT (`frontend/src/hooks.server.ts`) utilise un **single-flight en mémoire process** : pour un même `refresh_token`, un seul appel `/auth/refresh` est émis et son résultat est partagé/caché (~60 s) entre toutes les requêtes concurrentes. Cela évite la course de rotation qui, en PWA sur mauvaise connexion, déclenchait plusieurs rotations parallèles et déconnectait l'utilisateur (`refresh token not found`).
>
> Ce mécanisme repose sur le fait que **le frontend tourne en un seul conteneur Node**. Si un jour le frontend est scalé horizontalement (plusieurs réplicas), la coalescence en mémoire ne suffira plus : il faudra basculer sur une **grace period côté backend** (tolérer brièvement l'ancien refresh token après rotation), ou un store de coalescence partagé (Redis).

## 🚀 Démarrage Rapide

Le projet est entièrement conteneurisé avec Docker, ce qui simplifie grandement l'installation.

### Prérequis

Avant de commencer, assurez-vous d'avoir installé les outils suivants sur votre machine :

1.  **[Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)** : Pour lancer l'environnement de développement.
2.  **[Make](https://www.gnu.org/software/make/)** : Pour utiliser les commandes simplifiées du `Makefile`. (généralement pré-installé sur Linux/macOS).
3.  **[Go](https://go.dev/doc/install)** (v1.21+) : Uniquement si vous devez gérer les migrations manuellement.
4.  **[Node.js](https://nodejs.org/)** (v22+) : Uniquement si vous souhaitez développer le frontend en dehors de Docker.

### 1. Configuration Initiale

Commencez par cloner le projet et configurer vos variables d'environnement.

```bash
# Clonez le dépôt
git clone https://github.com/Doud75/Setly.git
cd Setly

# Copiez le fichier d'environnement d'exemple
cp .env.example .env
```

> **Important** : Le fichier `.env` contient des secrets (comme `JWT_SECRET`). Il est ignoré par Git et ne doit jamais être partagé.

### 2. Lancer l'environnement de développement

Grâce au `Makefile`, lancer tous les services est un jeu d'enfant.

```bash
# Construit les images Docker et démarre tous les services (frontend, backend, DB)
make up
```

Une fois la commande terminée :
*   🚀 **Le Frontend** sera accessible sur [http://localhost:4000](http://localhost:4000)
*   ⚙️ **L'API Backend** sera accessible sur [http://localhost:8089](http://localhost:8089)

Les migrations de la base de données sont appliquées automatiquement au démarrage grâce au service `migrator`.

### Commandes `Makefile` utiles

Voici les commandes que vous utiliserez le plus souvent :

*   `make up` : Démarre tous les services en arrière-plan.
*   `make down` : Arrête et supprime tous les conteneurs.
*   `make logs` : Affiche les logs de tous les services en temps réel.
*   `make shell-back` : Ouvre un terminal `sh` dans le conteneur du backend.
*   `make shell-front` : Ouvre un terminal `sh` dans le conteneur du frontend.

Pour voir toutes les commandes, consultez le fichier `Makefile`.

## 🧪 Lancer les Tests

Le projet est équipé d'une suite de tests complète.

### Tests Unitaires (Frontend)

Ces tests vérifient de petites parties isolées du code frontend (fonctions utilitaires, etc.).

```bash
# Lance les tests unitaires une seule fois
make test-unit

# Lance les tests unitaires en mode "watch" pour le développement
make test-unit-watch
```

### Tests End-to-End (Playwright)

Ces tests simulent une interaction utilisateur complète dans un navigateur. Ils nécessitent un environnement de test dédié.

```bash
# Lance toute la suite de tests E2E (crée un environnement de test complet)
make test
```

> Après l'exécution, un rapport de test HTML détaillé est généré dans `frontend/playwright-report/index.html`.

Vous pouvez également lancer des sous-ensembles de tests (ex: `make test-song-list`) pour un débogage plus rapide.

## 🤝 Comment Contribuer

Ce projet est ouvert aux contributions ! Le guide complet — workflow de fork, convention de commits, exécution des tests, migrations, ce que fait la CI sur votre PR — se trouve dans **[CONTRIBUTING.md](CONTRIBUTING.md)**.

En résumé : la branche `main` est protégée, tout passe par une Pull Request depuis un fork, et `make test-all` doit passer avant de l'ouvrir.

Pour comprendre le fonctionnement interne du projet avant de vous lancer, lisez **[ARCHITECTURE.md](ARCHITECTURE.md)**.

En participant, vous acceptez de respecter notre [code de conduite](CODE_OF_CONDUCT.md).

## 🔒 Sécurité

Pour signaler une faille, n'ouvrez pas d'issue publique : suivez la procédure décrite dans [SECURITY.md](SECURITY.md).

## 📄 Licence

Ce projet est distribué sous licence [MIT](LICENSE).
