-- confirm ci test
with source as (

  select *
  from {{ source('postgres', 'customers') }}

),

renamed as (

  select
    id::number(38, 0) as customer_id,
    name::varchar as customer_name,
    email::varchar as email,

    created_at::timestamp_tz as created_at,
    updated_at::timestamp_tz as updated_at,
    deleted_at::timestamp_tz as deleted_at,

    _airbyte_extracted_at::timestamp_tz as airbyte_extracted_at,

    deleted_at is not null as is_deleted

  from source

)

select *
from renamed
