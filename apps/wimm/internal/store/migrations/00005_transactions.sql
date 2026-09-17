-- +goose Up
-- The household's ledger, the consent that fills it, and the account columns
-- that let an account outlive the connection that found it (ADR 0021).
--
-- One transaction, and no concurrent index build: the two tables this touches
-- hold one row per account and one per bank a household has connected, and the
-- backfill below has to see the same rows the constraint after it checks. A
-- backfill is a job rather than a migration wherever the count is unbounded;
-- here it is bounded by the accounts of one household.

set lock_timeout = '5s';
set statement_timeout = '30s';

-- What a connection's consent covers, separate from whether it is working.
-- Every row that exists was granted balances and nothing else, which is what
-- the default states. Text with a check rather than an enum, for the reason
-- account_grants.level already is: a later value must not need ALTER TYPE.
alter table bank_connections
    add column scope text not null default 'balances'
        check (scope in ('balances', 'balances_and_transactions'));

alter table accounts
    -- Which bank an account is at. It lived on the connection while an account
    -- died with one; the account is now the permanent thing, so the permanent
    -- fact moves onto it.
    add column bank_id                     text,
    -- The date this account's transactions are believed complete through, and
    -- when that was last established. wimm's own facts, written after a sync
    -- finishes — including a partial one, so the next sync continues rather
    -- than starting the history again.
    add column transactions_synced_through date,
    add column transactions_synced_at      timestamptz;

-- Every existing gateway account has a connection, and that connection knows
-- its bank. Done here rather than in the application so the constraint below
-- can be added in the same transaction.
update accounts a
   set bank_id = c.bank_id
  from bank_connections c
 where a.connection_id = c.id
   and a.bank_id is null;

-- A gateway account without a bank cannot be re-attached after a disconnection,
-- because its bank is the half of its identity that survives. Manual accounts
-- are left free: an account held somewhere no gateway reaches may still name
-- the bank it is at.
alter table accounts
    add constraint accounts_gateway_account_has_a_bank check (
        source <> 'gateway' or bank_id is not null
    );

-- An account is identified by its bank and the gateway's cross-session hash,
-- for as long as it exists — across connections, not within one. This is what
-- makes re-attach have nothing to disambiguate: there is at most one row to
-- find. Partial, because an account no gateway sources has no reference to be
-- unique about.
create unique index accounts_bank_gateway_ref_key
    on accounts (bank_id, gateway_ref)
    where gateway_ref is not null;

-- One entry on one account.
--
-- On the account, never on the connection: an account survives a restore, a
-- disconnection and a reconnection, and its history survives with it. No bank
-- hands history back indefinitely and one of them grants access for a single
-- day, so a ledger keyed on the connection would empty every night.
create table transactions (
    id                uuid        primary key default uuidv7(),
    account_id        uuid        not null references accounts (id) on delete cascade,

    -- Booked is settled at the bank and append-only. Pending is not settled
    -- and is a replaceable set: each sync discards an account's pending rows
    -- and writes the ones the bank just returned, because a pending entry
    -- becomes a booked one under a different reference, amount and date, and
    -- matching the two is guesswork wrong in exactly the cases that matter.
    status            text        not null check (status in ('booked', 'pending')),

    -- The bank's own entry reference where it gives one, else a digest over
    -- booking date, amount, currency, counterparty and remittance. Never null:
    -- the fallback always produces a value.
    dedup_key         text        not null,
    -- Which of several rows sharing a key this is, so the same payment made
    -- twice is two transactions and the same transaction read twice is one.
    occurrence        int         not null default 1 check (occurrence >= 1),

    -- Signed minor units. A debit is negative, so direction never depends on a
    -- separate column agreeing with the sign.
    amount_minor      bigint      not null,
    currency          text        not null check (length(currency) = 3),

    booking_date      date,
    value_date        date,
    transaction_date  date,
    counterparty_name text,
    remittance        text,

    first_seen_at     timestamptz not null default now(),
    -- Written on every sync that returns this row, so a later change can find
    -- the ones a bank stopped returning.
    last_seen_at      timestamptz not null default now()
);

-- Booked rows only. A pending row has no stable identity across syncs, so
-- there is nothing about it to be unique on.
create unique index transactions_booked_key
    on transactions (account_id, dedup_key, occurrence)
    where status = 'booked';

-- The screen's only ordering, and the account filter's index. The ledger seeks
-- on this key in both directions rather than counting an offset, because a
-- sync inserts at the newest end while a member is reading.
create index transactions_account_date_idx
    on transactions (account_id, booking_date desc, id desc);

-- +goose Down
set lock_timeout = '5s';
set statement_timeout = '30s';

drop table if exists transactions;

drop index if exists accounts_bank_gateway_ref_key;

alter table accounts
    drop constraint if exists accounts_gateway_account_has_a_bank;

alter table accounts
    drop column if exists transactions_synced_at,
    drop column if exists transactions_synced_through,
    drop column if exists bank_id;

alter table bank_connections
    drop column if exists scope;
