# Architecture de Setly

Ce document est la carte du projet pour quelqu'un qui le découvre. Il explique comment les pièces s'assemblent et pourquoi, pas comment installer le projet — pour ça, voir le [README](README.md), et pour contribuer, [CONTRIBUTING.md](CONTRIBUTING.md).

## Vue d'ensemble

Le dépôt est un **monorepo** contenant deux applications qui se déploient ensemble :

```
.
├── backend/        API REST en Go, sans framework
├── frontend/       Application SvelteKit (PWA)
├── make/           Cibles Makefile réparties par domaine (docker, db, tests)
└── docker-compose*.yml
```

Quatre services tournent en développement : `frontend`, `backend`, `db` (PostgreSQL) et `redis`, plus un service éphémère `migrator` qui applique les migrations au démarrage.

## Le point le plus important : le navigateur n'appelle jamais le backend

C'est la décision de conception la moins évidente, et celle qui déroute le plus vite quelqu'un qui arrive sur le projet.

```
┌───────────┐   fetch('/api/…')    ┌──────────────────┐   HTTP + en-têtes   ┌──────────────┐
│ Navigateur│ ───────────────────► │ Serveur SvelteKit│ ──────────────────► │  Backend Go  │
│           │                      │    (Node.js)     │  Authorization      │              │
│  cookies  │ ◄─────────────────── │                  │  X-Band-ID          │              │
│ httpOnly  │                      └──────────────────┘                     └──────────────┘
└───────────┘                                                                       │
                                                                                    ▼
                                                                          PostgreSQL + Redis
```

Le serveur SvelteKit joue le rôle de **proxy authentifiant**. Concrètement :

- Les tokens JWT vivent dans des cookies gérés côté serveur (`jwt_token`, `refresh_token`). Le JavaScript qui tourne dans le navigateur n'y a jamais accès et ne voit jamais de token.
- `frontend/src/routes/api/[...slug]/+server.ts` intercepte toute requête vers `/api/*`, la relaie au backend Go, et injecte au passage les en-têtes `Authorization` et `X-Band-ID` à partir de `event.locals`.
- Le backend Go n'est donc **pas exposé publiquement** : seul le conteneur frontend le joint, via le réseau Docker interne.

Conséquence pratique quand tu développes : côté client, tu appelles simplement `/api/songs`. Tu n'ajoutes jamais d'en-tête d'authentification à la main — c'est le proxy qui s'en charge.

## Backend — Go, sans framework

`backend/main.go` câble toutes les dépendances et enregistre les routes sur le `net/http.ServeMux` standard. Toutes les routes sont préfixées par `/api/`.

### Découpage en couches

```
handler/     ← parse la requête HTTP, appelle le service, renvoie la réponse
   ↓
service/     ← logique métier, règles, orchestration
   ↓
repository/  ← accès PostgreSQL via pgx/v5
```

Chaque couche ne connaît que la suivante. Les services dépendent d'**interfaces** de repository, pas d'implémentations : c'est ce qui permet de les tester unitairement avec les mocks de `api/repository/mocks/`, sans base de données.

| Dossier | Rôle |
|---|---|
| `api/handler/` | Handlers HTTP |
| `api/service/` | Logique métier |
| `api/repository/` | Accès aux données (`mocks/` pour les tests) |
| `api/model/` | Structures des entités (`Song`, `Setlist`, `Band`, `User`, `Interlude`, `Invitation`…) |
| `api/apierror/` | Système d'erreurs typées |
| `api/middleware/` | Authentification, autorisation, limitation de débit |
| `auth/` | Création et validation des JWT, hachage des mots de passe |
| `cache/` | Client Redis |
| `config/` | Chargement des variables d'environnement |
| `db/migrations/` | Migrations SQL `golang-migrate` |
| `db/seed/` | Jeu de données pour les tests E2E |

### Gestion des erreurs

Les handlers **retournent une `error`** au lieu d'écrire eux-mêmes une réponse d'échec. Ils sont enveloppés par `handler.Wrap()`, qui traduit l'erreur en réponse JSON.

Les erreurs métier sont des `apierror.AppError`, qui portent :

- `Code` — un identifiant stable, exploitable par le frontend
- `HTTPStatus` — le code HTTP à renvoyer
- `IsUserError` — distingue une erreur imputable à l'utilisateur (à afficher tel quel) d'une erreur interne (à logger, à masquer)

En pratique : si tu écris un handler, retourne une `apierror` plutôt que d'appeler `http.Error`. Tu gagnes la sérialisation, le bon statut et le bon niveau de log gratuitement.

### Authentification et contexte de groupe

Deux middlewares d'authentification, selon ce dont la route a besoin :

