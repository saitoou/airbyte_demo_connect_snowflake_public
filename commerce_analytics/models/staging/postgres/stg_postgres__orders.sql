with source as (

  select *
  from {{ source('postgres', 'orders') }}

),

renamed as (

  select
    id::number(38, 0) as order_id,
    customer_id::number(38, 0) as customer_id,

    status::varchar as order_status,
    total_amount::number(12, 2) as total_amount,

    ordered_at::timestamp_tz as ordered_at,
    updated_at::timestamp_tz as updated_at,
    cancelled_at::timestamp_tz as cancelled_at,

    _airbyte_extracted_at::timestamp_tz as airbyte_extracted_at,

    status = 'cancelled' as is_cancelled

  from source

)

select *
from renamed
