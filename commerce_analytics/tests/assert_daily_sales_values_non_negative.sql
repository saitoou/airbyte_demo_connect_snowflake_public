select
  order_date,
  order_count,
  recognized_order_amount
from {{ ref('mart_daily_sales') }}
where
  order_count < 0
  or recognized_order_amount < 0
