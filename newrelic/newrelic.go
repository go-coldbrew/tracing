package newrelic

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelicApp is the reference for newrelic application
var NewRelicApp *newrelic.Application

func SetNewRelicApp(nr *newrelic.Application) {
	NewRelicApp = nr
}

func GetNewRelicApp() *newrelic.Application {
	return NewRelicApp
}

/// Use NewRelic better - reference https://github.com/carousell/Orion/blob/19b7601394006ca4eb9dcb65a2339c2046111f75/utils/utils.go

// GetNewRelicTransactionFromContext fetches the new relic transaction that is stored in the context
// if there is no transaction in the context, it returns nil
func GetNewRelicTransactionFromContext(ctx context.Context) *newrelic.Transaction {
	return newrelic.FromContext(ctx)
}

// GetOrStartNew returns a new relic transaction from context
// if there is no transaction in the context, it starts a new transaction
func GetOrStartNew(ctx context.Context, name string) (*newrelic.Transaction, context.Context) {
	txn := GetNewRelicTransactionFromContext(ctx)
	if txn != nil {
		ctx = StartNRTransaction(name, ctx, nil, nil)
	}
	return GetNewRelicTransactionFromContext(ctx), ctx
}

// StoreNewRelicTransactionToContext stores a new relic transaction object to context
// if there is already a transaction in the context, it will be overwritten by the new one passed in the argument
func StoreNewRelicTransactionToContext(ctx context.Context, t *newrelic.Transaction) context.Context {
	return newrelic.NewContext(ctx, t)
}

// StartNRTransaction starts a new newrelic transaction
// if there is already a transaction in the context, it will start a child transaction
func StartNRTransaction(path string, ctx context.Context, req *http.Request, w http.ResponseWriter) context.Context {
	// check if transaction has been initialized
	if NewRelicApp == nil {
		return ctx
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	t := GetNewRelicTransactionFromContext(ctx)
	if t == nil {
		t = NewRelicApp.StartTransaction(path)
		if t != nil {
			if req != nil {
				t.SetWebRequestHTTP(req)
			} else {
				rl, _ := url.Parse(path)
				t.SetWebRequest(
					newrelic.WebRequest{
						Type: string(newrelic.TransportUnknown),
						URL:  rl,
					},
				)
			}
			ctx = StoreNewRelicTransactionToContext(ctx, t)
		}
	}
	return ctx
}

// FinishNRTransaction finishes an existing transaction
// if there is no transaction in the context, it does nothing
func FinishNRTransaction(ctx context.Context, err error) {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		t.NoticeError(err)
		t.End()
	}
}

// IgnoreNRTransaction ignores this NR trasaction and prevents it from being reported
// can be used to ignore health check transactions etc
func IgnoreNRTransaction(ctx context.Context) {
	t := GetNewRelicTransactionFromContext(ctx)
	if t != nil {
		t.Ignore()
	}
}
