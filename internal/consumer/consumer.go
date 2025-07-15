package consumer

import (
	"context"
	"encoding/json"
	"orderHub_L0/internal/cache"
	"orderHub_L0/internal/db"
	"orderHub_L0/internal/model"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	db     *db.DB
	cache  *cache.Cache
	reader *kafka.Reader
	logger *zap.Logger
}

func NewConsumer(db *db.DB, cache *cache.Cache, brokers []string, topic string, logger *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		Topic:           topic,
		MinBytes:        10e3,
		MaxBytes:        10e6,
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
	})

	return &Consumer{
		db:     db,
		cache:  cache,
		reader: reader,
		logger: logger,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.reader.Close()
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read Kafka message", zap.Error(err))
				continue
			}

			var order model.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				c.logger.Error("Failed to unmarshal order", zap.String("message", string(msg.Value)), zap.Error(err))
				continue
			}

			tx, err := c.db.Conn.Beginx()
			if err != nil {
				c.logger.Error("Failed to start transaction", zap.String("order_uid", order.OrderUID), zap.Error(err))
				continue
			}

			_, err = tx.NamedExec(`
INSERT INTO orders (order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES (:order_uid, :track_number, :entry, :locale, :customer_id, :delivery_service, :shardkey, :sm_id, :date_created, :oof_shard)
`, order)
			if err != nil {
				c.logger.Error("Failed to insert order", zap.String("order_uid", order.OrderUID), zap.Error(err))
				tx.Rollback()
				continue
			}

		
			deliveryData := map[string]interface{}{
				"order_uid":     order.OrderUID,
				"delivery_name": order.Delivery.Name,
				"phone":         order.Delivery.Phone,
				"zip":           order.Delivery.Zip,
				"city":          order.Delivery.City,
				"address":       order.Delivery.Address,
				"region":        order.Delivery.Region,
				"email":         order.Delivery.Email,
			}
			_, err = tx.NamedExec(`
INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
VALUES (:order_uid, :delivery_name, :phone, :zip, :city, :address, :region, :email)
`, deliveryData)
			if err != nil {
				c.logger.Error("Failed to insert delivery", zap.String("order_uid", order.OrderUID), zap.Error(err))
				tx.Rollback()
				continue
			}

			_, err = tx.NamedExec(`
INSERT INTO payments (order_uid, transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES (:order_uid, :transaction, :currency, :provider, :amount, :payment_dt, :bank, :delivery_cost, :goods_total, :custom_fee)
`, order.Payment)
			if err != nil {
				c.logger.Error("Failed to insert payment", zap.String("order_uid", order.OrderUID), zap.Error(err))
				tx.Rollback()
				continue
			}

			
			for i, item := range order.Items {
				itemData := map[string]interface{}{
					"order_uid":    order.OrderUID,
					"chrt_id":      item.ChrtID,
					"track_number": item.TrackNumber,
					"price":        item.Price,
					"rid":          item.RID,
					"name":         item.Name,
					"sale":         item.Sale,
					"size":         item.Size,
					"total_price":  item.TotalPrice,
					"nm_id":        item.NMID,
					"brand":        item.Brand,
					"status":       item.Status,
				}
				_, err = tx.NamedExec(`
INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
VALUES (:order_uid, :chrt_id, :track_number, :price, :rid, :name, :sale, :size, :total_price, :nm_id, :brand, :status)
`, itemData)
				if err != nil {
					c.logger.Error("Failed to insert item", zap.String("order_uid", order.OrderUID), zap.Int("item_index", i), zap.Error(err))
					tx.Rollback()
					continue
				}
			}

			if err := tx.Commit(); err != nil {
				c.logger.Error("Failed to commit transaction", zap.String("order_uid", order.OrderUID), zap.Error(err))
				tx.Rollback()
				continue
			}

			c.cache.Set(order)
			c.logger.Info("Order processed and cached", zap.String("order_uid", order.OrderUID))
		}
	}
}
