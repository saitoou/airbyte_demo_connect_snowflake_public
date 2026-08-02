select
  order_id,
  customer_id,
  order_status,
  total_amount,
  ordered_at,
  updated_at,
  cancelled_at,
  is_cancelled,

  case
    when is_cancelled then 0::number(12, 2)
    else total_amount
  end as recognized_order_amount

from {{ ref('stg_postgres__orders') }}
