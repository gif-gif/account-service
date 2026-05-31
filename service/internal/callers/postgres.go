package callers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"account-service/service/internal/security"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (store *PostgresStore) Create(name string, description string) (CreateResult, error) {
	return store.CreateWithStatus(CreateRequest{Name: name, Description: description, Status: StatusActive})
}

func (store *PostgresStore) CreateWithStatus(request CreateRequest) (CreateResult, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	if request.Status == "" {
		request.Status = StatusActive
	}
	if !validStatus(request.Status) {
		return CreateResult{}, errors.New("invalid caller status")
	}

	apiKey, err := security.GenerateAPIKey()
	if err != nil {
		return CreateResult{}, err
	}

	var caller Caller
	err = store.pool.QueryRow(context.Background(), `
		insert into api_callers (name, api_key_hash, api_key_plaintext, status, description)
		values ($1, $2, $3, $4, $5)
		returning id::text, name, api_key_hash, api_key_plaintext, status, description, created_at, updated_at
	`, request.Name, security.HashAPIKey(apiKey), apiKey, request.Status, request.Description).Scan(
		&caller.ID,
		&caller.Name,
		&caller.APIKeyHash,
		&caller.PlaintextAPIKey,
		&caller.Status,
		&caller.Description,
		&caller.CreatedAt,
		&caller.UpdatedAt,
	)
	if err != nil {
		return CreateResult{}, fmt.Errorf("create caller: %w", err)
	}
	return CreateResult{Caller: caller, PlaintextAPIKey: apiKey}, nil
}

func (store *PostgresStore) List() ([]Caller, error) {
	rows, err := store.pool.Query(context.Background(), `
		select id::text, name, api_key_hash, api_key_plaintext, status, description, created_at, updated_at
		from api_callers
		order by created_at desc, name
	`)
	if err != nil {
		return nil, fmt.Errorf("list callers: %w", err)
	}
	defer rows.Close()

	callers := []Caller{}
	for rows.Next() {
		caller, err := scanCaller(rows)
		if err != nil {
			return nil, fmt.Errorf("scan caller: %w", err)
		}
		callers = append(callers, caller)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate callers: %w", err)
	}
	return callers, nil
}

func (store *PostgresStore) Authenticate(apiKey string) (Caller, bool) {
	rows, err := store.pool.Query(context.Background(), `
		select id::text, name, api_key_hash, api_key_plaintext, status, description, created_at, updated_at
		from api_callers
		where status = $1
	`, StatusActive)
	if err != nil {
		return Caller{}, false
	}
	defer rows.Close()

	for rows.Next() {
		caller, err := scanCaller(rows)
		if err != nil {
			return Caller{}, false
		}
		if security.VerifyAPIKey(apiKey, caller.APIKeyHash) {
			return caller, true
		}
	}
	return Caller{}, false
}

func (store *PostgresStore) RevealAPIKey(id string) (string, error) {
	var apiKey string
	err := store.pool.QueryRow(context.Background(), `
		select api_key_plaintext
		from api_callers
		where id = $1
	`, id).Scan(&apiKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("caller not found")
	}
	if err != nil {
		return "", fmt.Errorf("reveal api key: %w", err)
	}
	if apiKey == "" {
		return "", errors.New("api key secret not available")
	}
	return apiKey, nil
}

func (store *PostgresStore) Update(id string, request UpdateRequest) (Caller, error) {
	caller, err := store.get(id)
	if err != nil {
		return Caller{}, err
	}
	if request.Name != nil {
		caller.Name = strings.TrimSpace(*request.Name)
	}
	if request.Description != nil {
		caller.Description = strings.TrimSpace(*request.Description)
	}
	if request.Status != nil {
		status := Status(strings.TrimSpace(string(*request.Status)))
		if !validStatus(status) {
			return Caller{}, errors.New("invalid caller status")
		}
		caller.Status = status
	}

	err = store.pool.QueryRow(context.Background(), `
		update api_callers
		set name = $2, status = $3, description = $4, updated_at = now()
		where id = $1
		returning id::text, name, api_key_hash, api_key_plaintext, status, description, created_at, updated_at
	`, id, caller.Name, caller.Status, caller.Description).Scan(
		&caller.ID,
		&caller.Name,
		&caller.APIKeyHash,
		&caller.PlaintextAPIKey,
		&caller.Status,
		&caller.Description,
		&caller.CreatedAt,
		&caller.UpdatedAt,
	)
	if err != nil {
		return Caller{}, fmt.Errorf("update caller: %w", err)
	}
	return caller, nil
}

func (store *PostgresStore) Disable(id string) error {
	status := StatusDisabled
	_, err := store.Update(id, UpdateRequest{Status: &status})
	return err
}

func (store *PostgresStore) Delete(id string) error {
	tag, err := store.pool.Exec(context.Background(), `delete from api_callers where id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete caller: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("caller not found")
	}
	return nil
}

func (store *PostgresStore) get(id string) (Caller, error) {
	caller, err := scanCallerRow(store.pool.QueryRow(context.Background(), `
		select id::text, name, api_key_hash, api_key_plaintext, status, description, created_at, updated_at
		from api_callers
		where id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Caller{}, errors.New("caller not found")
	}
	if err != nil {
		return Caller{}, fmt.Errorf("get caller: %w", err)
	}
	return caller, nil
}

type callerScanner interface {
	Scan(dest ...any) error
}

func scanCallerRow(row callerScanner) (Caller, error) {
	var caller Caller
	err := row.Scan(
		&caller.ID,
		&caller.Name,
		&caller.APIKeyHash,
		&caller.PlaintextAPIKey,
		&caller.Status,
		&caller.Description,
		&caller.CreatedAt,
		&caller.UpdatedAt,
	)
	return caller, err
}

func scanCaller(rows pgx.Rows) (Caller, error) {
	return scanCallerRow(rows)
}

var _ Store = (*PostgresStore)(nil)
