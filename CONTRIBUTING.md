# Contribuer à Setly 🎸

Merci de l'intérêt que tu portes au projet ! Toute contribution est la bienvenue : correction de bug, nouvelle fonctionnalité, amélioration de la documentation ou simple signalement d'un problème.

Ce document décrit comment travailler sur le projet. Si quelque chose n'y est pas clair, c'est un bug de notre côté — [ouvre une discussion](https://github.com/Doud75/Setly/discussions) et on corrigera.

## Sommaire

- [Code de conduite](#code-de-conduite)
- [Signaler un bug, proposer une fonctionnalité ou une évolution design](#signaler-un-bug-proposer-une-fonctionnalité-ou-une-évolution-design)
- [Mettre en place l'environnement](#mettre-en-place-lenvironnement)
- [Workflow de contribution](#workflow-de-contribution)
- [Convention de commits](#convention-de-commits)
- [Lancer les tests](#lancer-les-tests)
- [Qualité de code](#qualité-de-code)
- [Migrations de base de données](#migrations-de-base-de-données)
- [Ce que fait la CI sur ta Pull Request](#ce-que-fait-la-ci-sur-ta-pull-request)
- [Publier une version (mainteneurs)](#publier-une-version-mainteneurs)

---

## Code de conduite

Ce projet suit un [code de conduite](CODE_OF_CONDUCT.md). En participant, tu acceptes de le respecter.

Pour signaler une **faille de sécurité**, n'ouvre pas d'issue publique : suis la procédure décrite dans [SECURITY.md](SECURITY.md).

---

## Signaler un bug, proposer une fonctionnalité ou une évolution design

Passe par les [issues GitHub](https://github.com/Doud75/Setly/issues). Trois formulaires sont proposés — un pour les bugs, un pour les propositions de fonctionnalité, un pour le design et l'interface — et ils te guident sur les informations à fournir.

Le formulaire **🎨 Design / UI** couvre tout ce qui touche au visuel et à l'expérience : incohérence de style, écran mal adapté au mobile, composant à harmoniser, accessibilité, mode sombre. Les captures d'écran et les maquettes (avant / après) y valent mieux qu'une longue description.

Avant d'ouvrir une issue, vérifie rapidement qu'elle n'existe pas déjà. Pour une simple question d'usage, préfère les [Discussions](https://github.com/Doud75/Setly/discussions).

---

## Mettre en place l'environnement

Le projet tourne entièrement dans Docker. Le [README](README.md#-démarrage-rapide) détaille les prérequis ; en résumé :

```bash
git clone https://github.com/Doud75/Setly.git
cd Setly
cp .env.example .env    # indispensable : le Makefile et docker compose lisent ce fichier
make up
```

- Frontend : <http://localhost:4000>
- API backend : <http://localhost:8089>

Les migrations sont appliquées automatiquement au démarrage par le service `migrator`.

> Le fichier `.env` contient des secrets (dont `JWT_SECRET`) et n'est **jamais** commité. En revanche `.env.test` **est** volontairement versionné : il ne contient que des valeurs de test jetables, ce qui permet de lancer la suite E2E sans aucune configuration.

Quelques commandes utiles au quotidien :

| Commande | Effet |
|---|---|
| `make up` | Démarre tous les services en arrière-plan |
| `make down` | Arrête et supprime les conteneurs |
| `make logs` | Suit les logs de tous les services |
| `make shell-back` / `make shell-front` | Ouvre un shell dans le conteneur correspondant |
| `make db-connect` | Ouvre `psql` sur la base de développement |
| `make redis-cli` | Ouvre le CLI Redis |

`make` seul (ou `make help`) affiche les cibles principales.

---

## Workflow de contribution

La branche `main` est protégée : **aucun push direct**. Tout passe par une Pull Request.

1. **Forke le dépôt** depuis l'interface GitHub, puis clone ton fork :

   ```bash
   git clone https://github.com/<ton-pseudo>/Setly.git
   cd Setly
   git remote add upstream https://github.com/Doud75/Setly.git
   ```

2. **Crée une branche** à partir d'un `main` à jour, avec un nom explicite :

   ```bash
   git switch main
   git pull upstream main
   git switch -c feat/mode-sombre
   ```

   Préfixes usuels : `feat/` pour une fonctionnalité, `fix/` pour un correctif, `docs/` pour la documentation, `refactor/` pour une restructuration sans changement de comportement.

3. **Développe.** Garde la PR focalisée sur un seul sujet : une PR qui corrige un bug *et* refactorise trois modules est beaucoup plus longue à relire, donc beaucoup plus lente à être fusionnée.

4. **Teste.** Lance au minimum `make test-all` (voir [Lancer les tests](#lancer-les-tests)). Idéalement, ajoute des tests couvrant ton changement.

5. **Vérifie le formatage** : `cd frontend && npm run lint`.

6. **Ouvre la Pull Request** vers `main`. Le template pré-remplit la description ; complète-le. Si la PR répond à une issue, mentionne-la (`Closes #42`).

7. **Revue de code.** Un mainteneur relit ton code. Une fois la PR approuvée et la CI au vert, elle est fusionnée.

Si `main` a avancé pendant ta PR, mets ta branche à jour avec `git pull --rebase upstream main`.

> **Fusionner ne déploie pas.** La production est déployée à partir de tags de version, pas de `main` : ton changement partira en ligne à la prochaine version publiée par un mainteneur. Tout ce que tu as à faire de ton côté, c'est décrire ton changement sous `## [Non publié]` dans le [`CHANGELOG.md`](CHANGELOG.md) s'il est visible par les utilisateurs — c'est ce texte qui deviendra les notes de la version.

---

## Convention de commits

L'historique du projet utilise un préfixe de type suivi d'un espace, de deux-points, puis de la description :

```
feat : ajout du mode sombre
fix : correction du calcul de durée des interludes
docs : mise à jour du guide d'installation
```

Types utilisés : `feat`, `fix`, `docs`, `refactor`, `test`, `chore`.

Ce n'est pas du [Conventional Commits](https://www.conventionalcommits.org/) strict et aucun outil ne le vérifie automatiquement — mais reste cohérent avec l'existant pour que l'historique demeure lisible. Un message en français ou en anglais, peu importe : sois simplement descriptif.

---

## Lancer les tests

| Commande | Ce qu'elle couvre |
|---|---|
| `make test-backend` | Tests unitaires Go (`backend/`) |
| `make test-backend-cover` | Idem + rapport de couverture dans `backend/coverage.html` |
| `make test-unit` | Tests unitaires frontend (Vitest) |
| `make test-unit-watch` | Vitest en mode watch, pour développer |
| `make test` | Suite End-to-End complète (Playwright) |
| `make test-all` | Les trois suites d'affilée |

**`make test` ne demande aucune configuration préalable.** La cible monte un environnement dédié à partir de `docker-compose.test.yml` (base PostgreSQL, Redis, backend et frontend sur le port 4001), applique les migrations, injecte des données de test, lance Playwright, puis démolit l'environnement.

Un rapport HTML détaillé est généré dans `frontend/playwright-report/index.html`.

### Cibles E2E granulaires

La suite complète est longue. Pendant le développement, cible la partie qui t'intéresse :

```bash
make test-song          # tout ce qui touche aux chansons
make test-setlist       # tout ce qui touche aux setlists
make test-auth          # connexion / inscription
make test-song-list     # uniquement la liste des chansons
make test-setlist-detail
```

Chacune gère son propre environnement de test. La liste complète est dans [`make/tests.mk`](make/tests.mk).

### Lancer un seul fichier de spec

Pour une boucle de feedback plus courte, monte l'environnement une fois et enchaîne les exécutions :

```bash
make test-up
cd frontend && npx playwright test tests/song/list.spec.ts
# ... itère autant de fois que nécessaire ...
make test-down
```

---

## Qualité de code

### Frontend

Le code est formaté avec **Prettier** et vérifié avec **ESLint** :

```bash
cd frontend
npm run lint      # vérifie le formatage (prettier --check) puis les règles ESLint
npm run format    # applique le formatage
npm run check     # vérification des types Svelte / TypeScript
```

Lance `npm run format` avant de commiter : la CI vérifie le formatage Prettier.

### Backend

Le code Go suit les conventions standard du langage. Formate avec `gofmt` avant de commiter :

```bash
cd backend
gofmt -w .
go vet ./...
```

### Quelques conventions du projet

- Les handlers HTTP retournent une `error` et sont enveloppés par `handler.Wrap()`, qui la sérialise en réponse JSON. Ne construis pas de réponse d'erreur à la main.
- Les erreurs métier sont des `apierror.AppError` typées (avec `Code`, `HTTPStatus`, `IsUserError`).
- La logique métier vit dans `api/service/`, l'accès aux données dans `api/repository/`. Les services dépendent d'interfaces de repository, ce qui permet de les tester avec les mocks de `api/repository/mocks/`.
- Le frontend utilise **Svelte 5 avec les runes** (`$state`, `$derived`, `$effect`), pas l'API réactive de Svelte 4.

Pour une vue d'ensemble de l'architecture, lis [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Migrations de base de données

Les migrations utilisent [`golang-migrate`](https://github.com/golang-migrate/migrate) et vivent dans `backend/db/migrations/`.

```bash
make migrate-create name=add_song_tempo   # crée la paire de fichiers .up.sql / .down.sql
make migrate                              # applique les migrations en attente
```

**Toute migration doit fournir son `down`.** Un fichier `.down.sql` vide ou incorrect rend le retour arrière impossible en production. Vérifie que ta migration s'applique *et* se retire proprement avant d'ouvrir ta PR.

Si ton changement modifie le schéma, pense aussi aux données de test dans `backend/db/seed/`.

---

## Ce que fait la CI sur ta Pull Request

La CI ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) se déclenche automatiquement sur toute PR vers `main`, **y compris depuis un fork**. Elle n'utilise aucun secret : tu n'as rien à configurer.

Trois jobs :

| Job | Contenu | Bloquant ? |
|---|---|---|
| **Security Scan** | `govulncheck` sur les dépendances Go, scan Trivy du système de fichiers | Non — informatif |
| **Lint & Build** | Build Go, tests unitaires backend avec couverture, ESLint, build SvelteKit | **Oui** |
| **Test** | Tests unitaires Vitest, suite E2E Playwright complète | **Oui** |

Le job de sécurité est volontairement **non bloquant** : une vulnérabilité publiée dans une dépendance transitive n'est pas de ton fait et ne doit pas faire échouer ta PR. Ces scans tournent en version bloquante une fois par semaine ([`security.yml`](.github/workflows/security.yml)), et les mainteneurs s'en occupent.

En cas d'échec E2E, le rapport Playwright est téléchargeable depuis l'onglet *Artifacts* de l'exécution GitHub Actions — c'est le point de départ pour comprendre ce qui a cassé.

---

## Publier une version (mainteneurs)

Cette section ne concerne pas les contributeurs : elle décrit comment un mainteneur met une version en production.

Le déploiement est déclenché par un **tag de version** `vX.Y.Z`, jamais par un merge sur `main`. Une seule commande fait tout :

| Commande | Effet |
|---|---|
| `make release-dry-run` | Affiche la prochaine version et le `CHANGELOG.md` résultant, sans rien modifier |
| `make release-patch` | Correctifs uniquement (`0.1.0` → `0.1.1`) |
| `make release-minor` | Fonctionnalités rétrocompatibles (`0.1.1` → `0.2.0`) |
| `make release-major` | Changement cassant (`0.2.0` → `1.0.0`) |

La commande ([`scripts/release.sh`](scripts/release.sh)) enchaîne :

1. bascule la section `## [Non publié]` du `CHANGELOG.md` en `## [X.Y.Z] - AAAA-MM-JJ` et met à jour les liens de comparaison en bas de fichier ;
2. commit `chore : release vX.Y.Z` poussé sur `main` ;
3. tag annoté `vX.Y.Z` poussé sur `origin` — ce qui déclenche [`cd.yml`](.github/workflows/cd.yml).

Elle demande confirmation avant de pousser quoi que ce soit, et refuse de publier si l'un de ces prérequis n'est pas rempli :

- tu es sur `main`, l'arbre de travail est propre et identique à `origin/main` ;
- la section `## [Non publié]` du `CHANGELOG.md` n'est pas vide — une version sans notes n'est pas publiable ;
- le tag calculé n'existe pas déjà.

Le commit de release va directement sur `main`, sans passer par une PR : c'est le seul cas où la règle « aucun push direct » est contournée, et il faut les droits d'administration sur le dépôt.

### Choisir le type de version

Le projet suit le [versionnage sémantique](https://semver.org/lang/fr/). Tant que la version majeure est `0`, on reste souple : `minor` pour une nouvelle fonctionnalité, `patch` pour un correctif.

### Ce que fait le déploiement

`cd.yml` construit les images backend et frontend, les publie sur GitHub Container Registry sous trois tags (`X.Y.Z`, `X.Y`, `latest`), puis se connecte en SSH au VPS. Là, il positionne le dépôt sur le tag (`git checkout refs/tags/vX.Y.Z`) avant de jouer les migrations — les migrations proviennent donc du commit exact à partir duquel les images ont été construites — et lance `docker compose up` en passant le tag via `IMAGE_TAG`.

### Revenir en arrière

Relance [`cd.yml`](.github/workflows/cd.yml) via **Actions → CD → Run workflow** en sélectionnant un tag antérieur : les images semver restent publiées sur GHCR, l'ancienne version est redéployée telle quelle.

⚠️ **Les migrations ne se rejouent pas à l'envers automatiquement.** Si la version fautive en contenait, il faut appliquer le `down` correspondant à la main sur le VPS avant de redéployer.

---

Merci encore, et bon code ! 🎶
