package scache

import (
	"context"
	"log"
	"net/http"

	"github.com/a-h/scache/data"
	"github.com/a-h/scache/expiry"

	"github.com/a-h/scache/changes"

	"github.com/a-h/scache/cache"
)

// AddMiddleware adds the cache to the context on each HTTP request, which ensures
// that the cache is always up-to-date.
func AddMiddleware(next http.Handler, s expiry.Stream) http.Handler {
	return &Middleware{
		Observer: changes.NewObserver(s),
		Cache:    cache.New(),
		Next:     next,
		Logger:   log.Printf,
		Notifier: changes.NewNotifier(s),
	}
}

// Middleware is HTTP middleware that adds the cache to the HTTP context of the current request.
type Middleware struct {
	Observer *changes.Observer
	Notifier changes.Notifier
	Cache    *cache.Cache
	Next     http.Handler
	Logger   func(format string, v ...interface{})
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mw.Cache.RemoveExpired()

	if mw.Cache.Count() == 0 {
		mw.Observer.Reset()
	} else {
		toRemove, err := mw.Observer.Observe()
		if err != nil {
			mw.Logger("error observing stream: %v", err)
		}
		for _, tr := range toRemove {
			mw.Cache.Remove(tr.String())
		}
	}

	// Add the cache to the context.
	ctx := context.WithValue(r.Context(), cacheContextKey, mw.Cache)

	// Execute the handler, which can now use the Get function to retrieve items from the cache.
	mw.Next.ServeHTTP(w, r.WithContext(ctx))
}

type contextKey string

const cacheContextKey = contextKey("scache")

type cacheContextContent struct {
	Cache    *cache.Cache
	Notifier changes.Notifier
}

// Get a value from the cache, if available.
func Get(r *http.Request, key data.ID, v interface{}) (ok bool) {
	c, hasCache := r.Context().Value(cacheContextKey).(*cacheContextContent)
	if !hasCache {
		return
	}
	v, ok = c.Cache.Get(key.String())
	return
}

// Invalidate invalidates some data.
func Invalidate(r *http.Request, key data.ID) (ok bool, err error) {
	c, ok := r.Context().Value(cacheContextKey).(*cacheContextContent)
	if !ok {
		return
	}
	err = c.Notifier.NotifyDataChanged(key)
	if err != nil {
		c.Cache.Remove(key.String())
		//TODO: Log error appropriately.
		ok = false
	}
	return
}
