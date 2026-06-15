# make/tests.mk - Toutes les commandes de test

# Cible principale pour lancer tous les tests
test-all: test-unit test-backend test
	@echo "✅ All tests (Unit & E2E) finished successfully."

# ==============================================================================
# --- TESTS UNITAIRES (Vitest & Go) ---
# ==============================================================================

test-unit:
	@echo "--- Running Frontend Unit Tests ---"
	@cd frontend && npx vitest run
	@echo "✅ Frontend Unit Tests finished."

test-backend:
	@echo "--- Running Backend Unit Tests ---"
	@cd backend && go test -v ./...
	@echo "✅ Backend Unit Tests finished."

test-backend-cover:
	@echo "--- Running Backend Unit Tests with Coverage ---"
	@cd backend && go test -v -coverprofile=coverage.out ./...
	@cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Backend Unit Tests Coverage finished. Report available in backend/coverage.html"

govulncheck:
	@echo "--- Running govulncheck inside Docker (golang:1.25.11-bookworm) ---"
	@docker run --rm \
		-v $(CURDIR)/backend:/app \
		-w /app \
		golang:1.25.11-bookworm \
		sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./..."
	@echo "✅ govulncheck finished."

test-unit-watch:
	@echo "--- Running Unit Tests in watch mode ---"
	@cd frontend && npx vitest

# ==============================================================================
# --- TESTS END-TO-END (Playwright) ---
# ==============================================================================

# Cible principale pour tous les tests E2E
test: test-up run-playwright test-down
	@echo "✅ All E2E Tests finished. Report available in frontend/playwright-report/index.html"

# --- Cibles de test par suite logique ---

test-setlist: test-up run-playwright-setlist test-down
	@echo "✅ All Setlist tests finished."

test-song: test-up run-playwright-song test-down
	@echo "✅ All Song tests finished."

test-auth: test-up run-playwright-auth test-down
	@echo "✅ All Auth tests finished."

test-auth-login: test-up run-playwright-auth-login test-down
	@echo "✅ Auth login test finished."

test-auth-signup: test-up run-playwright-auth-signup test-down
	@echo "✅ Auth signup test finished."

test-invitation: test-up run-playwright-invitation test-down
	@echo "✅ All Invitation tests finished."

test-invitation-join: test-up run-playwright-invitation-join test-down
	@echo "✅ Invitation join test finished."

test-settings: test-up run-playwright-settings test-down
	@echo "✅ All Settings tests finished."

test-interlude: test-up run-playwright-interlude test-down
	@echo "✅ All Interlude tests finished."

test-multi-group: test-up run-playwright-multi-group test-down
	@echo "✅ Multi-group E2E test finished."

test-cross-band: test-up run-playwright-cross-band test-down
	@echo "✅ Cross-band security E2E test finished."

test-duplicate-band: test-up run-playwright-duplicate-band test-down
	@echo "✅ Duplicate band name E2E test finished."

test-dashboard: test-up run-playwright-dashboard test-down
	@echo "✅ Dashboard E2E test finished."

test-leave-band: test-up run-playwright-leave-band test-down
	@echo "✅ Leave band E2E test finished."

test-default-band: test-up run-playwright-default-band test-down
	@echo "✅ Default band E2E test finished."

test-offline: test-up run-playwright-offline test-down
	@echo "✅ Offline E2E tests finished."

# --- Cibles de test par fichier spécifique (granulaire) ---

# NOUVELLE CIBLE POUR LES ACTIONS D'ADMINISTRATION
test-setlist-actions: test-up run-playwright-setlist-actions test-down
	@echo "✅ Setlist admin actions test finished."

test-setlist-detail: test-up run-playwright-setlist-detail test-down
	@echo "✅ Setlist detail tests finished."

test-setlist-add: test-up run-playwright-setlist-add test-down
	@echo "✅ Setlist add tests finished."

test-setlist-new: test-up run-playwright-setlist-new test-down
	@echo "✅ Setlist new tests finished."

test-setlist-duplicate: test-up run-playwright-setlist-duplicate test-down
	@echo "✅ Setlist duplicate test finished."

test-song-list: test-up run-playwright-song-list test-down
	@echo "✅ Song list tests finished."

