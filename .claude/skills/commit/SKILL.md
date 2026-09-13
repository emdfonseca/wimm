---
name: commit
description: Write a conventional commit for staged or specified changes. Use when the user asks to commit, or when /ship passes and the user says commit. Never commits on its own.
disable-model-invocation: false
allowed-tools: Bash(git *)
---

# Commit

1. `git diff --cached --stat` first. Nothing staged → stage the files for one logical change explicitly by path. Never `git add -A` or `git add .`.
2. One logical change per commit. Two changes → two commits.
3. Never hand-stage generated files; `just gen` produces them and they are committed with their input.
4. Message:
   - `type(scope): subject` — type in feat, fix, refactor, test, docs, chore, build, ci; scope = the app or package dir name.
   - Subject imperative, no period, ≤ 72 chars.
   - Body says why, not what. Omit it when the subject is enough.
   - End with the attribution trailer only if the session provides one.
5. `git commit` with the message via heredoc; report the resulting `git log -1 --oneline`.
