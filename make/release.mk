# make/release.mk - Publication d'une version (mainteneurs)
#
# Ces cibles mettent à jour le CHANGELOG, poussent un commit de release sur main
# et un tag vX.Y.Z, ce qui déclenche le déploiement en production (cd.yml).

release-patch:
	@./scripts/release.sh patch

release-minor:
	@./scripts/release.sh minor

release-major:
	@./scripts/release.sh major

# Affiche la prochaine version et le diff du CHANGELOG sans rien modifier.
release-dry-run:
	@./scripts/release.sh patch --dry-run

.PHONY: release-patch release-minor release-major release-dry-run
