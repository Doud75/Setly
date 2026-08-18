# Backend — API Setly

Ce service est l'API en Go de l'application Setly. Il gère l'authentification des utilisateurs, la gestion des groupes (bands), et toutes les données relatives aux chansons et aux setlists.

## Stack Technique

- **Langage** : Go 1.21+ (sans framework web)
- **Base de données** : PostgreSQL
- **Driver DB** : `pgx/v5` (pool de connexions)
- **Migrations** : `golang-migrate/migrate`
- **Authentification** : JWT (JSON Web Tokens)
- **Containerisation** : Docker

---

## Prérequis

Avant de commencer, assurez-vous d'avoir installé :
- [Go](https://go.dev/doc/install) (version 1.21 ou supérieure)
- [Docker](https://docs.docker.com/get-docker/) et [Docker Compose](https://docs.docker.com/compose/install/)
- L'outil de migration :
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
  *(Si la commande `migrate` n'est pas trouvée, assurez-vous que `$(go env GOPATH)/bin` est dans votre `$PATH`)*

---

## Démarrage Rapide (avec Docker)

Toutes les commandes sont à lancer depuis la **racine du dépôt**, pas depuis le dossier `backend`.

### 1. Configuration (Première fois uniquement)

Copiez le fichier d'exemple pour les variables d'environnement.
```bash
cp .env.example .env
```
Personnalisez les valeurs dans `.env` si nécessaire (surtout `JWT_SECRET`).

### 2. Lancer la Base de Données et les Migrations (Première fois uniquement)

```bash
# 1. Démarre uniquement le conteneur de la base de données
make db-up

# 2. Applique les schémas de table à la base de données
make migrate
```

### 3. Lancer le Projet Complet

Pour démarrer tous les services (backend, frontend, db) :
```bash
make up
```
L'API sera accessible à l'adresse `http://localhost:8089`.

---

## Commandes Utiles (`Makefile`)

- `make up`: Démarre tous les services en mode détaché.
- `make down`: Arrête et supprime tous les conteneurs.
- `make logs`: Affiche les logs de tous les services en temps réel.
- `make shell-back`: Ouvre un terminal `sh` à l'intérieur du conteneur du backend.

## Gestion de la Base de Données

- **Créer une nouvelle migration** : `make migrate-create name=le_nom_de_la_migration`
- **Appliquer les migrations** : `make migrate`
- **Se connecter à la base de données** : `make db-connect`

---

## Endpoints de l'API

L'API expose 36 routes, toutes préfixées par `/api/`. La référence à jour, avec les payloads
et les en-têtes attendus, est la **collection Postman** du dépôt :

- [`backend/postman/setly.postman_collection.json`](postman/setly.postman_collection.json) — à importer dans Postman
- [`backend/postman/README.md`](postman/README.md) — comment s'en servir

En deux mots :

| Domaine | Routes | Auth |
| --- | --- | --- |
| Auth | `POST /api/auth/{signup,login,refresh,logout}` | publique |
| Utilisateur | `GET /api/user/{info,bands}`, `PUT /api/user/{password,default-band}` | JWT seul |
| Groupes | `POST /api/bands`, `GET` et `POST /api/bands/{bandId}/members`, `PUT .../role`, `DELETE .../{userId}`, `DELETE .../me` | JWT + `X-Band-ID` |
| Invitations | `POST /api/bands/{bandId}/invitations`, `GET /api/invitations/{token}`, `POST .../accept` | mixte |
| Chansons | `GET` et `POST /api/song`, `GET`, `PUT`, `DELETE /api/song/{id}` | JWT + `X-Band-ID` |
| Setlists | `GET` et `POST /api/setlist`, `GET`, `PUT`, `DELETE /api/setlist/{id}`, `POST .../duplicate`, `POST .../items`, `PUT .../items/order`, `PUT`, `DELETE /api/setlist/item/{itemId}` | JWT + `X-Band-ID` |
| Interludes | `GET` et `POST /api/interlude`, `PUT /api/interlude/{id}` | JWT + `X-Band-ID` |
| Système | `GET /api/health` | publique |

Deux règles transverses :

- Les routes du groupe actif exigent l'en-tête **`X-Band-ID`** en plus du `Authorization: Bearer <jwt>` ;
  sans lui l'API répond `400 En-tête X-Band-ID manquant.`
- Les suppressions et la gestion des membres passent par `middleware.AdminOnly` : `403` pour un simple membre.

Les erreurs suivent toutes le format de `api/apierror` :
`{ "error": { "code": "...", "message": "..." } }`.
