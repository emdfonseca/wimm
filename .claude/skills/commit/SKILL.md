---
name: commit
description: Stage one logical change and write its conventional commit. Invoke with /commit; it never runs on its own.
disable-model-invocation: true
allowed-tools: Bash(git *) Bash(just check *)
---

# Commit

1. `git diff --cached --stat` first. Nothing staged → stage the files for one logical change explicitly by path. Never `git add -A` or `git add .`.
2. One logical change per commit. Two changes → two commits.
3. `just check <dir>` passes for every touched `apps/*` or `packages/*` dir. Red never gets committed.
4. Generated files (`gen/`) are committed together with the input that produced them, never in a separate commit and never edited by hand.
5. Message:
   - `type(scope): subject` — type in feat, fix, refactor, test, docs, chore, build, ci; scope = the app or package dir name.
   - Subject imperative, no period, ≤ 72 chars.
   - Body says why, not what. Omit it when the subject is enough.
   - End with the attribution trailer only if the session provides one.
6. `git commit` with the message via heredoc; report the resulting `git log -1 --oneline`.
