package storage

import (
	"database/sql"
	"log"
	"order-service/internal/models"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) SaveOrder(order models.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Сохранение Delivery
	var deliveryID int
	err = tx.QueryRow(`
        INSERT INTO delivery (name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING delivery_id`,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	).Scan(&deliveryID)
	if err != nil {
		return err
	}

	// Сохранение Payment
	var paymentID int
	err = tx.QueryRow(`
        INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING payment_id`,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	).Scan(&paymentID)
	if err != nil {
		return err
	}

	// Сохранение Order
	_, err = tx.Exec(`
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature, customer_id,
            delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_id, payment_id
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
		deliveryID,
		paymentID,
	)
	if err != nil {
		return err
	}

	// сохранение Items
	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO item (
                order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) GetOrderByUID(orderUID string) (models.Order, error) {
	var order models.Order
	var deliveryID, paymentID int

	err := s.db.QueryRow(`
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
               delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_id, payment_id
        FROM orders
        WHERE order_uid = $1`, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
		&deliveryID,
		&paymentID,
	)
	if err == sql.ErrNoRows {
		return models.Order{}, err
	}
	if err != nil {
		return models.Order{}, err
	}

	// получение Delivery
	err = s.db.QueryRow(`
        SELECT name, phone, zip, city, address, region, email
        FROM delivery
        WHERE delivery_id = $1`, deliveryID).Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return models.Order{}, err
	}

	// получение Payment
	err = s.db.QueryRow(`
        SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM payment
        WHERE payment_id = $1`, paymentID).Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDt,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return models.Order{}, err
	}

	// получение Items
	rows, err := s.db.Query(`
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM item
        WHERE order_uid = $1`, orderUID)
	if err != nil {
		return models.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		); err != nil {
			return models.Order{}, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (s *PostgresStorage) GetAllOrders() ([]models.Order, error) {
	rows, err := s.db.Query("SELECT order_uid FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderUIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		orderUIDs = append(orderUIDs, uid)
	}

	var orders []models.Order
	for _, uid := range orderUIDs {
		order, err := s.GetOrderByUID(uid)
		if err != nil {
			log.Printf("Failed to load order %s: %v", uid, err)
			continue
		}
		orders = append(orders, order)
	}
	return orders, nil
}
