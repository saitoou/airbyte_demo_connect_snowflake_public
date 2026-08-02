select
  ordered_at::date as order_date,

  count(*) as order_count,

  count_if(not is_cancelled) as valid_order_count,

  count_if(is_cancelled) as cancelled_order_count,

  sum(total_amount)::number(14, 2) as gross_order_amount,

  sum(recognized_order_amount)::number(14, 2)
    as recognized_order_amount

from {{ ref('fct_orders') }}
group by
  ordered_at::date
