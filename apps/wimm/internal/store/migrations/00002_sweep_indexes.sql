-- +goose NO TRANSACTION
-- The sweep is the first query to read sessions and enrolment tickets by time.
-- Nothing else scans either table that way, so these indexes arrive with it
-- rather than having been there all along; ceremony_challenges already has the
-- one it needs.
--
-- Built concurrently, and therefore outside a transaction, because the
-- instances this lands on have been accumulating these exact rows for as long
-- as they have been running: an index build that blocks writes to sessions
-- blocks every signed-in request.
--
-- Each index is dropped before it is created. A concurrent build that is
-- interrupted leaves an invalid index behind, and re-running the migration has
-- to replace it rather than see a name already taken and move on.
--
-- No statement_timeout: a concurrent build takes as long as the table is
-- large, and holds no lock anything is queued behind while it does.

-- +goose Up
set lock_timeout = '5s';

drop index concurrently if exists enrolment_tickets_expires_at_idx;
create index concurrently enrolment_tickets_expires_at_idx
    on enrolment_tickets (expires_at);

drop index concurrently if exists sessions_expires_at_idx;
create index concurrently sessions_expires_at_idx
    on sessions (expires_at);

-- A revoked session is removed by how long ago it was revoked, not by when it
-- would have expired, so that branch of the delete scans its own column.
-- Partial because the overwhelming majority of rows have never been revoked.
drop index concurrently if exists sessions_revoked_at_idx;
create index concurrently sessions_revoked_at_idx
    on sessions (revoked_at) where revoked_at is not null;

-- +goose Down
set lock_timeout = '5s';

drop index concurrently if exists sessions_revoked_at_idx;
drop index concurrently if exists sessions_expires_at_idx;
drop index concurrently if exists enrolment_tickets_expires_at_idx;
