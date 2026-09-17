-- +goose Up
-- The banking schema: which banks a household has connected, which accounts
-- those connections expose, who owns each one and what each other member may
-- see of it. Owned by wimmd (ADR 0001, 0018, 0019).
--
-- Every lifetime is a timestamptz compared against now() in SQL, never against
-- a process clock.

-- One household's live read access to one bank. "Connection" rather than
-- "session": session already means a signed-in browser (ADR 0016), and
-- "expired session" would be ambiguous between signing in again and confirming
-- again at your bank.
create table bank_connections (
    id                 uuid        primary key default uuidv7(),
    -- Stored rather than compiled in, so a later gateway is new connections
    -- rather than a reinterpretation of old rows.
    gateway            text        not null check (length(btrim(gateway)) > 0),
    -- The gateway's session handle. It reads this household's accounts until
    -- the grant runs out, which makes it a live credential rather than an
    -- identifier: sealed with a key that lives outside the database, and null
    -- once access has ended.
    gateway_ref_sealed bytea,
    key_id             text,
    bank_id            text        not null,
    bank_name          text        not null,
    bank_logo_url      text,
    -- Whose consent is holding this open, and who will have to restore it.
    -- It confers no authority over who sees what (ADR 0019).
    connected_by       uuid        not null references members (id),
    created_at         timestamptz not null default now(),
    -- The date the bank granted, which is that bank's limit and not a wimm
    -- policy. It is 90 days at some banks and 1 day at others.
    consent_expires_at timestamptz not null,
    -- Set when access is known to have ended, whether at the date above or
    -- early. Both are the same thing to a member: their bank stopped updating.
    expired_at         timestamptz,
    disconnected_at    timestamptz,

    -- A sealed value and the key that opens it travel together or not at all.
    constraint bank_connections_sealed_with_key check (
        (gateway_ref_sealed is null) = (key_id is null)
    )
);

create index bank_connections_connected_by_idx on bank_connections (connected_by);

-- An authorisation begun and not yet returned from. Separate from the table
-- above because most of these are abandoned: a member who closes the tab at
-- their bank leaves one behind, and it must not look like a connection.
create table pending_bank_connections (
    id           uuid        primary key default uuidv7(),
    -- wimm's own opaque value, returned in the callback. Hashed for the same
    -- reason an enrolment link is: it is a bearer value while it is live.
    state_hash   bytea       not null,
    gateway      text        not null,
    -- The gateway's handle for the in-flight authorisation. Not sealed: it
    -- opens nothing on its own and is useless once consumed.
    gateway_ref  text        not null,
    bank_id      text        not null,
    bank_name    text        not null,
    redirect_url text        not null,
    -- Set when this is a restore rather than a first connection, so the return
    -- re-enters the flow at the right place and carries owners forward.
    restores     uuid        references bank_connections (id) on delete cascade,
    started_by   uuid        not null references members (id) on delete cascade,
    created_at   timestamptz not null default now(),
    expires_at   timestamptz not null,
    -- Set when the return is exchanged. A consumed row is kept until the sweep
    -- removes it, so a replayed return is refused as spent rather than as
    -- never-issued: those are different answers and only one of them is true.
    consumed_at  timestamptz
);

create unique index pending_bank_connections_state_hash_key
    on pending_bank_connections (state_hash);
create index pending_bank_connections_started_by_idx
    on pending_bank_connections (started_by);

