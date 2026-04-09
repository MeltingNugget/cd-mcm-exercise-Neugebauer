package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()
	// TODO: Add test -- create a product and verify GetByID returns it
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

// TODO: Add tests for Update, Delete of existing product, and GetByID with invalid ID
func TestUpdateProduct(t *testing.T) {
	s := NewMemoryStore()
	// TODO: Add test -- create a product and verify GetByID returns it
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
	// TODO: Add test -- create a product and verify GetByID returns it
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
		t.Errorf("expected ErrNotFound when getting trying to retrieve products without ever having created any")
	}
}
