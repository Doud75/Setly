# Description

<!-- Que fait cette PR, et pourquoi ? Si le contexte est dans une issue, un renvoi suffit. -->

## Issue liée

<!-- Ex. : Closes #42 — ou « aucune » si la PR ne répond à aucune issue. -->

## Type de changement

- [ ] 🐛 Correction de bug (changement non cassant qui corrige un problème)
- [ ] ✨ Nouvelle fonctionnalité (changement non cassant qui ajoute une capacité)
- [ ] 💥 Changement cassant (correctif ou fonctionnalité qui modifie un comportement existant)
- [ ] 📖 Documentation uniquement
- [ ] 🧹 Refactorisation / maintenance (aucun changement de comportement)

## Comment tester

<!--
Les étapes qu'un relecteur doit suivre pour vérifier ton changement.
Précise les commandes lancées et, pour un changement visuel, joins une capture d'écran ou un avant/après.
-->

## Checklist

- [ ] `make test-all` passe en local
- [ ] `cd frontend && npm run lint` passe
- [ ] J'ai ajouté ou mis à jour les tests couvrant mon changement
- [ ] La documentation est à jour si mon changement le nécessite (README, `ARCHITECTURE.md`, `CONTRIBUTING.md`)
- [ ] `CHANGELOG.md` est mis à jour sous `[Unreleased]` si le changement est visible par les utilisateurs

### Si cette PR touche à la base de données

- [ ] Une migration a été créée avec `make migrate-create name=...`
- [ ] Le fichier `.down.sql` annule correctement la migration (testé dans les deux sens)
- [ ] Les données de seed (`backend/db/seed/`) ont été adaptées si nécessaire

## Notes pour la revue

<!-- Points d'attention, décisions discutables, parties dont tu n'es pas sûr. C'est le bon endroit pour être honnête. -->
