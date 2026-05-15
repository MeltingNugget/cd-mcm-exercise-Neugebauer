package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func init() {
	sql.Register("fakepgstore", &fakeDriver{})
}

type fakeDriver struct{}

type fakeConn struct{}

type fakeStmt struct {
	query string
}

type fakeRows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

type fakeResult struct {
	rows int64
}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{}, nil
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{query: query}, nil
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not supported")
}

func (c *fakeConn) Ping(ctx context.Context) error {
	return nil
}

func (c *fakeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return fakeQuery(strings.TrimSpace(query), args)
}

func (c *fakeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return fakeExec(strings.TrimSpace(query), args)
}

func (c *fakeConn) CheckNamedValue(nv *driver.NamedValue) error {
	return nil
}

func (c *fakeConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return &fakeStmt{query: query}, nil
}

func (s *fakeStmt) Close() error {
	return nil
}

func (s *fakeStmt) NumInput() int {
	return -1
}

func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return fakeExec(strings.TrimSpace(s.query), named)
}

func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return fakeQuery(strings.TrimSpace(s.query), named)
}

func (s *fakeStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return fakeExec(strings.TrimSpace(s.query), args)
}

func (s *fakeStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return fakeQuery(strings.TrimSpace(s.query), args)
}

func fakeExec(query string, args []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.HasPrefix(query, "CREATE TABLE IF NOT EXISTS"):
		return fakeResult{rows: 0}, nil
	case strings.HasPrefix(query, "UPDATE products SET name"):
		if len(args) == 3 && argAsInt64(args[2], 0) == 1 {
			return fakeResult{rows: 1}, nil
		}
		return fakeResult{rows: 0}, nil
	case strings.HasPrefix(query, "DELETE FROM products WHERE id"):
		if len(args) == 1 && argAsInt64(args[0], 0) == 1 {
			return fakeResult{rows: 1}, nil
		}
		return fakeResult{rows: 0}, nil
	}
	return nil, errors.New("unexpected exec query")
}

func argAsInt64(arg driver.NamedValue, fallback int64) int64 {
	switch v := arg.Value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	case uint64:
		return int64(v)
	case uint32:
		return int64(v)
	case uint:
		return int64(v)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err == nil {
			return i
		}
	}
	return fallback
}

func fakeQuery(query string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.HasPrefix(query, "SELECT id, name, price FROM products ORDER BY id"):
		return &fakeRows{
			cols: []string{"id", "name", "price"},
			rows: [][]driver.Value{{int64(1), "Product A", 9.99}, {int64(2), "Product B", 19.99}},
		}, nil
	case strings.HasPrefix(query, "SELECT id, name, price FROM products WHERE id"):
		if len(args) == 1 && argAsInt64(args[0], 0) == 1 {
			return &fakeRows{cols: []string{"id", "name", "price"}, rows: [][]driver.Value{{int64(1), "Product A", 9.99}}}, nil
		}
		return &fakeRows{cols: []string{"id", "name", "price"}, rows: [][]driver.Value{}}, nil
	case strings.HasPrefix(query, "INSERT INTO products"):
		return &fakeRows{cols: []string{"id"}, rows: [][]driver.Value{{int64(1)}}}, nil
	}
	return nil, errors.New("unexpected query")
}

func (r *fakeRows) Columns() []string {
	return r.cols
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.pos]
	copy(dest, row)
	r.pos++
	return nil
}

func (r fakeResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r fakeResult) RowsAffected() (int64, error) {
	return r.rows, nil
}

func TestPostgresStoreCRUD(t *testing.T) {
	db, err := sql.Open("fakepgstore", "")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	defer db.Close()

	s := &PostgresStore{DB: db}

	if err := s.EnsureTable(); err != nil {
		t.Fatalf("EnsureTable failed: %v", err)
	}

	products, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}

	p, err := s.GetByID(1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if p.Name != "Product A" {
		t.Fatalf("expected Product A, got %s", p.Name)
	}

	if _, err := s.GetByID(99); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for missing product, got %v", err)
	}

	created, err := s.Create(model.Product{Name: "Created", Price: 11.11})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected ID 1, got %d", created.ID)
	}

	updated, err := s.Update(1, model.Product{Name: "Updated", Price: 22.22})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.ID != 1 || updated.Name != "Updated" {
		t.Fatalf("expected updated product, got %#v", updated)
	}

	if _, err := s.Update(99, model.Product{Name: "Nope", Price: 1}); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound updating missing product, got %v", err)
	}

	if err := s.Delete(1); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if err := s.Delete(99); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound deleting missing product, got %v", err)
	}
}

func TestNewPostgresStore_InvalidHost(t *testing.T) {
	// Ping schlägt fehl → return nil, err (Zeile 26-28)
	_, err := NewPostgresStore("invalidhost", "5432", "catalog", "catalog123", "productcatalog")
	if err == nil {
		t.Fatal("expected error for invalid host, got nil")
	}
}

func TestNewPostgresStore_InvalidCredentials(t *testing.T) {
	// sql.Open schlägt nicht fehl, aber Ping schlägt fehl → Zeile 23-25
	_, err := NewPostgresStore("localhost", "5432", "wronguser", "wrongpassword", "productcatalog")
	if err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}
}
