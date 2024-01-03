package main

import (
	"context"
	"database/sql"
	"fmt"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type XPTOEntity struct {
	ID uint64
}

type XPTORepository struct {
	db DB
}

func (repository *XPTORepository) Create(entity *XPTOEntity) error {
	entity.ID = uint64(1)
	return nil
}

type OTPXEntity struct{}

type OTPXRepository struct {
	db DB
}

func (repository *OTPXRepository) Update(entity OTPXEntity) error {
	return nil
}

type RepositoryCoordinator struct {
	db             *sql.DB
	xptoRepository *XPTORepository
	otpxRepository *OTPXRepository
}

func (c *RepositoryCoordinator) XPTORepository() *XPTORepository {
	return c.xptoRepository
}

func (c *RepositoryCoordinator) OTPXRepository() *OTPXRepository {
	return c.otpxRepository
}

type TransactionFunc func(t Transactor) error

func (c *RepositoryCoordinator) Transaction(f TransactionFunc) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	err = f(Transactor{tx: tx})
	if err != nil {
		rollBackErr := tx.Rollback()
		if rollBackErr != nil {
			return rollBackErr
		}

		return err
	}

	return nil
}

type Transactor struct {
	tx             *sql.Tx
	xptoRepository *XPTORepository
	otpxRepository *OTPXRepository
}

func (t Transactor) XPTORepository() *XPTORepository {
	if t.xptoRepository == nil {
		t.xptoRepository = &XPTORepository{db: t.tx}
	}

	return t.xptoRepository
}

func (t Transactor) OTPXRepository() *OTPXRepository {
	if t.otpxRepository == nil {
		t.otpxRepository = &OTPXRepository{db: t.tx}
	}

	return t.otpxRepository
}

func main() {
	db, err := sql.Open("", "")
	if err != nil {
		panic(err)
	}
	r := RepositoryCoordinator{db: db}

	xpto := &XPTOEntity{}

	err = r.Transaction(func(t Transactor) error {
		err := t.XPTORepository().Create(xpto)
		if err != nil {
			return err
		}

		err = t.otpxRepository.Update(OTPXEntity{})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(xpto.ID)
}
