package db

import (
	"context"
	"database/sql"
	"fmt"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/go-jet/jet/v2/qrm"
)

func Fetch[A interface{}, B interface{}](query func(context.Context, qrm.Queryable, interface{}) error, convert func(*A) B, tx *sql.Tx) (nullable.Nullable[B], error) {
	dest := new(A)
	ctx, cancelTimeout := DefaultQueryContext()
	defer cancelTimeout()
	err := query(ctx, tx, dest)

	if err != nil {
		if err == qrm.ErrNoRows || err == sql.ErrNoRows {
			return nullable.Empty[B](), nil
		}
		return nullable.Empty[B](), err
	}

	return nullable.Of(convert(dest)), nil
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
		return util.Wrap("cannot get affected rows", err)
	}
	if rows != expectedRows {
		return util.Wrap(fmt.Sprintf("wrong number of rows affected: %d, expected %d", rows, expectedRows), err)
	}
	return nil
}
