-- Тестовые данные для таблицы delivery
INSERT INTO delivery (name, phone, zip, city, address, region, email) VALUES
('Иванов Иван Иванович', '+79123456789', '123456', 'Москва', 'ул. Ленина, д. 10, кв. 5', 'Московская область', 'ivanov@example.com'),
('Петрова Мария Сергеевна', '+79098765432', '654321', 'Санкт-Петербург', 'пр. Невский, д. 25, кв. 12', 'Ленинградская область', 'petrova@example.com'),
('Сидоров Алексей Владимирович', '+79998887766', '111222', 'Екатеринбург', 'ул. Мира, д. 15, кв. 8', 'Свердловская область', 'sidorov@example.com');

-- Тестовые данные для таблицы payment
INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES
('txn_001', 'req_001', 'RUB', 'Visa', 15000, 1698765432, 'Сбербанк', 500, 14500, 0),
('txn_002', 'req_002', 'RUB', 'MasterCard', 25000, 1698765500, 'ВТБ', 700, 24300, 0),
('txn_003', 'req_003', 'RUB', 'Mir', 18000, 1698765600, 'Альфа-Банк', 600, 17400, 0);

-- Тестовые данные для таблицы orders
INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_id, payment_id) VALUES
('order_001', 'track_001', 'entry_1', 'ru', 'sig_001', 'cust_001', 'DHL', 'shard_1', 1001, '2023-10-31 10:30:00', 'oof_1', 1, 1),
('order_002', 'track_002', 'entry_2', 'ru', 'sig_002', 'cust_002', 'FedEx', 'shard_2', 1002, '2023-10-31 11:45:00', 'oof_2', 2, 2),
('order_003', 'track_003', 'entry_3', 'ru', 'sig_003', 'cust_003', 'Почта России', 'shard_3', 1003, '2023-10-31 12:15:00', 'oof_3', 3, 3);

-- Тестовые данные для таблицы item
INSERT INTO item (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES
('order_001', 10001, 'track_001', 5000, 'rid_001', 'Смартфон', 10, 'M', 4500, 1001, 'Samsung', 1),
('order_001', 10002, 'track_001', 3000, 'rid_002', 'Наушники', 0, 'S', 3000, 1002, 'Sony', 1),
('order_001', 10003, 'track_001', 7000, 'rid_003', 'Чехол', 5, 'L', 6650, 1003, 'Apple', 1),
('order_002', 10004, 'track_002', 15000, 'rid_004', 'Ноутбук', 15, 'XL', 12750, 1004, 'Lenovo', 1),
('order_002', 10005, 'track_002', 10000, 'rid_005', 'Монитор', 5, 'XXL', 9500, 1005, 'LG', 1),
('order_003', 10006, 'track_003', 8000, 'rid_006', 'Планшет', 0, 'M', 8000, 1006, 'Huawei', 1),
('order_003', 10007, 'track_003', 5000, 'rid_007', 'Клавиатура', 20, 'S', 4000, 1007, 'Logitech', 1),
('order_003', 10008, 'track_003', 3000, 'rid_008', 'Мышь', 10, 'S', 2700, 1008, 'Microsoft', 1),
('order_003', 10009, 'track_003', 2000, 'rid_009', 'Веб-камера', 0, 'M', 2000, 1009, 'Logitech', 1);