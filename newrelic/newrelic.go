package newrelic

import (
	"context"
	"net/http"
	"strings"

	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

const (
	newRelicTransactionID = "NewRelicTransaction"
)

var (
	// NewRelicApp is the reference for newrelic application
	NewRelicApp *newrelic.Application
)

func SetNewRelicApp(nr *newrelic.Application) {
	NewRelicApp = nr
}

/// Use NewRelic better - reference https://github.com/carousell/Orion/blob/19b7601394006ca4eb9dcb65a2339c2046111f75/utils/utils.go

//GetNewRelicTransactionFromContext fetches the new relic transaction that is stored in the context
func GetNewRelicTransactionFromContext(ctx context.Context) *newrelic.Transaction {
	t := ctx.Value(newRelicTransactionID)
	if t != nil {
		txn, ok := t.(*newrelic.Transaction)
		if ok {
			return txn
		}
	}
	return nil
}

func GetOrStartNew(ctx context.Context, name string) (*newrelic.Transaction, context.Context) {
	ctx = StartNRTransaction(name, ctx, nil, nil)
	return GetNewRelicTransactionFromContext(ctx), ctx
}

//StoreNewRelicTransactionToContext stores a new relic transaction object to context
func StoreNewRelicTransactionToContext(ctx context.Context, t *newrelic.Transaction) context.Context {
	if t == nil {
		return ctx
	}
	return context.WithValue(ctx, newRelicTransactionID, t)
}

//StartNRTransaction starts a new newrelic transaction
func StartNRTransaction(path string, ctx context.Context, req *http.Request, w http.ResponseWriter) context.Context {
	if req == nil {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		req, _ = http.NewRequest("", path, nil)
	}
	// check if transaction has been initialized
	if NewRelicApp == nil {
		return ctx
	}
	t := GetNewRelicTransactionFromContext(ctx)
	if t == nil {
		t = NewRelicApp.StartTransaction(path)
		if t != nil {
			t.SetWebRequestHTTP(req)
			ctx = StoreNewRelicTransactionToContext(ctx, t)
		}
	}
	return ctx
}

//FinishNRTransaction finishes an existing transaction
func FinishNRTransaction(ctx context.Context, err error) {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		t.NoticeError(err)
		t.End()
	}
}

//IgnoreNRTransaction ignores this NR trasaction and prevents it from being reported
func IgnoreNRTransaction(ctx context.Context) {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		t.Ignore()
	}
}
