create extension if not exists pgcrypto;

create table if not exists accounts (
    id uuid primary key default gen_random_uuid(),
    username text not null,
    password_encrypted text not null,
    login_url text not null,
    access_token_encrypted text not null,
    refresh_token_encrypted text not null,
    region text not null,
    account_type text not null check (account_type in ('claude', 'aws', 'gpt', 'kiro-aws', 'kiro-offical', 'claudecode', 'codex')),
    status text not null check (status in ('active', 'disabled', 'exhausted', 'login_failed', 'token_expired', 'region_blocked', 'error')),
    quota_total bigint not null default 0 check (quota_total >= 0),
    quota_used bigint not null default 0 check (quota_used >= 0),
    quota_remaining bigint not null default 0 check (quota_remaining >= 0),
    quota_reset_at timestamptz,
    max_concurrent_leases integer not null default 1 check (max_concurrent_leases > 0),
    tags text[] not null default '{}',
    metadata jsonb not null default '{}',
    notes text not null default '',
    kiro_expires_at timestamptz,
    kiro_profile_arn text not null default '',
    kiro_auth_method text not null default '',
    kiro_provider text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_accounts_status_region_type on accounts (status, region, account_type);
create index if not exists idx_accounts_quota_remaining on accounts (quota_remaining desc);
create index if not exists idx_accounts_tags on accounts using gin (tags);

create table if not exists api_callers (
    id uuid primary key default gen_random_uuid(),
    name text not null unique,
    api_key_hash text not null,
    status text not null check (status in ('active', 'disabled')),
    description text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_api_callers_status on api_callers (status);

create table if not exists account_leases (
    id uuid primary key default gen_random_uuid(),
    account_id uuid not null references accounts (id) on delete cascade,
    caller_id uuid not null references api_callers (id) on delete restrict,
    purpose text not null default '',
    request_filters jsonb not null default '{}',
    status text not null check (status in ('active', 'released', 'expired')),
    leased_at timestamptz not null default now(),
    expires_at timestamptz not null,
    released_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_account_leases_account_status on account_leases (account_id, status);
create index if not exists idx_account_leases_expires_status on account_leases (expires_at, status);
create index if not exists idx_account_leases_caller_status on account_leases (caller_id, status);

create table if not exists admin_users (
    id uuid primary key default gen_random_uuid(),
    username text not null unique,
    password_hash text not null,
    status text not null check (status in ('active', 'disabled')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

insert into admin_users (username, password_hash, status)
values ('admin', crypt('strongpass', gen_salt('bf')), 'active')
on conflict (username) do nothing;

create table if not exists admin_sessions (
    id uuid primary key default gen_random_uuid(),
    admin_user_id uuid not null references admin_users (id) on delete cascade,
    session_hash text not null unique,
    expires_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_admin_sessions_hash on admin_sessions (session_hash);
create index if not exists idx_admin_sessions_expires_at on admin_sessions (expires_at);

create table if not exists audit_logs (
    id uuid primary key default gen_random_uuid(),
    actor_type text not null check (actor_type in ('api_caller', 'admin')),
    actor_id uuid,
    action text not null,
    resource_type text not null,
    resource_id uuid,
    request_id text not null,
    ip_address text not null default '',
    user_agent text not null default '',
    metadata jsonb not null default '{}',
    created_at timestamptz not null default now()
);

create index if not exists idx_audit_logs_created_at on audit_logs (created_at desc);
create index if not exists idx_audit_logs_actor on audit_logs (actor_type, actor_id);
create index if not exists idx_audit_logs_action on audit_logs (action);