- **`JWTAuthUserOnly`** — valide le JWT et identifie l'utilisateur. Suffisant pour les routes qui ne dépendent d'aucun groupe (profil, liste des groupes de l'utilisateur…).
- **`JWTAuth`** — valide le JWT **et** exige un en-tête `X-Band-ID`. Le middleware vérifie que l'utilisateur appartient bien à ce groupe, puis attache l'appartenance et le rôle au contexte de la requête.
- **`AdminOnly`** — s'empile par-dessus `JWTAuth` pour les actions réservées aux administrateurs du groupe.

> ⚠️ **Règle de sécurité à ne pas enfreindre.** L'identifiant du groupe doit toujours être lu **depuis le contexte de la requête**, jamais depuis le corps ou les paramètres d'URL. Lire un `band_id` fourni par le client permettrait à un utilisateur d'accéder aux données d'un groupe auquel il n'appartient pas. C'est le principal vecteur de fuite dans une application multi-tenant comme celle-ci.

## Frontend — SvelteKit et Svelte 5

Le frontend utilise **Svelte 5 avec les runes** (`$state`, `$derived`, `$effect`), pas la syntaxe réactive de Svelte 4. Styling avec Tailwind CSS.

### Organisation des routes

```
src/routes/
├── (auth)/       Connexion, inscription — layout non authentifié
├── (app)/        Application authentifiée : dashboard, song, setlist, interlude, settings
├── (public)/     Pages accessibles sans compte
├── api/[...slug] Le proxy vers le backend Go
├── switch-band/  Changement de groupe actif
└── offline/      Page de repli du service worker
```

`(app)/+layout.server.ts` fait office de garde : il redirige vers `/login` si l'utilisateur n'est pas connecté, et vers `/dashboard` s'il n'a pas de groupe actif sélectionné.

### `hooks.server.ts` — le cœur de la session

`src/hooks.server.ts` s'exécute sur chaque requête serveur. Il lit les cookies, **rafraîchit automatiquement le token expiré**, récupère les informations de l'utilisateur et remplit `event.locals` (`user`, `token`, `activeBandId`) — que toutes les routes serveur consomment ensuite.

Le mécanisme de rafraîchissement porte une contrainte de scalabilité importante, détaillée dans la [section dédiée du README](README.md#-authentification--rafraîchissement-des-tokens) : il repose sur un *single-flight* en mémoire, valable tant que le frontend tourne dans **un seul conteneur Node**. À lire avant toute réflexion sur une mise à l'échelle horizontale.

### Autres dossiers

| Dossier | Contenu |
|---|---|
| `lib/components/` | Composants réutilisables, groupés par domaine (`setlist/`, `song/`, `ui/`) |
| `lib/server/` | Code exécuté uniquement côté serveur (ex. `songActions.ts`) |
| `lib/stores/` | Stores Svelte (ex. `modalStore.ts`) |
| `lib/utils/` | Utilitaires client, dont `pdfGenerator.ts` pour l'export PDF |
| `lib/types.ts` | Types TypeScript partagés |
| `src/service-worker.ts` | Service worker : mode hors-ligne et installation PWA |

## Multi-groupe

Un utilisateur peut appartenir à **plusieurs groupes**. Le groupe courant est mémorisé dans le cookie `active_band_id` et transmis au backend via l'en-tête `X-Band-ID`.

- `/dashboard` liste les groupes de l'utilisateur et permet d'en choisir un.
- `/switch-band` bascule d'un groupe à l'autre.
- Le middleware backend revalide l'appartenance à chaque requête — le cookie n'est jamais une source de confiance à lui seul.

C'est la dimension la plus facile à casser accidentellement : toute nouvelle requête de données doit être filtrée par le groupe issu du contexte. Les tests E2E `make test-cross-band` et `make test-multi-group` existent précisément pour attraper ces régressions.

## Persistance

- **PostgreSQL** — stockage principal, accédé via `pgx/v5` (pool de connexions). Le schéma évolue uniquement par migrations `golang-migrate` dans `backend/db/migrations/`, en paires `.up.sql` / `.down.sql`.
- **Redis** — cache et données à durée de vie courte, via le wrapper `backend/cache/`.

## Tests

| Niveau | Outil | Emplacement |
|---|---|---|
| Unitaires backend | `go test` + mocks de repository | à côté du code testé, dans `backend/` |
| Unitaires frontend | Vitest | `frontend/` |
| End-to-End | Playwright | `frontend/tests/` |

Les tests E2E tournent contre un environnement complet et isolé décrit par `docker-compose.test.yml` (base et Redis dédiés, frontend sur le port 4001), amorcé par `backend/db/seed/`. Voir [CONTRIBUTING.md](CONTRIBUTING.md#lancer-les-tests) pour les commandes.

## Déploiement

- **CI** (`.github/workflows/ci.yml`) — sur chaque PR : scans de sécurité, build, lint, tests unitaires et E2E.
- **CD** (`.github/workflows/cd.yml`) — sur chaque tag de version `vX.Y.Z` : construction et publication des images Docker sur GitHub Container Registry, taguées en semver (`X.Y.Z`, `X.Y`, `latest`), puis déploiement par SSH sur le VPS (migrations, `docker compose up`, vérification de santé).

Un merge sur `main` ne déploie rien : le déploiement est déclenché par la publication d'une version, via `make release-patch|minor|major` (voir [CONTRIBUTING.md](CONTRIBUTING.md#publier-une-version-mainteneurs)). Sur le VPS, le dépôt est positionné sur le tag déployé (`git checkout refs/tags/vX.Y.Z`) avant de jouer les migrations : celles-ci proviennent donc exactement du commit à partir duquel les images ont été construites. Le tag d'image déployé est passé au compose de production via `IMAGE_TAG`, ce qui rend un retour arrière possible en rejouant `cd.yml` (`workflow_dispatch`) sur un tag antérieur.
