//************************************************************************//
// API "congo": Models
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/goadesign/gorma/example
// --design=github.com/goadesign/gorma/example/design
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package models

import (
	"github.com/goadesign/goa"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"time"
)

// TestModel
type Test struct {
	DeletedAt *time.Time // nullable timestamp (soft delete)
	CreatedAt time.Time  // timestamp
	UpdatedAt time.Time  // timestamp
}

// TableName overrides the table name settings in Gorm to force a specific table name
// in the database.
func (m Test) TableName() string {
	return "tests"

}

// TestDB is the implementation of the storage interface for
// Test.
type TestDB struct {
	Db gorm.DB
}

// NewTestDB creates a new storage type.
func NewTestDB(db gorm.DB) *TestDB {
	return &TestDB{Db: db}
}

// DB returns the underlying database.
func (m *TestDB) DB() interface{} {
	return &m.Db
}

// TestStorage represents the storage interface.
type TestStorage interface {
	DB() interface{}
	List(ctx context.Context) []Test
	Get(ctx context.Context) (Test, error)
	Add(ctx context.Context, test *Test) (*Test, error)
	Update(ctx context.Context, test *Test) error
	Delete(ctx context.Context) error
}

// TableName overrides the table name settings in Gorm to force a specific table name
// in the database.
func (m *TestDB) TableName() string {
	return "tests"

}

// CRUD Functions

// Get returns a single Test as a Database Model
// This is more for use internally, and probably not what you want in  your controllers
func (m *TestDB) Get(ctx context.Context) (Test, error) {
	now := time.Now()
	defer goa.Info(ctx, "Test:Get", goa.KV{"duration", time.Since(now)})
	var native Test
	err := m.Db.Table(m.TableName()).Where("").Find(&native).Error
	if err == gorm.RecordNotFound {
		return Test{}, nil
	}

	return native, err
}

// List returns an array of Test
func (m *TestDB) ListTest(ctx context.Context) []Test {
	now := time.Now()
	defer goa.Info(ctx, "Test:List", goa.KV{"duration", time.Since(now)})
	var objs []Test
	err := m.Db.Table(m.TableName()).Find(&objs).Error
	if err != nil && err != gorm.RecordNotFound {
		goa.Error(ctx, "error listing Test", goa.KV{"error", err.Error()})
		return objs
	}

	return objs
}

// Add creates a new record.  /// Maybe shouldn't return the model, it's a pointer.
func (m *TestDB) Add(ctx context.Context, model *Test) (*Test, error) {
	now := time.Now()
	defer goa.Info(ctx, "Test:Add", goa.KV{"duration", time.Since(now)})
	err := m.Db.Create(model).Error
	if err != nil {
		goa.Error(ctx, "error updating Test", goa.KV{"error", err.Error()})
		return model, err
	}

	return model, err
}

// Update modifies a single record.
func (m *TestDB) Update(ctx context.Context, model *Test) error {
	now := time.Now()
	defer goa.Info(ctx, "Test:Update", goa.KV{"duration", time.Since(now)})
	obj, err := m.Get(ctx)
	if err != nil {
		return err
	}
	err = m.Db.Model(&obj).Updates(model).Error

	return err
}

// Delete removes a single record.
func (m *TestDB) Delete(ctx context.Context) error {
	now := time.Now()
	defer goa.Info(ctx, "Test:Delete", goa.KV{"duration", time.Since(now)})
	var obj Test
	err := m.Db.Delete(&obj).Where("").Error

	if err != nil {
		goa.Error(ctx, "error retrieving Test", goa.KV{"error", err.Error()})
		return err
	}

	return nil
}
