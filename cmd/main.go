package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"order-service/internal/kafka"
	"order-service/internal/storage"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// инициализация подключения к БД
	connStr := os.Getenv("DB_CONN_STR")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable"
	}

	order_db, err := storage.NewPostgresStorage(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// инициализация кэша
	cache := storage.NewCache()

	// восстановление кэша из БД при запуске
	orders, err := order_db.GetAllOrders()
	if err != nil {
		log.Fatal("Failed to load orders from database: ", err)
	}
	for _, order := range orders {
		cache.Set(order)
	}

	// инициализация Kafka consumer
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	if kafkaBrokers[0] == "" {
		kafkaBrokers = []string{"localhost:9092"}
	}

	consumer, err := kafka.NewKafkaConsumer(kafkaBrokers, "orders", order_db, cache)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatal("Kafka consumer error: ", err)
		}
	}()

	http.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		uid := r.URL.Path[len("/order/"):]
		if uid == "" {
			http.Error(w, "Order UID is required", http.StatusBadRequest)
			return
		}

		// попытка получить данные о заказе из кэша
		order, success := cache.Get(uid)
		if !success {
			order, err = order_db.GetOrderByUID(uid)
			if err != nil {
				http.Error(w, "Order not found", http.StatusNotFound)
				return
			}
			cache.Set(order)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	})

	// cтатический файл для веб-интерфейса
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server started on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("HTTP server error: ", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down...")
	cancel()
	time.Sleep(1 * time.Second)
}