test-song-new: test-up run-playwright-song-new test-down
	@echo "✅ Song new tests finished."

test-song-edit: test-up run-playwright-song-edit test-down
	@echo "✅ Song edit tests finished."

test-song-detail: test-up run-playwright-song-detail test-down
	@echo "✅ Song detail tests finished."

test-settings-account: test-up run-playwright-settings-account test-down
	@echo "✅ Settings account test finished."

test-settings-members: test-up run-playwright-settings-members test-down
	@echo "✅ Settings members test finished."

test-interlude-new: test-up run-playwright-interlude-new test-down
	@echo "✅ Interlude new tests finished."

test-interlude-behavior: test-up run-playwright-interlude-behavior test-down
	@echo "✅ Interlude behavior tests finished."

# ==============================================================================
# --- CIBLES UTILITAIRES POUR LES TESTS E2E ---
# ==============================================================================

test-up:
	@echo "--- Cleaning up previous test environment ---"
	@docker compose -f docker-compose.test.yml --env-file .env.test down -v --remove-orphans
	@echo "--- Building and starting test environment (DB, Backend with seed, Frontend) ---"
	@docker compose -f docker-compose.test.yml --env-file .env.test up --build -d || \
        (echo "🔴 'docker compose up' failed. Displaying logs:"; \
        docker compose -f docker-compose.test.yml --env-file .env.test logs; \
        exit 1)
	@echo "--- Waiting for frontend to be healthy before running tests ---"
	@until curl -s -f http://localhost:4001 > /dev/null; do \
		echo "Waiting for frontend_test service on port 4001..."; \
		sleep 2; \
	done
	@echo "--- Test environment is ready ---"

test-down:
	@echo "--- Tearing down test environment ---"
	@docker compose -f docker-compose.test.yml --env-file .env.test down -v --remove-orphans

# --- Cibles d'exécution Playwright ---

run-playwright:
	@echo "--- Running ALL Playwright tests ---"
	@cd frontend && npx playwright test

# NOUVELLE CIBLE D'EXÉCUTION
run-playwright-setlist-actions:
	@echo "--- Running SETLIST ADMIN ACTIONS Playwright test ---"
	@cd frontend && npx playwright test tests/setlist/actions.spec.ts

run-playwright-setlist:
	@echo "--- Running SETLIST Playwright tests (directory) ---"
	@cd frontend && npx playwright test tests/setlist/

run-playwright-setlist-detail:
	@echo "--- Running SETLIST DETAIL Playwright tests ---"
	@cd frontend && npx playwright test tests/setlist/detail.spec.ts

run-playwright-setlist-add:
	@echo "--- Running SETLIST ADD Playwright tests ---"
	@cd frontend && npx playwright test tests/setlist/add.spec.ts

run-playwright-setlist-new:
	@echo "--- Running SETLIST NEW Playwright tests ---"
	@cd frontend && npx playwright test tests/setlist/new.spec.ts

run-playwright-setlist-duplicate:
	@echo "--- Running SETLIST DUPLICATE Playwright test ---"
	@cd frontend && npx playwright test tests/setlist/duplicate.spec.ts

run-playwright-song:
	@echo "--- Running SONG Playwright tests (directory) ---"
	@cd frontend && npx playwright test tests/song/

run-playwright-auth:
	@echo "--- Running AUTH Playwright tests (directory) ---"
	@cd frontend && npx playwright test tests/auth/

run-playwright-auth-login:
	@echo "--- Running AUTH LOGIN Playwright test ---"
	@cd frontend && npx playwright test tests/auth/login.spec.ts

run-playwright-song-list:
	@echo "--- Running SONG LIST Playwright tests ---"
	@cd frontend && npx playwright test tests/song/list.spec.ts

run-playwright-song-new:
	@echo "--- Running SONG NEW Playwright tests ---"
	@cd frontend && npx playwright test tests/song/new.spec.ts

run-playwright-song-edit:
	@echo "--- Running SONG EDIT Playwright tests ---"
	@cd frontend && npx playwright test tests/song/edit.spec.ts

run-playwright-song-detail:
	@echo "--- Running SONG DETAIL Playwright tests ---"
	@cd frontend && npx playwright test tests/song/detail.spec.ts

