# Frontend — Setly

Ce service est le frontend en SvelteKit de l'application Setly. Il fournit l'interface utilisateur pour interagir avec l'API backend et est conçu pour être une Progressive Web App (PWA) installable et utilisable hors ligne.

## Stack Technique

- **Framework** : SvelteKit (utilisant Svelte 5 et les Runes)
- **Styling** : Tailwind CSS
- **Langage** : TypeScript
- **PWA** : `@vite-pwa/sveltekit`
- **Bundler** : Vite
- **Containerisation** : Docker

---

## Prérequis

- [Node.js](https://nodejs.org/) (version 22 ou supérieure)
- `npm` (généralement inclus avec Node.js)
- [Docker](https://docs.docker.com/get-docker/) et [Docker Compose](https://docs.docker.com/compose/install/)

---

## Démarrage Rapide

Toutes les commandes sont à lancer depuis la **racine du dépôt**, pas depuis le dossier `frontend`.

### 1. Configuration (Première fois uniquement)

Assurez-vous d'avoir créé le fichier `.env` à la racine du projet en copiant `.env.example`.

### 2. Installation des Dépendances (Première fois uniquement)

```bash
# Naviguer dans le dossier frontend
cd frontend

# Installer les paquets npm
npm install

# Revenir à la racine
cd ..
```

### 3. Lancer le Projet Complet (Recommandé)

Cette méthode lance le frontend, le backend et la base de données ensemble, assurant que tout communique correctement.

```bash
make up
```
L'application sera accessible à l'adresse `http://localhost:4000`.

### 4. Lancer le Frontend Seul (pour le développement UI)

Si vous travaillez uniquement sur des changements visuels et que le backend tourne déjà (ou que vous n'en avez pas besoin), vous pouvez utiliser le serveur de développement de Vite.

```bash
# Naviguer dans le dossier frontend
cd frontend

# Lancer le serveur de dev
npm run dev
```
L'application sera accessible à l'adresse `http://localhost:5173` (ou un autre port si celui-ci est pris).

---

## Scripts Utiles (`package.json`)

Les commandes suivantes sont à lancer depuis le dossier `frontend/`.

- `npm run dev`: Démarre le serveur de développement avec rechargement à chaud (HMR).
- `npm run build`: Construit l'application pour la production.
- `npm run preview`: Prévisualise la version de production.
- `npm run lint`: Vérifie les erreurs de style et de code avec ESLint.
- `npm run check`: Lance le vérificateur de types de Svelte (TypeScript).

---

## Structure du Projet

- `src/routes/`: Contient toutes les pages de l'application. Le nom des dossiers définit les URLs.
    - `(auth)/`: Un groupe de layout pour les pages de connexion, inscription, etc.
- `src/lib/`: Code réutilisable (composants, stores, utilitaires).
    - `components/ui/`: Composants d'interface génériques (Button, Input...).
    - `stores/`: Stores Svelte pour la gestion d'état global (ex: `auth.ts`).
- `static/`: Fichiers publics (icônes, `robots.txt`...).

##