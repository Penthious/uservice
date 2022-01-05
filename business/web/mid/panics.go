package mid

import (
	"context"
	"fmt"
	"github.com/Penthious/uservice/business/sys/metrics"
	"github.com/Penthious/uservice/foundation/web"
	"net/http"
	"runtime/debug"
)

func Panics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					err = fmt.Errorf("panic [%v] trace[%s]", rec, string(trace))

					metrics.AddPanics(ctx)
				}
			}()

			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
