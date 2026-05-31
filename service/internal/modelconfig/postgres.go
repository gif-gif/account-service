package modelconfig

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repo *PostgresRepository) List(ctx context.Context) ([]Item, error) {
	rows, err := repo.pool.Query(ctx, `
		select id::text, kind, key, value, status, display_order, created_at, updated_at
		from model_config_items
		order by kind, display_order, key
	`)
	if err != nil {
		return nil, fmt.Errorf("list model config items: %w", err)
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Kind, &item.Key, &item.Value, &item.Status, &item.DisplayOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan model config item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate model config items: %w", err)
	}
	return items, nil
}

func (repo *PostgresRepository) Create(ctx context.Context, request CreateItemRequest) (Item, error) {
	var item Item
	err := repo.pool.QueryRow(ctx, `
		insert into model_config_items (kind, key, value, status, display_order)
		values ($1, $2, $3, $4, $5)
		returning id::text, kind, key, value, status, display_order, created_at, updated_at
	`, request.Kind, request.Key, request.Value, request.Status, request.DisplayOrder).Scan(
		&item.ID,
		&item.Kind,
		&item.Key,
		&item.Value,
		&item.Status,
		&item.DisplayOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Item{}, fmt.Errorf("create model config item: %w", err)
	}
	return item, nil
}

func (repo *PostgresRepository) Update(ctx context.Context, id string, request UpdateItemRequest) (Item, error) {
	current, err := repo.get(ctx, id)
	if err != nil {
		return Item{}, err
	}
	if request.Kind != nil {
		current.Kind = *request.Kind
	}
	if request.Key != nil {
		current.Key = *request.Key
	}
	if request.Value != nil {
		current.Value = *request.Value
	}
	if request.Status != nil {
		current.Status = *request.Status
	}
	if request.DisplayOrder != nil {
		current.DisplayOrder = *request.DisplayOrder
	}

	var item Item
	err = repo.pool.QueryRow(ctx, `
		update model_config_items
		set kind = $2, key = $3, value = $4, status = $5, display_order = $6, updated_at = now()
		where id = $1
		returning id::text, kind, key, value, status, display_order, created_at, updated_at
	`, id, current.Kind, current.Key, current.Value, current.Status, current.DisplayOrder).Scan(
		&item.ID,
		&item.Kind,
		&item.Key,
		&item.Value,
		&item.Status,
		&item.DisplayOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Item{}, fmt.Errorf("update model config item: %w", err)
	}
	return item, nil
}

func (repo *PostgresRepository) Delete(ctx context.Context, id string) error {
	tag, err := repo.pool.Exec(ctx, `delete from model_config_items where id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete model config item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("model config item not found")
	}
	return nil
}

func (repo *PostgresRepository) get(ctx context.Context, id string) (Item, error) {
	var item Item
	err := repo.pool.QueryRow(ctx, `
		select id::text, kind, key, value, status, display_order, created_at, updated_at
		from model_config_items
		where id = $1
	`, id).Scan(&item.ID, &item.Kind, &item.Key, &item.Value, &item.Status, &item.DisplayOrder, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, errors.New("model config item not found")
	}
	if err != nil {
		return Item{}, fmt.Errorf("get model config item: %w", err)
	}
	return item, nil
}
