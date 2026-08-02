insert into customers (name, email)
values
    ('Sato Taro', 'taro@example.com'),
    ('Suzuki Hanako', 'hanako@example.com'),
    ('Takahashi Jiro', 'jiro@example.com');

insert into orders (
    customer_id,
    status,
    total_amount
)
values
    (1, 'paid', 3200.00),
    (1, 'pending', 1800.00),
    (2, 'paid', 5400.00);