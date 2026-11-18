---
name: setly-issue-to-pr
description: Take one GitHub issue of Doud75/Setly from "agent-ready" to an open pull request without human interaction — branch, BMAD build-auto (spec, implement, multi-layer review, commit), local verification, push, PR. Use when asked to "work on an issue", "take the next issue", "issue #N", or when run on a loop/schedule.
argument-hint: "[issue number] — omit to take the next agent-ready issue"
---

# Setly — issue → PR (one issue per run)

Runs **unattended**: never ask the user a question, never wait for approval. When something blocks, record it on the issue (step 7) and stop. One issue per invocation; never start a second one.

Rules from `CLAUDE.md` apply in full: work only on your own `feat|fix|docs|refactor/gh-<N>-<slug>` branch, never commit or push to `main`, never merge, never release, never run E2E.

Labels: `agent-ready` (queued by the user) → `agent-wip` (taken) → `agent-pr` (PR open) or `agent-blocked` (needs the user). `priority` (set by the user) puts an `agent-ready` issue ahead of the queue; the agent never adds or removes it.

## 1. Pick the issue

- Argument given → use that issue number. Refuse (stop with a one-line message, no label change) if it is closed, or already labelled `agent-wip` / `agent-pr`.
- No argument → `gh issue list --state open --label agent-ready --search "-label:agent-wip -label:agent-pr -label:agent-blocked" --json number,title,labels,createdAt --limit 50`. Among them, take the **oldest one labelled `priority`**; if none is labelled `priority`, take the **oldest** one. None at all → say "Aucune issue agent-ready." and stop (nothing else).
- Read it fully: `gh issue view <N> --json number,title,body,labels,comments`. Also read any issue it references as a dependency ("dépend de #X", "après #X", "bloqué par #X"): if such a dependency is still open and not merged, go to step 7 with reason `dépend de #X, encore ouverte`.
- Skip-and-block (step 7) an issue whose body asks for a product or design decision that is not made in the issue or its comments (e.g. "à définir", "à discuter", missing mockup for a `[Design]` issue that requires one). Implementation choices are yours; product choices are not.

## 2. Prepare the branch

1. `git status --porcelain` must be empty. If not, stop with `arbre de travail sale` — do not stash, reset or clean.
2. `git fetch origin && git switch main && git pull --ff-only origin main`.
3. Type: `fix` if labelled `bug`; `docs` if labelled `documentation` only; `refactor` if labelled `chore` and no behaviour change is asked; else `feat`.
4. Branch `<type>/gh-<N>-<slug>` (slug: 3–5 kebab-case words from the title, ASCII).
   - **Resume case**: if an earlier blocked run of this workflow pushed a branch for this issue (its name appears in one of the agent's comments on the issue, and it matches `*/gh-<N>-*`), `git switch <that branch>` (tracking origin), then `git merge --ff-only origin/<that branch>`, and reuse it. The user's latest comments are the answers to the blocking questions: include them in the intent at step 3.
   - Any other existing branch with that name (locally or on origin) → block with `branche déjà existante` — never reuse or overwrite it.
5. Otherwise `git switch -c <branch>`.
6. Mark the issue taken: `gh issue edit <N> --add-label agent-wip --remove-label agent-ready`.

## 3. Build with BMAD

Invoke the `bmad-build-auto` skill with this intent (fill it in verbatim — title, body and every comment, so the spec is grounded in the issue, not a paraphrase):

```
gh-<N> — <title>

Source: GitHub issue #<N> (Doud75/Setly). Labels: <labels>.

<issue body>

--- Commentaires ---
<author> (<date>): <comment body>
...
```

`bmad-build-auto` writes the spec to `_bmad-output/implementation-artifacts/spec-gh-<N>-*.md`, implements, reviews and **commits** on the current branch. It ends with a status:

- `done` → step 4.
- `blocked` (or anything else) → step 7 with its blocking condition, quoting the spec's `## Auto Run Result`.

If the spec's frontmatter has `followup_review_recommended: true`, invoke `bmad-build-auto` once more with the spec file path (a follow-up review pass). Do it at most once.

## 4. Verify locally

Run, from the repo root, only what the change touches:

- Backend touched → `cd backend && go build ./... && go vet ./...`, then `make test-backend`.
- Frontend touched → `cd frontend && npm run format && npm run lint && npm run check && npm run build`, then `make test-unit`.

On failure: fix it (up to **2** fix attempts in total), commit each fix (`fix : <what>` + the Co-Authored-By line). Still failing → step 7 with the failing command and the last 30 lines of output.

Never run E2E. If the change needs a new or updated Playwright spec, write it; CI runs it on the PR.

## 5. Finish the change

- `CHANGELOG.md`: if the change is visible to users and the spec did not already do it, add one French line under `## [Non publié]` in the matching subsection, and commit it (`docs : changelog #<N>`).
- `git status --porcelain` must be empty; `git log --oneline origin/main..HEAD` must show at least one commit. Otherwise step 7.

## 6. Push and open the PR

1. `git push -u origin <branch>`.
2. `gh pr create --base main --head <branch> --title "<type> : <issue title, without the [Design] prefix> (#<N>)" --body-file <tmp file>`, the body following `.github/PULL_REQUEST_TEMPLATE.md`:
   - **Description**: 3–6 lines, what and why, from the spec's summary.
   - **Issue liée**: `Closes #<N>`.
   - **Type de changement**: tick the one box that matches.
   - **Comment tester**: concrete steps for the reviewer (pages, actions, expected result) + the commands you ran with their outcome.
   - **Checklist**: tick only what is true. Leave `make test-all` unticked and write "E2E : laissé à la CI".
   - **Notes pour la revue**: the spec path, findings deferred or rejected by the BMAD review, residual risks, any assumption you made. Be honest.
   - Last line: `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.
   - Add the issue's `frontend` / `backend` / `a11y` / `security` labels to the PR.
3. `gh issue edit <N> --add-label agent-pr --remove-label agent-wip`.
4. `gh issue comment <N> --body "PR ouverte : <url> — en attente de relecture."`
5. Never merge, never enable auto-merge, never request a review from anyone.

Final message (and only output that matters): issue number, PR URL, one-line summary, verification results.

## 7. Blocked

1. `gh issue edit <N> --add-label agent-blocked --remove-label agent-wip --remove-label agent-ready`.
2. `gh issue comment <N>` in French: what was attempted, the exact blocking condition, the precise question(s) the user must answer to unblock it, and where partial work is (branch name if commits exist, spec path).
3. If the branch has commits, push it (`git push -u origin <branch>`) so nothing is lost — but open **no** PR.
4. `git switch main`. Leave the working tree clean; never discard work.

Unblocking is the user's job: they answer on the issue and swap `agent-blocked` back to `agent-ready`.
