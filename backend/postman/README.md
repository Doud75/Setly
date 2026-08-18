# Collection Postman — API Setly

`setly.postman_collection.json` couvre les **36 routes** exposées par `backend/main.go`
(collection Postman v2.1, importable aussi dans Insomnia, Bruno ou Hoppscotch).

## Importer

Postman → **Import** → glisser `backend/postman/setly.postman_collection.json`.

Aucun fichier d'environnement n'est nécessaire : les variables vivent dans la collection
(`baseUrl` pointe sur `http://localhost:8089`).

## Utiliser

1. `make up` à la racine du dépôt.
2. Renseigner les variables `username` et `password` de la collection (onglet *Variables*).
   Elles sont vides dans le fichier versionné — voir *Ne rien commiter de sensible* plus bas.
   Le nom d'utilisateur `testuser` du seed (`backend/db/seed`) n'existe que dans
   l'environnement de test, pas en développement.
3. Lancer **Auth → Connexion**. Son script de test enregistre `token`, `refreshToken` et
   `bandId` dans les variables de la collection — toutes les autres requêtes s'en servent.
   Sur une base vierge, lancer **Auth → Inscription** à la place.
4. Les dossiers se parcourent de haut en bas : chaque requête de création enregistre l'id
   qu'elle retourne (`songId`, `setlistId`, `itemId`, `interludeId`, `inviteToken`), donc la
   requête suivante trouve toujours une ressource à manipuler.

## Ne rien commiter de sensible

Le dépôt est public. Le fichier versionné ne contient **aucun identifiant** : `username`,
`password`, `token`, `refreshToken` et tous les ids sont exportés vides, et le corps de
*Ajouter un membre* porte le placeholder `REMPLACE_MOI`.

Deux réflexes pour que ça le reste :

- Postman distingue la *initial value* (exportée, donc versionnée) de la *current value*
  (locale). Les scripts de test écrivent la current value — un token ne part donc pas dans
  un export. En revanche « Persist All » recopie current → initial : à éviter sur cette
  collection.
- Les environnements Postman exportés sont ignorés par git (`*.postman_environment.json`
  dans le [`.gitignore`](../../.gitignore) racine). Si tu préfères garder tes identifiants
  dans un environnement plutôt que dans les variables de collection, c'est la voie sûre.

## Ce qu'il faut savoir sur l'API

- **Auth** : `Bearer {{token}}` est défini au niveau de la collection et hérité partout ;
  les routes publiques (`signup`, `login`, `refresh`, `GET /api/invitations/{token}`,
  `health`) le désactivent individuellement.
- **Groupe actif** : la plupart des routes passent par `middleware.JWTAuth`, qui exige
  l'en-tête **`X-Band-ID`**. Sans lui : `400 En-tête X-Band-ID manquant.` Les routes du
  dossier *Utilisateur* passent par `JWTAuthUserOnly` et s'en dispensent.
- **Rôles** : les requêtes marquées *(admin)* passent par `middleware.AdminOnly` et
  répondent `403` à un simple membre.
- **Erreurs** : format unique produit par `api/apierror` —
  `{ "error": { "code": "...", "message": "..." } }`.
- **Rate limiting** : `POST /api/auth/login` et `/api/auth/signup` sont limités quand
  `RATE_LIMIT_ENABLED` est actif.

⚠️ La collection attaque le backend **en direct** sur le port 8089. Le frontend, lui, passe
par son proxy SvelteKit (`/api/*` sur le port 4000, voir
`frontend/src/routes/api/[...slug]/+server.ts`) qui injecte `Authorization` et `X-Band-ID`
depuis les cookies — un comportement que Postman ne reproduit pas.

## Tenir la collection à jour

Toute PR qui ajoute, supprime ou modifie une route de `backend/main.go` met à jour ce
fichier. Pour vérifier qu'aucune route ne manque :

```bash
python3 - <<'PY'
import json, re
src = open('backend/main.go').read()
routes = {(m.group(1)) for m in re.finditer(r'mux\.Handle(?:Func)?\("([^"]+)"', src)}
col = json.load(open('backend/postman/setly.postman_collection.json'))
got = {r['request']['method'] + ' /' + '/'.join(r['request']['url']['path'])
       for f in col['item'] for r in f['item']}
norm = lambda s: re.sub(r'\{\{[^}]+\}\}|\{[^}]+\}', '{}', s)
missing = {r for r in routes if not any(norm(r) in norm(g) or norm(g).endswith(norm(r)) for g in got)}
print('routes:', len(routes), '| requêtes:', len(got), '| manquantes:', missing or 'aucune')
PY
```

## En ligne de commande

```bash
npx newman run backend/postman/setly.postman_collection.json
```

Utile pour un smoke test manuel contre une instance locale. Ce n'est pas branché sur la CI :
les tests d'API du projet sont les tests Go (`make test-backend`) et la suite Playwright.
