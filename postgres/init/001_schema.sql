create table customers (
    id bigint generated always as identity primary key,
    name varchar(100) not null,
    email varchar(255) not null unique,
    created_at timestamptz not null default current_timestamp,
    updated_at timestamptz not null default current_timestamp,
    deleted_at timestamptz
);

create table orders (
    id bigint generated always as identity primary key,
    customer_id bigint not null references customers(id),
    status varchar(30) not null,
    total_amount numeric(12, 2) not null,
    ordered_at timestamptz not null default current_timestamp,
    updated_at timestamptz not null default current_timestamp,
    cancelled_at timestamptz,

    constraint orders_status_check
        check (
            status in (
                'pending',
                'paid',
                'shipped',
                'completed',
                'cancelled'
            )
        )
);

create table generated_batches (
    batch_id varchar(100) primary key,
    customer_count integer not null
        check (customer_count >= 0),
    order_count integer not null
        check (order_count >= 0),
    created_at timestamptz not null default current_timestamp
);

create user airbyte_reader
    with password 'test';

grant connect
    on database airbyte_source_db
    to airbyte_reader;

grant usage
    on schema public
    to airbyte_reader;

grant select
    on all tables in schema public
    to airbyte_reader;

alter default privileges
    for role app_user
    in schema public
    grant select on tables
    to airbyte_reader;
