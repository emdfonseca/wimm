-- +goose Up
-- An account always has an owner (ADR 0022). Zero owners stops being
-- representable: a deferred constraint trigger refuses it at commit, so no
-- route, handler or future caller can produce one. "Not in wimm" becomes its
-- own reversible fact, left_out_at, replacing the unowned-and-ungranted state
-- ADR 0019 designed. A household name is added alongside it because both are
-- one column each on accounts and belong in the same migration as the
-- invariant that makes the backfill below safe to leave in place.
--
-- One transaction, no concurrent index build: the tables here hold one row per
-- account, one per connection and one per transaction the household has ever
-- synced, all bounded by a single household's own history.

set lock_timeout = '5s';
set statement_timeout = '30s';

alter table accounts
    -- Null: the account is in wimm. A value: it is not, and when it stopped
    -- being. Clearing it brings the account back.
    add column left_out_at    timestamptz,
    -- The name an owner gave the account, shown to everybody who may see it.
    -- Absent from the gateway upsert's `do update set` list, which is the
    -- entire mechanism that keeps a reconnect from overwriting it.
    add column household_name text;

-- Backfill 1: every account that reaches this migration with no owner row is
-- an existing orphan (ADR 0019's unowned-and-ungranted state, reachable
-- before this migration and never again after it). Each is owned by its
-- connection's connected_by member and left out, so nothing that was
-- invisible becomes visible by surprise.
--
-- A manual account cannot be orphaned: there is no route that creates one
-- without an owner, only test fixtures, so every orphan this finds is a
-- gateway account and has a connection to read connected_by from.
create temporary table settle_account_orphans as
select a.id as account_id, c.connected_by as owner_id
from accounts a
join bank_connections c on c.id = a.connection_id
where not exists (select 1 from account_owners o where o.account_id = a.id);

insert into account_owners (account_id, member_id)
select account_id, owner_id from settle_account_orphans;

update accounts a
   set left_out_at = now()
  from settle_account_orphans o
 where a.id = o.account_id;

-- +goose StatementBegin
do $$
declare
    remaining int;
begin
    select count(*) into remaining
    from accounts a
    where not exists (select 1 from account_owners o where o.account_id = a.id);

    if remaining > 0 then
        raise exception
            'settle_account_orphans backfill left % account(s) with no owner', remaining;
    end if;
end $$;
-- +goose StatementEnd

drop table settle_account_orphans;

-- Backfill 2: a bank that returned no booking date left the column at the Go
-- zero time rather than null, because store.Transaction.BookingDate is a
-- time.Time and not a pointer. Fall back through the value date and then the
-- transaction date, the same order identity.go already uses for its digest.
update transactions
   set booking_date = coalesce(value_date, transaction_date)
 where booking_date = date '0001-01-01'
   and coalesce(value_date, transaction_date) is not null;

-- Nothing writes an actual null here today — the zero-dated sentinel above is
-- the only way this column has ever disagreed with "not null" — so this locks
-- the invariant in rather than changing behaviour.
alter table transactions
    alter column booking_date set not null;

-- The minimum-owner guard. Deferred because SetAccountOwners deletes every
-- owner row and inserts the new set inside one transaction: a non-deferred
-- trigger would fire in the gap and refuse a change that is legal at commit,
-- which is the only moment the question has a stable answer.
--
-- Each fires per row and checks the account still exists, because deleting an
-- account — or a connection, which cascades to its accounts — cascades to its
-- owner rows, and a trigger that did not look would refuse every
-- disconnection.
-- +goose StatementBegin
create function settle_account_has_an_owner() returns trigger as $$
declare
    target uuid;
begin
    -- OLD and NEW are typed to whichever table fired this trigger, and OLD is
    -- unbound on an INSERT: referencing old.account_id there raises "record
    -- \"old\" has no field \"account_id\"" before coalesce ever runs, so the
    -- two triggering tables are told apart by name rather than by which of
    -- the two rows happens to be set.
    if tg_table_name = 'accounts' then
        target := new.id;
    else
        target := old.account_id;
    end if;

    if not exists (select 1 from accounts where id = target) then
        return null;
    end if;
    if not exists (select 1 from account_owners where account_id = target) then
        raise exception 'account % would be left with no owner', target;
    end if;
    return null;
end;
$$ language plpgsql;
-- +goose StatementEnd

create constraint trigger account_owners_leaves_an_owner
    after delete or update on account_owners
    deferrable initially deferred
    for each row
    execute function settle_account_has_an_owner();

create constraint trigger accounts_are_created_with_an_owner
    after insert on accounts
    deferrable initially deferred
    for each row
    execute function settle_account_has_an_owner();

-- +goose Down
set lock_timeout = '5s';
set statement_timeout = '30s';

drop trigger if exists accounts_are_created_with_an_owner on accounts;
drop trigger if exists account_owners_leaves_an_owner on account_owners;
drop function if exists settle_account_has_an_owner();

alter table transactions
    alter column booking_date drop not null;

alter table accounts
    drop column if exists household_name,
    drop column if exists left_out_at;
