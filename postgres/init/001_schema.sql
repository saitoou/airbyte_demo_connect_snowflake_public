create table customers (
    id bigint generated always as identity primary key,
    name varchar(100) not null,
    email varchar(255) not null unique,
    created_at timestamptz not null default current_timestamp,
    updated_at timestamptz not null default current_timestamp
);

create table orders (
    id bigint generated always as identity primary key,
    customer_id bigint not null references customers(id),
    status varchar(30) not null,
    total_amount numeric(12, 2) not null,
    ordered_at timestamptz not null default current_timestamp,
    updated_at timestamptz not null default current_timestamp
);

create user airbyte_reader
    with password 'airbyte-local-password';

grant connect
    on database airbyte_source_db
    to airbyte_reader;

grant usage
    on schema public
    to airbyte_reader;

grant select
    on all tables in schema public
    to airbyte_reader;

alter default privileges in schema public
    grant select on tables
    to airbyte_reader;