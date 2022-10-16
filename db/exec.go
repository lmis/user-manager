package db

import (
	"context"
	"database/sql"
	"fmt"
	"user-manager/util"
	"user-manager/util/nullable"
)

func Fetch[A interface{}](query func(ctx context.Context) (A, error)) (nullable.Nullable[A], error) {
	ctx, cancelTimeout := DefaultQueryContext()
	defer cancelTimeout()
	res, err := query(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nullable.Empty[A](), nil
		}
		return nullable.Empty[A](), err
	}
	return nullable.Of(res), nil
}

func ExecSingleMutation(query func(ctx context.Context) (int64, error)) error {
	ctx, cancelTimeout := DefaultQueryContext()
	defer cancelTimeout()
	rows, err := query(ctx)
	if err != nil {
		return err
	}
	if rows != 1 {
		return util.Error(fmt.Sprintf("too many rows affected: %d", rows))
	}
	return nil
}