-- One account the household has. **Not necessarily one at a bank wimm can
-- reach**: an account exists on its own terms, and where its figures come from
-- is a separate fact recorded in the columns below.
--
-- This is why the table is `accounts` rather than `bank_accounts`, and why
-- connection_id is nullable. An account entered by hand, or held somewhere no
-- open-banking gateway reaches, is the same kind of thing as one Enable
-- Banking returns — it is owned, shared and totalled identically, and only its
-- source differs. Coupling the existence of an account to a gateway would make
-- every later account type a schema change rather than a row.
create table accounts (
    id                 uuid        primary key default uuidv7(),

    -- Where this account's figures come from. `gateway` is the only kind this
    -- change creates; `manual` is named because the distinction is the reason
    -- this column exists, and a check constraint is one ALTER TABLE to widen.
    source             text        not null default 'gateway'
                                   check (source in ('gateway', 'manual')),

    -- Null for an account no gateway sources. Set, it cascades: an account
    -- that exists only because a bank exposed it goes when that bank does.
    connection_id      uuid        references bank_connections (id) on delete cascade,
    -- The gateway's cross-session identity hash. It survives a restore, which
    -- the per-session uid does not, so it is what everything keys on: a schema
    -- keyed on the uid would lose every ownership and sharing choice the first
    -- time a member restored.
    gateway_ref        text,
    -- The per-session identifier the gateway wants back on a read. Rewritten
    -- on every restore, and sealed because it is part of a live credential.
    gateway_uid_sealed bytea,
    key_id             text,

    name               text,
    -- Enough trailing characters to tell two accounts at one bank apart. The
    -- full number is never stored: an IBAN at rest is a liability with no use
    -- here.
    number_suffix      text,
    account_type       text,
    holder_name        text,
    currency           text        not null check (length(currency) = 3),
    -- The last reading, and when it was taken. A balance without a read time
    -- is not a balance in this system, so the two are null together.
    balance_minor      bigint,
    balance_read_at    timestamptz,
    created_at         timestamptz not null default now(),

    constraint accounts_sealed_with_key check (
        (gateway_uid_sealed is null) = (key_id is null)
    ),
    constraint accounts_balance_has_a_read_time check (
        (balance_minor is null) = (balance_read_at is null)
    ),
    -- A gateway account has a connection and a reference; anything else has
    -- none of the gateway columns at all. Stated per branch rather than as one
    -- equality: the equality form accepts a manual account carrying a stray
    -- gateway_ref, because both sides come out false. Without this, "source"
    -- and the columns it describes are free to disagree, and the disagreement
    -- is discovered by a reader.
    constraint accounts_gateway_columns_match_source check (
        case source
            when 'gateway' then connection_id is not null and gateway_ref is not null
            else connection_id is null and gateway_ref is null
                 and gateway_uid_sealed is null and key_id is null
        end
    )
);

create index accounts_connection_id_idx on accounts (connection_id);

-- Unique per connection, and partial because an account with no connection has
-- no gateway reference to be unique about.
create unique index accounts_connection_gateway_ref_key
    on accounts (connection_id, gateway_ref)
    where connection_id is not null;

-- Who an account belongs to. Many-to-many: a joint account is owned by both
-- partners, which is the case this table exists for. An owner sees their
-- account in full, whatever anyone else has been granted (ADR 0019).
create table account_owners (
    account_id uuid        not null references accounts (id) on delete cascade,
    member_id  uuid        not null references members (id) on delete cascade,
    created_at timestamptz not null default now(),

    primary key (account_id, member_id)
);

-- Needed for "every account this member owns", across every source.
create index account_owners_member_id_idx on account_owners (member_id);

-- What one member may see of one account they do not own.
--
-- Hidden is the absence of a row rather than a value, so the common read —
-- what may this member see — is a join returning what exists rather than a
-- filter over what does not, and a member removed from the household takes
-- their visibility with them by cascade.
--
-- level is text with a check rather than a Postgres enum type: a fourth level
-- on an enum needs ALTER TYPE ... ADD VALUE, which cannot run inside a
-- transaction and so cannot share a migration with anything else.
--
-- There is deliberately no connection_id here. A grant is keyed on the account
-- and the member alone, so "every grant for this member, across everything the
-- household holds" is one query against one table — and it keeps working for
-- account kinds that have no connection at all.
create table account_grants (
    account_id uuid        not null references accounts (id) on delete cascade,
    member_id  uuid        not null references members (id) on delete cascade,
    level      text        not null check (level in ('balance', 'details')),
    granted_by uuid        not null references members (id),
    created_at timestamptz not null default now(),

    primary key (account_id, member_id)
);

create index account_grants_member_id_idx on account_grants (member_id);

-- +goose Down
drop table if exists account_grants;
drop table if exists account_owners;
drop table if exists accounts;
drop table if exists pending_bank_connections;
drop table if exists bank_connections;
