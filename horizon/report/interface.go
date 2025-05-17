package report

import "context"

type Report[T any] interface {

	// This converts data to csv
	Generate(ctx context.Context, data T) (header []string, body [][]string, err error)
}
