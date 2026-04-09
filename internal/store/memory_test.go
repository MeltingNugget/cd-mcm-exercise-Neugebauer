package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	tests := []struct {
		name string
		id   int
	}{
		{"case 1", 1},
		{"case 2", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := s.Create(model.Product{Name: tt.name, ID: tt.id})
			retrieved, err := s.GetByID(product.ID)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if retrieved.ID != product.ID {
				t.Errorf("retrieved product does not match created product")
			}
			if tt.name == "case 2" && retrieved.ID != 2 { // product.ID is manually set to 3 but the store should assign ID 2 due to auto increment
				t.Errorf("retrieved product ID does not match expected ID")
			}
		})
	}

	product := s.Create(model.Product{Name: "Test Product", Price: 9.99})
	retrieved, err := s.GetByID(product.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.ID != product.ID {
		t.Errorf("retrieved product does not match created product")
	}
}

func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}

func TestUpdateProduct(t *testing.T) {
	s := NewMemoryStore()
	product := s.Create(model.Product{Name: "Test Product", Price: 9.99})

	update := model.Product{Name: "Updated Product", Price: 19.99}
	/*updated := */ s.Update(product.ID, update)

	retrieved, err := s.GetByID(product.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.ID != product.ID {
		t.Errorf("retrieved product does not match created product")
	}
	if retrieved.Name != update.Name || retrieved.Price != update.Price {
		t.Errorf("retrieved product does not match updated product")
	}
}

func TestDeleteProduct(t *testing.T) {
	s := NewMemoryStore()
	product := s.Create(model.Product{Name: "Test Product", Price: 9.99})

	err := s.Delete(product.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = s.GetByID(product.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound when getting deleted product")
	}
}

func TestGetByIDNotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetByID(1)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound while trying to retrieve products without ever having created any")
	}
}
