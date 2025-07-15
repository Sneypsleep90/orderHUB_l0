package service

import (
	"context"
	"go.uber.org/zap"
	"orderHub_L0/internal/cache"
	"orderHub_L0/internal/db"
	"orderHub_L0/internal/model"
	"time"
)

type Service struct {
	db     *db.DB
	cache  *cache.Cache
	logger *zap.Logger
}

func NewService(db *db.DB, cache *cache.Cache, logger *zap.Logger) *Service {
	return &Service{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (s *Service) LoadCache(ctx context.Context) error {
	type orderRow struct {
		OrderUID        string    `db:"order_uid"`
		TrackNumber     string    `db:"track_number"`
		Entry           string    `db:"entry"`
		Locale          string    `db:"locale"`
		CustomerID      string    `db:"customer_id"`
		DeliveryService string    `db:"delivery_service"`
		ShardKey        string    `db:"shardkey"`
		SmID            int       `db:"sm_id"`
		DateCreated     time.Time `db:"date_created"`
		OofShard        string    `db:"oof_shard"`

		DeliveryName string `db:"delivery_name"`
		Phone        string `db:"phone"`
		Zip          string `db:"zip"`
		City         string `db:"city"`
		Address      string `db:"address"`
		Region       string `db:"region"`
		Email        string `db:"email"`

		Transaction  string `db:"transaction"`
		Currency     string `db:"currency"`
		Provider     string `db:"provider"`
		Amount       int    `db:"amount"`
		PaymentDt    int64  `db:"payment_dt"`
		Bank         string `db:"bank"`
		DeliveryCost int    `db:"delivery_cost"`
		GoodsTotal   int    `db:"goods_total"`
		CustomFee    int    `db:"custom_fee"`
	}

	var rows []orderRow
	query := `
		SELECT 
			o.order_uid, 
			o.track_number, 
			o.entry, 
			o.locale, 
			o.customer_id, 
			o.delivery_service, 
			o.shardkey, 
			o.sm_id, 
			o.date_created, 
			o.oof_shard,
			d.name AS delivery_name, 
			d.phone AS phone, 
			d.zip AS zip, 
			d.city AS city, 
			d.address AS address, 
			d.region AS region, 
			d.email AS email,
			p.transaction, 
			p.currency, 
			p.provider, 
			p.amount, 
			p.payment_dt, 
			p.bank, 
			p.delivery_cost, 
			p.goods_total, 
			p.custom_fee
		FROM orders o
		JOIN deliveries d ON o.order_uid = d.order_uid
		JOIN payments p ON o.order_uid = p.order_uid
	`

	err := s.db.Conn.SelectContext(ctx, &rows, query)
	if err != nil {
		s.logger.Error("Failed to load orders from database", zap.Error(err))
		return err
	}

	var orders []model.Order
	for _, row := range rows {
		order := model.Order{
			OrderUID:        row.OrderUID,
			TrackNumber:     row.TrackNumber,
			Entry:           row.Entry,
			Locale:          row.Locale,
			CustomerID:      row.CustomerID,
			DeliveryService: row.DeliveryService,
			ShardKey:        row.ShardKey,
			DateCreated:     row.DateCreated,
			Delivery: model.Delivery{
				Name:    row.DeliveryName,
				Phone:   row.Phone,
				Zip:     row.Zip,
				City:    row.City,
				Address: row.Address,
				Region:  row.Region,
				Email:   row.Email,
			},
			Payment: model.Payment{
				Transaction:  row.Transaction,
				Currency:     row.Currency,
				Provider:     row.Provider,
				Amount:       row.Amount,
				PaymentDt:    row.PaymentDt,
				Bank:         row.Bank,
				DeliveryCost: row.DeliveryCost,
				GoodsTotal:   row.GoodsTotal,
				CustomFee:    row.CustomFee,
			},
		}

		var items []model.Item
		itemQuery := `
			SELECT 
				chrt_id, 
				track_number, 
				price, 
				rid, 
				name, 
				sale, 
				size, 
				total_price, 
				nm_id, 
				brand, 
				status
			FROM items
			WHERE order_uid = $1
		`
		err := s.db.Conn.SelectContext(ctx, &items, itemQuery, row.OrderUID)
		if err != nil {
			s.logger.Error("Failed to load items for order", zap.String("order_uid", row.OrderUID), zap.Error(err))
			return err
		}
		order.Items = items

		orders = append(orders, order)
	}

	for _, order := range orders {
		s.cache.Set(order)
	}

	s.logger.Info("Cache loaded successfully", zap.Int("order_count", len(orders)))
	return nil
}

func (s *Service) GetOrder(ctx context.Context, orderUID string) (model.Order, error) {
	if order, exists := s.cache.Get(orderUID); exists {
		s.logger.Info("Order retrieved from cache", zap.String("order_uid", orderUID))
		return order, nil
	}

	var order model.Order
	query := `
		SELECT 
			o.order_uid, 
			o.track_number, 
			o.entry, 
			o.locale, 
			o.customer_id, 
			o.delivery_service, 
			o.shardkey, 
			o.sm_id, 
			o.date_created, 
			o.oof_shard,
			d.name AS delivery_name, 
			d.phone AS phone, 
			d.zip AS zip, 
			d.city AS city, 
			d.address AS address, 
			d.region AS region, 
			d.email AS email,
			p.transaction, 
			p.currency, 
			p.provider, 
			p.amount, 
			p.payment_dt, 
			p.bank, 
			p.delivery_cost, 
			p.goods_total, 
			p.custom_fee
		FROM orders o
		JOIN deliveries d ON o.order_uid = d.order_uid
		JOIN payments p ON o.order_uid = p.order_uid
		WHERE o.order_uid = $1
	`
	err := s.db.Conn.GetContext(ctx, &order, query, orderUID)
	if err != nil {
		s.logger.Error("Failed to get order from database", zap.String("order_uid", orderUID), zap.Error(err))
		return model.Order{}, err
	}

	var items []model.Item
	query = `
		SELECT 
			chrt_id, 
			track_number, 
			price, 
			rid, 
			name, 
			sale, 
			size, 
			total_price, 
			nm_id, 
			brand, 
			status
		FROM items
		WHERE order_uid = $1
	`
	err = s.db.Conn.SelectContext(ctx, &items, query, orderUID)
	if err != nil {
		s.logger.Error("Failed to load items for order", zap.String("order_uid", orderUID), zap.Error(err))
		return model.Order{}, err
	}
	order.Items = items

	s.cache.Set(order)
	s.logger.Info("Order retrieved from database and cached", zap.String("order_uid", orderUID))
	return order, nil
}
