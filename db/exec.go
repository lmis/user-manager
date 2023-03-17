package db

import (
	"context"
	"database/sql"
	"fmt"
	"user-manager/util/errors"

	"github.com/go-jet/jet/v2/qrm"
)

func FetchMaybe[A interface{}](query func(context.Context, qrm.Queryable, interface{}) error, tx *sql.Tx) (*A, error) {
	dest := new(A)
	ctx, cancelTimeout := DefaultQueryContext()
	defer cancelTimeout()
	err := query(ctx, tx, dest)

	if err != nil {
		if err == qrm.ErrNoRows || err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return dest, nil
}

func ExecSingleMutation(query func(context.Context, qrm.Executable) (sql.Result, error), tx *sql.Tx) error {
	ctx, cancelTimeout := DefaultQueryContext()
	defer cancelTimeout()
	res, err := query(ctx, tx)
	if err == nil {
		err = CheckAffectedRows(res, 1)
	}
	return err
}

func CheckAffectedRows(r sql.Result, expectedRows int64) error {
	rows, err := r.RowsAffected()
	if err != nil {
		return errors.Wrap("cannot get affected rows", err)
	}
	if rows != expectedRows {
		return errors.Wrap(fmt.Sprintf("wrong number of rows affected: %d, expected %d", rows, expectedRows), err)
	}
	return nil
}
