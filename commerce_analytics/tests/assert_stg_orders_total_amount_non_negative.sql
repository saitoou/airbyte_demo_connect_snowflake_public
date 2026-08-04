select
    order_id,
    total_amount
from {{ ref('stg_postgres__orders') }}
where total_amount < 0