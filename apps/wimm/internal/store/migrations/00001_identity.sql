-- +goose Up
-- The identity schema: who a member is, how they prove it, and what makes a
-- browser signed in. Owned by wimmd (ADR 0001, 0016).
--
-- Every lifetime is a timestamptz compared against now() in SQL, never against
-- a process clock, so a skewed instance cannot extend a link or a session.

create table members (
    id          uuid        primary key default uuidv7(),
    email       text        not null,
    first_name  text        not null check (length(btrim(first_name)) > 0),
    last_name   text        not null check (length(btrim(last_name)) > 0),
    created_at  timestamptz not null default now()
);

-- The operator asserts the address; nothing verifies it. Case folding is the
-- only normalisation, so Ada@example.com and ada@example.com are one person.
create unique index members_email_key on members (lower(email));

-- An enrolment link is a bearer credential: whoever holds the value becomes
-- that member. Only its hash is stored, so a copy of this table yields no
-- working links.
create table enrolment_links (
    id             uuid        primary key default uuidv7(),
    member_id      uuid        not null references members (id) on delete cascade,
    value_hash     bytea       not null,
    created_at     timestamptz not null default now(),
    expires_at     timestamptz not null,
    -- Set when the value is first exchanged for a ticket, which is what stops
    -- it travelling in the URL.
    redeemed_at    timestamptz,
    -- Set when a passkey is actually enrolled through it. This is what makes
    -- the link single-use; a refused attempt leaves it null.
    enrolled_at    timestamptz,
    -- Set when a later link is issued to the same member.
    invalidated_at timestamptz
);

create unique index enrolment_links_value_hash_key on enrolment_links (value_hash);
create index enrolment_links_member_id_idx on enrolment_links (member_id);

-- At most one live link per member. A link is live while it has not been used
-- to enrol and has not been superseded; expiry is a time comparison, so it
-- cannot be part of the index predicate.
create unique index enrolment_links_one_live_per_member
    on enrolment_links (member_id)
    where enrolled_at is null and invalidated_at is null;

-- What RedeemEnrolmentLink hands back. Short-lived and opaque, carried in an
-- HttpOnly cookie rather than a path, and hashed for the same reason the link
-- is.
create table enrolment_tickets (
    id                uuid        primary key default uuidv7(),
    enrolment_link_id uuid        not null references enrolment_links (id) on delete cascade,
    member_id         uuid        not null references members (id) on delete cascade,
    value_hash        bytea       not null,
    created_at        timestamptz not null default now(),
    expires_at        timestamptz not null
);

create unique index enrolment_tickets_value_hash_key on enrolment_tickets (value_hash);

-- A passkey. "Credential" here because the WebAuthn vocabulary is clearer in
-- code; never in anything a member reads.
--
-- The library's own credential record is stored whole, in data: it carries the
-- public key, the attestation, the authenticator flags and the signature
-- counter, and every field has to be handed back on validation. Splitting it
-- into columns would mean re-deriving a shape the library owns.
create table credentials (
    id            uuid        primary key default uuidv7(),
    member_id     uuid        not null references members (id) on delete cascade,
    credential_id bytea       not null,
    data          jsonb       not null,
    created_at    timestamptz not null default now(),
    last_used_at  timestamptz
);

-- A credential id is unique across the whole relying party, not per member.
create unique index credentials_credential_id_key on credentials (credential_id);
create index credentials_member_id_idx on credentials (member_id);

create type ceremony_kind as enum ('registration', 'authentication');

-- WebAuthn ceremony state. A row with a lifetime in seconds, deleted on use,
-- so wimmd holds no request-spanning state and needs no sticky routing.
create table ceremony_challenges (
    id            uuid          primary key default uuidv7(),
    kind          ceremony_kind not null,
    -- Null for authentication: the credential is discoverable, so the member
    -- is not known until the assertion comes back.
    member_id     uuid          references members (id) on delete cascade,
    -- The link a registration ceremony is enrolling against, so finishing it
    -- closes that link and nothing else.
    enrolment_link_id uuid      references enrolment_links (id) on delete cascade,
    -- The library's own session data for the ceremony, stored verbatim.
    session_data  jsonb         not null,
    -- Where the member was heading when they were sent to sign in. Held here
    -- rather than in the URL, so there is nothing to tamper with.
    intended_path text,
    created_at    timestamptz   not null default now(),
    expires_at    timestamptz   not null
);

create index ceremony_challenges_expires_at_idx on ceremony_challenges (expires_at);

-- What makes a browser signed in: an opaque identifier naming this row.
-- Revocable, which is the entire recovery story for a lost device.
create table sessions (
    id           uuid        primary key default uuidv7(),
    member_id    uuid        not null references members (id) on delete cascade,
    value_hash   bytea       not null,
    created_at   timestamptz not null default now(),
    last_seen_at timestamptz not null default now(),
    expires_at   timestamptz not null,
    revoked_at   timestamptz
);

create unique index sessions_value_hash_key on sessions (value_hash);
create index sessions_member_id_idx on sessions (member_id);

-- +goose Down
drop table sessions;
drop table ceremony_challenges;
drop type ceremony_kind;
drop table credentials;
drop table enrolment_tickets;
drop table enrolment_links;
drop table members;
