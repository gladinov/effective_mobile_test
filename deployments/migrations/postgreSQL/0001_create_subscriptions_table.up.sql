
create extension if not exists "pgcrypto";

create table subscriptions (
    id uuid primary key default gen_random_uuid(),
    service_name text not null,
    price integer not null check (price > 0),
    user_id uuid not null,
    start_date date not null,
    end_date date
);
