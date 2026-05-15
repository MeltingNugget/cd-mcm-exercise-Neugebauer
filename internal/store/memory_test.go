package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()
	createdProduct := model.Product{Name: "Test Product", Price: 19.99}
	s.Create(createdProduct)
	returnedProduct, err := s.GetByID(1)
	if err != nil {
		t.Errorf("failed to get product: %v", err)
	}
	if returnedProduct.Name != "Test Product" {
		t.Errorf("expected 'Test Product', got %s", returnedProduct.Name)
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
	createdProduct := model.Product{Name: "Test Product", Price: 19.99}
	s.Create(createdProduct)

	updatedProduct := model.Product{Name: "Updated Product", Price: 29.99}
	returnedProduct, err := s.Update(1, updatedProduct)
	if err != nil {
		t.Errorf("failed to update product: %v", err)
	}
	if returnedProduct.Name != "Updated Product" {
		t.Errorf("expected 'Updated Product', got %s", returnedProduct.Name)
	}

	// Second test, because while Update returns the updated product and we can check that directly,
	// we should also check that the store was actually updated by calling GetByID again.
	returnedProduct, err = s.GetByID(1)
	if err != nil {
		t.Errorf("failed to get product after update: %v", err)
	}
	if returnedProduct.Name != "Updated Product" {
		t.Errorf("expected 'Updated Product', got %s", returnedProduct.Name)
	}
}

func TestDeleteExisting(t *testing.T) {
	s := NewMemoryStore()
	createdProduct := model.Product{Name: "Test Product", Price: 19.99}
	s.Create(createdProduct)

	err := s.Delete(1)
	if err != nil {
		t.Errorf("failed to delete product: %v", err)
	}

	_, err = s.GetByID(1)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound after deleting product")
	}
}

func TestGetByIDInvalid(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetByID(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when getting non-existent product")
	}
}
