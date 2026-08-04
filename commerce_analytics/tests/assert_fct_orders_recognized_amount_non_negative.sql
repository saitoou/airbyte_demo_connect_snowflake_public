select
  order_id,
  recognized_order_amount
from {{ ref('fct_orders') }}
where recognized_order_amount < 0
