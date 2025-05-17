package app

import "context"

type App interface {
	// Run starts or initializes any resources (connections, pools).
	Run(ctx context.Context) error
}
