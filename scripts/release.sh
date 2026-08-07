#!/usr/bin/env bash
#
# Publie une nouvelle version : bascule la section [Non publié] du CHANGELOG
# en version datée, commit sur main, pose un tag annoté et le pousse — ce qui
# déclenche .github/workflows/cd.yml, donc le déploiement en production.
#
# Usage : scripts/release.sh <patch|minor|major> [--dry-run]
#
set -euo pipefail

REPO_URL="https://github.com/Doud75/Setly"
CHANGELOG="CHANGELOG.md"
BRANCH="main"
UNRELEASED_HEADING="## [Non publié]"

die() {
	echo "❌ $*" >&2
	exit 1
}

# --- Arguments ---------------------------------------------------------------

BUMP="${1:-}"
DRY_RUN=false

case "$BUMP" in
patch | minor | major) ;;
*) die "Usage : scripts/release.sh <patch|minor|major> [--dry-run]" ;;
esac

if [ "${2:-}" = "--dry-run" ]; then
	DRY_RUN=true
elif [ -n "${2:-}" ]; then
	die "Argument inconnu : $2 (attendu : --dry-run)"
fi

cd "$(git rev-parse --show-toplevel)"

# --- Garde-fous --------------------------------------------------------------

current_branch="$(git rev-parse --abbrev-ref HEAD)"
[ "$current_branch" = "$BRANCH" ] ||
	die "Tu es sur '$current_branch'. Une release se fait depuis '$BRANCH'."

[ -z "$(git status --porcelain)" ] ||
	die "L'arbre de travail n'est pas propre. Commite ou remise tes changements d'abord."

echo "🔄 Récupération de l'état distant..."
git fetch --quiet origin "$BRANCH" --tags

local_head="$(git rev-parse HEAD)"
remote_head="$(git rev-parse "origin/$BRANCH")"
[ "$local_head" = "$remote_head" ] ||
	die "'$BRANCH' local diffère de 'origin/$BRANCH'. Fais un 'git pull' (ou push) avant de publier."

[ -f "$CHANGELOG" ] || die "$CHANGELOG introuvable."
grep -qxF "$UNRELEASED_HEADING" "$CHANGELOG" ||
	die "Section '$UNRELEASED_HEADING' introuvable dans $CHANGELOG."

unreleased="$(awk -v heading="$UNRELEASED_HEADING" '
	$0 == heading { inside = 1; next }
	inside && /^## / { exit }
	inside { print }
' "$CHANGELOG")"

if [ -z "$(printf '%s' "$unreleased" | tr -d '[:space:]')" ]; then
	die "La section '$UNRELEASED_HEADING' est vide : il n'y a rien à publier.
   Documente les changements dans $CHANGELOG avant de lancer une release."
fi

# --- Calcul de la version ----------------------------------------------------

last_tag="$(git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || true)"

if [ -n "$last_tag" ]; then
	IFS='.' read -r major minor patch <<<"${last_tag#v}"
else
	major=0
	minor=0
	patch=0
fi

case "$BUMP" in
patch) patch=$((patch + 1)) ;;
minor)
	minor=$((minor + 1))
	patch=0
	;;
major)
	major=$((major + 1))
	minor=0
	patch=0
	;;
esac

version="${major}.${minor}.${patch}"
tag="v${version}"
today="$(date +%Y-%m-%d)"

if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
	die "Le tag $tag existe déjà en local."
fi
if git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null 2>&1; then
	die "Le tag $tag existe déjà sur origin."
fi

# --- Réécriture du CHANGELOG -------------------------------------------------

if [ -n "$last_tag" ]; then
	version_link="[${version}]: ${REPO_URL}/compare/${last_tag}...${tag}"
else
	version_link="[${version}]: ${REPO_URL}/releases/tag/${tag}"
fi
unreleased_link="[Non publié]: ${REPO_URL}/compare/${tag}...HEAD"

tmp_changelog="$(mktemp)"
trap 'rm -f "$tmp_changelog"' EXIT

awk \
	-v heading="$UNRELEASED_HEADING" \
	-v version="$version" \
	-v today="$today" \
	-v unreleased_link="$unreleased_link" \
	-v version_link="$version_link" '
	$0 == heading && !heading_done {
		print
		print ""
		print "## [" version "] - " today
		heading_done = 1
		next
	}
	/^\[Non publié\]: / && !link_done {
		print unreleased_link
		print version_link
		link_done = 1
		next
	}
	{ print }
	END {
		if (!link_done) {
			print ""
			print unreleased_link
			print version_link
		}
	}
' "$CHANGELOG" >"$tmp_changelog"

# --- Récapitulatif -----------------------------------------------------------

echo
echo "📦 Version actuelle : ${last_tag:-aucune}"
echo "🏷️  Nouvelle version : $tag  ($BUMP)"
echo
echo "--- Contenu publié -------------------------------------------------------"
printf '%s\n' "$unreleased" | sed '/^[[:space:]]*$/d'
echo "--------------------------------------------------------------------------"
echo

if [ "$DRY_RUN" = true ]; then
	echo "🔍 Mode --dry-run : diff du CHANGELOG, aucun fichier modifié."
	echo
	diff -u "$CHANGELOG" "$tmp_changelog" || true
	exit 0
fi

echo "Cette commande va :"
echo "  1. mettre à jour $CHANGELOG"
echo "  2. commiter et pousser sur '$BRANCH'"
echo "  3. pousser le tag $tag → déclenche le DÉPLOIEMENT EN PRODUCTION"
echo
[ -t 0 ] || die "Confirmation impossible : entrée non interactive."
read -r -p "Publier $tag ? [y/N] " reply
case "$reply" in
y | Y | yes | YES) ;;
*) die "Annulé." ;;
esac

# --- Publication -------------------------------------------------------------

cp "$tmp_changelog" "$CHANGELOG"

git add "$CHANGELOG"
git commit -m "chore : release $tag"

echo "⬆️  Push du commit de release sur '$BRANCH'..."
if ! git push origin "$BRANCH"; then
	die "Le push sur '$BRANCH' a échoué (quelqu'un a probablement poussé entre-temps).
   Aucun tag n'a été créé. Pour annuler le commit local :
     git reset --hard origin/$BRANCH
   Puis relance la release une fois à jour."
fi

git tag -a "$tag" -m "$tag"

echo "⬆️  Push du tag $tag..."
git push origin "$tag"

echo
echo "✅ $tag publié. Le déploiement est en cours :"
echo "   ${REPO_URL}/actions"
