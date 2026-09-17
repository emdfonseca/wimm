## Skipped

No user story. This change removes rows nobody can reach: every lifetime is
already checked in SQL, so an expired challenge, session or ticket cannot be
used before the sweep and cannot be used after it. Nothing a member or the
operator does changes, and nothing they see changes.

Writing one would produce "as an operator I want the database not to grow",
which is the shape ADR 0012 names as the tell that there is no actor — the
operator does not do this, and cannot observe it except in a log line.

The behaviour that does need pinning is the system's own, and it is in
`specs/identity/expired-row-sweeping/spec.md`.
