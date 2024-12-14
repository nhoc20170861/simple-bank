package db

import "github.com/lib/pq"

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
