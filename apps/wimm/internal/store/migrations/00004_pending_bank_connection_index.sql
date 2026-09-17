-- +goose NO TRANSACTION
-- The sweep gains a fourth statement: abandoned authorisations. Most pending
-- connections are abandoned — a member who closes the tab at their bank leaves
-- one behind — so this table accumulates faster than the three the sweep
-- already scans, and it is scanned by time like all of them.
--
-- Built concurrently, and therefore outside a transaction, for the reason ADR
-- 0017 established: an index build that blocks writes blocks the flow a member
-- is standing in the middle of.
--
-- Dropped before it is created. An interrupted concurrent build leaves an
-- invalid index behind, and re-running has to replace it rather than see the
-- name taken and move on.

-- +goose Up
set lock_timeout = '5s';

drop index concurrently if exists pending_bank_connections_expires_at_idx;
create index concurrently pending_bank_connections_expires_at_idx
    on pending_bank_connections (expires_at);

-- +goose Down
set lock_timeout = '5s';

drop index concurrently if exists pending_bank_connections_expires_at_idx;
