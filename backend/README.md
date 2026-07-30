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

### Authentification

- `POST /api/auth/signup` : Crée un nouveau groupe (band) et son premier utilisateur.
  ```json
  {
    "band_name": "Mon Super Groupe",
    "username": "mon_user",
    "password": "mon_password"
  }
  ```

- `POST /api/auth/join` : Crée un nouvel utilisateur et le rattache à un groupe existant.
  ```json
  {
    "band_name": "Mon Super Groupe",
    "username": "autre_user",
    "password": "autre_password"
  }
  ```

- `POST /api/auth/login` : Connecte un utilisateur et retourne un token JWT.
  ```json
  {
    "username": "mon_user",
    "password": "mon_password"
  }
  ```