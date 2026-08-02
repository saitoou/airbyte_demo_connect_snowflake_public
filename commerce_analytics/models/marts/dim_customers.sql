select
    customer_id,
    customer_name,
    email,
    created_at,
    updated_at,
    deleted_at,
    is_deleted,
    not is_deleted as is_active
from {{ ref('stg_postgres__customers') }}