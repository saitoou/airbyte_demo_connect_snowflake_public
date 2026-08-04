package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	DatabaseURL  string
	BatchID      string
	CustomerSize int
	OrderSize    int
	Seed         int64
}

type Customer struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Order struct {
	CustomerID  int64
	Status      string
	TotalAmount int64
	OrderedAt   time.Time
	UpdatedAt   time.Time
	CancelledAt *time.Time
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("generator failed: %v", err)
	}
}

func run() error {
	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	created, err := reserveBatch(
		ctx,
		tx,
		cfg.BatchID,
		cfg.CustomerSize,
		cfg.OrderSize,
	)
	if err != nil {
		return err
	}

	if !created {
		log.Printf(
			"batch %q has already been processed; skipping",
			cfg.BatchID,
		)
		return nil
	}

	random := rand.New(rand.NewSource(cfg.Seed))
	now := time.Now().UTC()

	customers := generateCustomers(
		random,
		cfg.CustomerSize,
		now,
	)

	// PostgreSQLへINSERTして、採番されたIDを取得する
	customers, err = insertCustomers(
		ctx,
		tx,
		customers,
	)
	if err != nil {
		return err
	}

	// 顧客IDが確定してから注文を作る
	orders := generateOrders(
		random,
		customers,
		cfg.OrderSize,
		now,
	)

	orderCount, err := insertOrders(ctx, tx, orders)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.Printf(
		"batch completed: batch_id=%s customers=%d orders=%d",
		cfg.BatchID,
		len(customers),
		orderCount,
	)

	return nil
}

func parseConfig() (Config, error) {
	var cfg Config

	flag.StringVar(
		&cfg.BatchID,
		"batch-id",
		"",
		"Unique batch identifier used to prevent duplicate insertion",
	)
	flag.IntVar(
		&cfg.CustomerSize,
		"customers",
		10,
		"Number of customers to generate",
	)
	flag.IntVar(
		&cfg.OrderSize,
		"orders",
		50,
		"Number of orders to generate",
	)
	flag.Int64Var(
		&cfg.Seed,
		"seed",
		time.Now().UnixNano(),
		"Random seed",
	)
	flag.Parse()

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")

	switch {
	case cfg.DatabaseURL == "":
		return Config{}, errors.New("DATABASE_URL is required")
	case cfg.BatchID == "":
		return Config{}, errors.New("--batch-id is required")
	case cfg.CustomerSize <= 0:
		return Config{}, errors.New("--customers must be greater than 0")
	case cfg.OrderSize < 0:
		return Config{}, errors.New("--orders must be 0 or greater")
	}

	return cfg, nil
}

func reserveBatch(
	ctx context.Context,
	tx pgx.Tx,
	batchID string,
	customerCount int,
	orderCount int,
) (bool, error) {
	var insertedBatchID string

	err := tx.QueryRow(
		ctx,
		`
			insert into generated_batches (
				batch_id,
				customer_count,
				order_count
			)
			values ($1, $2, $3)
			on conflict (batch_id) do nothing
			returning batch_id
		`,
		batchID,
		customerCount,
		orderCount,
	).Scan(&insertedBatchID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("reserve batch: %w", err)
	}

	return true, nil
}

func generateCustomers(
	random *rand.Rand,
	count int,
	now time.Time,
) []Customer {
	firstNames := []string{
		"Haruto",
		"Yuto",
		"Sota",
		"Yui",
		"Aoi",
		"Rin",
	}

	lastNames := []string{
		"Sato",
		"Suzuki",
		"Takahashi",
		"Tanaka",
		"Ito",
	}

	customers := make([]Customer, 0, count)

	for i := 0; i < count; i++ {
		name := fmt.Sprintf(
			"%s %s",
			lastNames[random.Intn(len(lastNames))],
			firstNames[random.Intn(len(firstNames))],
		)

		createdAt := now.Add(
			-time.Duration(random.Intn(30*24)) * time.Hour,
		)

		customers = append(customers, Customer{
			Name: name,
			Email: fmt.Sprintf(
				"customer-%d-%03d@example.com",
				now.UnixNano(),
				i+1,
			),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
			DeletedAt: nil,
		})
	}

	return customers
}

func insertCustomers(
	ctx context.Context,
	tx pgx.Tx,
	customers []Customer,
) ([]Customer, error) {
	for i := range customers {
		err := tx.QueryRow(
			ctx,
			`
				insert into customers (
					name,
					email,
					created_at,
					updated_at,
					deleted_at
				)
				values ($1, $2, $3, $4, $5)
				returning id
			`,
			customers[i].Name,
			customers[i].Email,
			customers[i].CreatedAt,
			customers[i].UpdatedAt,
			customers[i].DeletedAt,
		).Scan(&customers[i].ID)

		if err != nil {
			return nil, fmt.Errorf(
				"insert customer at index %d: %w",
				i,
				err,
			)
		}
	}

	return customers, nil
}

func generateOrders(
	random *rand.Rand,
	customers []Customer,
	count int,
	now time.Time,
) []Order {
	orders := make([]Order, 0, count)

	for i := 0; i < count; i++ {
		selectedCustomer := customers[random.Intn(len(customers))]
		status := randomOrderStatus(random)

		orderedAt := selectedCustomer.CreatedAt.Add(
			time.Duration(random.Intn(14*24+1)) * time.Hour,
		)

		if orderedAt.After(now) {
			orderedAt = now
		}

		updatedAt := orderedAt
		var cancelledAt *time.Time

		if status == "cancelled" {
			cancelledTime := orderedAt.Add(
				time.Duration(random.Intn(48)+1) * time.Hour,
			)

			if cancelledTime.After(now) {
				cancelledTime = now
			}

			updatedAt = cancelledTime
			cancelledAt = &cancelledTime
		}

		orders = append(orders, Order{
			CustomerID: selectedCustomer.ID,
			Status:     status,

			// 500円から50,000円
			TotalAmount: int64(random.Intn(49_501) + 500),

			OrderedAt:   orderedAt,
			UpdatedAt:   updatedAt,
			CancelledAt: cancelledAt,
		})
	}

	return orders
}

func randomOrderStatus(random *rand.Rand) string {
	value := random.Intn(100)

	switch {
	case value < 10:
		return "pending"
	case value < 30:
		return "paid"
	case value < 50:
		return "shipped"
	case value < 90:
		return "completed"
	default:
		return "cancelled"
	}
}

func insertOrders(
	ctx context.Context,
	tx pgx.Tx,
	orders []Order,
) (int64, error) {
	var insertedCount int64

	for i, order := range orders {
		commandTag, err := tx.Exec(
			ctx,
			`
				insert into orders (
					customer_id,
					status,
					total_amount,
					ordered_at,
					updated_at,
					cancelled_at
				)
				values ($1, $2, $3, $4, $5, $6)
			`,
			order.CustomerID,
			order.Status,
			order.TotalAmount,
			order.OrderedAt,
			order.UpdatedAt,
			order.CancelledAt,
		)
		if err != nil {
			return 0, fmt.Errorf(
				"insert order at index %d: %w",
				i,
				err,
			)
		}

		insertedCount += commandTag.RowsAffected()
	}

	return insertedCount, nil
}
