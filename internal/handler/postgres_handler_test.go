package handler

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mrckurz/CI-CD-MCM/internal/store"
)

func init() {
	sql.Register("fakepghandler", &fakeDriver{})
}

type fakeDriver struct{}

type fakeConn struct{}

type fakeStmt struct{ query string }

type fakeRows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

type fakeResult struct{ rows int64 }

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{}, nil
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{query: query}, nil
}

func (c *fakeConn) Close() error { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) { return nil, errors.New("not supported") }
func (c *fakeConn) Ping(ctx context.Context) error { return nil }
func (c *fakeConn) CheckNamedValue(nv *driver.NamedValue) error { return nil }
func (c *fakeConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return &fakeStmt{query: query}, nil
}

func (c *fakeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return fakeQuery(strings.TrimSpace(query), args)
}

func (c *fakeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return fakeExec(strings.TrimSpace(query), args)
}

func (s *fakeStmt) Close() error { return nil }
func (s *fakeStmt) NumInput() int { return -1 }
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

func fakeQuery(query string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.HasPrefix(query, "SELECT id, name, price FROM products ORDER BY id"):
		return &fakeRows{cols: []string{"id", "name", "price"}, rows: [][]driver.Value{{int64(1), "Product A", 9.99}}}, nil
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

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	for i := range r.rows[r.pos] {
		dest[i] = r.rows[r.pos][i]
	}
	r.pos++
	return nil
}

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rows, nil }

func newFakePostgresHandler(t *testing.T) *PostgresHandler {
	db, err := sql.Open("fakepghandler", "")
	if err != nil {
		t.Fatalf("failed to open fake postgres db: %v", err)
	}

	return NewPostgresHandler(&store.PostgresStore{DB: db})
}

func newPostgresRouter(t *testing.T) (*mux.Router, *PostgresHandler) {
	h := newFakePostgresHandler(t)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r, h
}

func TestPostgresHealthOK(t *testing.T) {
	r, _ := newPostgresRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresGetProducts(t *testing.T) {
	r, _ := newPostgresRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresGetProductNotFound(t *testing.T) {
	r, _ := newPostgresRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/products/99", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPostgresCreateUpdateDelete(t *testing.T) {
	r, _ := newPostgresRouter(t)

	createBody := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(createBody))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	updateBody := `{"name":"Updated Widget","price":19.99}`
	req = httptest.NewRequest(http.MethodPut, "/products/1", strings.NewReader(updateBody))
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/products/1", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