run-playwright-auth-signup:
	@echo "--- Running AUTH SIGNUP Playwright test ---"
	@cd frontend && npx playwright test tests/auth/signup.spec.ts

run-playwright-invitation:
	@echo "--- Running INVITATION Playwright tests (directory) ---"
	@cd frontend && npx playwright test tests/invitation/

run-playwright-invitation-join:
	@echo "--- Running INVITATION JOIN Playwright test ---"
	@cd frontend && npx playwright test tests/invitation/join.spec.ts

run-playwright-multi-group:
	@echo "--- Running MULTI-GROUP Playwright test ---"
	@cd frontend && npx playwright test tests/group/multi-group.spec.ts

run-playwright-cross-band:
	@echo "--- Running CROSS-BAND SECURITY Playwright test ---"
	@cd frontend && npx playwright test tests/group/cross-band-security.spec.ts

run-playwright-duplicate-band:
	@echo "--- Running DUPLICATE BAND NAME Playwright test ---"
	@cd frontend && npx playwright test tests/group/duplicate-band-name.spec.ts

run-playwright-dashboard:
	@echo "--- Running DASHBOARD Playwright test ---"
	@cd frontend && npx playwright test tests/group/dashboard.spec.ts

run-playwright-leave-band:
	@echo "--- Running LEAVE BAND Playwright test ---"
	@cd frontend && npx playwright test tests/group/leave-band.spec.ts

run-playwright-default-band:
	@echo "--- Running DEFAULT BAND Playwright test ---"
	@cd frontend && npx playwright test tests/group/default-band.spec.ts

run-playwright-settings:
	@echo "--- Running SETTINGS Playwright tests (directory) ---"
	@cd frontend && npx playwright test tests/settings/

run-playwright-settings-account:
	@echo "--- Running SETTINGS ACCOUNT Playwright test ---"
	@cd frontend && npx playwright test tests/settings/account.spec.ts

run-playwright-settings-members:
	@echo "--- Running SETTINGS MEMBERS Playwright test ---"
	@cd frontend && npx playwright test tests/settings/members.spec.ts

run-playwright-interlude:
	@echo "--- Running INTERLUDE Playwright tests (directory) ---"
	@cd frontend && npx playwright test tests/interlude/

run-playwright-interlude-new:
	@echo "--- Running INTERLUDE NEW Playwright tests ---"
	@cd frontend && npx playwright test tests/interlude/new.spec.ts

run-playwright-interlude-behavior:
	@echo "--- Running INTERLUDE BEHAVIOR Playwright tests ---"
	@cd frontend && npx playwright test tests/interlude/behavior.spec.ts

run-playwright-offline:
	@echo "--- Running OFFLINE Playwright tests ---"
	@cd frontend && npx playwright test tests/offline/

# --- Déclaration .PHONY pour toutes les cibles ---
.PHONY: test-all test-unit test-unit-watch test-backend test-backend-cover govulncheck \
		test test-setlist test-song test-settings test-interlude \
		test-multi-group test-cross-band test-duplicate-band test-dashboard test-leave-band test-default-band \
		test-auth test-auth-login test-auth-signup \
		test-invitation test-invitation-join \
		test-setlist-actions test-setlist-detail test-setlist-add test-setlist-new test-setlist-duplicate \
		test-song-list test-song-new test-song-edit test-song-detail \
		test-settings-account test-settings-members \
		test-interlude-new test-interlude-behavior \
		test-up test-down \
		run-playwright run-playwright-setlist-actions run-playwright-setlist run-playwright-setlist-detail \
		run-playwright-setlist-add run-playwright-setlist-new run-playwright-setlist-duplicate \
		run-playwright-song run-playwright-song-list run-playwright-song-new run-playwright-song-edit run-playwright-song-detail \
		run-playwright-multi-group run-playwright-cross-band run-playwright-duplicate-band run-playwright-dashboard run-playwright-leave-band run-playwright-default-band \
		run-playwright-settings run-playwright-settings-account run-playwright-settings-members \
		run-playwright-interlude run-playwright-interlude-new run-playwright-interlude-behavior \
		run-playwright-auth run-playwright-auth-login run-playwright-auth-signup \
		run-playwright-invitation run-playwright-invitation-join \
		test-offline run-playwright-offline