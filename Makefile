-include .env
export

.DEFAULT_GOAL := help

include make/*.mk

help:
	@echo "Usage: make [cible]"
	@echo ""
	@echo "Cibles principales:"
	@echo "  up             Démarre tous les services Docker."
	@echo "  down           Arrête tous les services Docker."
	@echo "  logs           Affiche les logs des services."
	@echo "  test-all       Lance tous les tests (unitaires et E2E)."
	@echo "  test           Lance uniquement les tests E2E."
	@echo "  test-unit      Lance uniquement les tests unitaires."
	@echo ""
	@echo "Publication d'une version (mainteneurs) :"
	@echo "  release-dry-run  Affiche la prochaine version et le CHANGELOG, sans rien modifier."
	@echo "  release-patch    Publie un correctif (0.1.0 -> 0.1.1) et deploie en production."
	@echo "  release-minor    Publie une fonctionnalite (0.1.1 -> 0.2.0) et deploie en production."
	@echo "  release-major    Publie un changement cassant (0.2.0 -> 1.0.0) et deploie en production."
	@echo ""
	@echo "Consultez les fichiers dans le dossier 'make/' pour voir toutes les cibles disponibles."

.PHONY: help