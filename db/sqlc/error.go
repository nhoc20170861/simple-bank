package db

import (
	"errors"

	"github.com/lib/pq"
)

const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

var ErrForeignKeyViolation = &pq.Error{
	Code: pq.ErrorCode(ForeignKeyViolation),
}

var ErrUniqueViolation = &pq.Error{
	Code: pq.ErrorCode(UniqueViolation),
}

func ErrorCode(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code)
	}
	return ""
}
