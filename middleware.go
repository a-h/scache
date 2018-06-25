package scache

import (
	"context"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/a-h/scache/data"
	"github.com/a-h/scache/expiry"

	"github.com/a-h/scache/changes"

	"github.com/a-h/scache/cache"
)

// AddMiddleware adds the cache to the context on each HTTP request, which ensures
// that the cache is always up-to-date.
func AddMiddleware(next http.Handler, s expiry.Stream, minCacheDuration, maxCacheDuration time.Duration) http.Handler {
	c := cache.New()
	c.Expiration = cache.ExpireBetween(minCacheDuration, maxCacheDuration)
	return &Middleware{
		Observer: changes.NewObserver(s),
		Cache:    c,
		Next:     next,
		Notifier: changes.NewNotifier(s),
	}
}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

var logger = logrus.WithField("pkg", "github.com/a-h/scache")

// Middleware is HTTP middleware that adds the cache to the HTTP context of the current request.
type Middleware struct {
	Observer *changes.Observer
	Notifier changes.Notifier
	Cache    *cache.Cache
	Next     http.Handler
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := time.Now()

	mw.Cache.RemoveExpired()

	if mw.Cache.Count() == 0 {
		// There's a chance that something could have snuck into the cache between
		// removing expired records, and reading the count, which means that sometimes
		// we might update from the stream when we didn't really need to, but that's
		// better than having a global lock.
		mw.Observer.Reset()
	} else {
		toRemove, err := mw.Observer.Observe()
		if err != nil {
			logger.WithError(err).Error("error observing stream")
		}
		for _, tr := range toRemove {
			mw.Cache.Remove(tr.String())
		}
	}

	// Add the cache content to the context.
	ccc := cacheContextContent{
		Cache:    mw.Cache,
		Notifier: mw.Notifier,
	}
	ctx := context.WithValue(r.Context(), cacheContextKey, &ccc)

	// Execute the handler, which can now use the Get function to retrieve items from the cache.
	mw.Next.ServeHTTP(w, r.WithContext(ctx))

	timeSpent := time.Now().Sub(st)
	logger.
		WithField("timeSpent", timeSpent).
		WithField("timeSaved", ccc.TimeSaved).
		Info("complete")
}

type contextKey string

const cacheContextKey = contextKey("scache")

type cacheContextContent struct {
	Cache     *cache.Cache
	Notifier  changes.Notifier
	TimeSaved time.Duration
}

// Get a value from the cache, if available.
func Get(r *http.Request, key data.ID, v interface{}) (ok bool) {
	c, hasCache := r.Context().Value(cacheContextKey).(*cacheContextContent)
	if !hasCache {
		return
	}
	var timeSaved time.Duration
	v, timeSaved, ok = c.Cache.GetWithDuration(key.String())
	c.TimeSaved += timeSaved
	return
}

// Add a value to the cache.
func Add(r *http.Request, key data.ID, v interface{}) (ok bool) {
	return AddWithDuration(r, key, v, time.Duration(0))
}

// AddWithDuration adds a value to the cache, while recording how much time it would save
// each time it's retrieved from the cache.
func AddWithDuration(r *http.Request, key data.ID, v interface{}, d time.Duration) (ok bool) {
	c, hasCache := GetCacheFromContext(r.Context())
	if !hasCache {
		return
	}
	c.PutWithDuration(key.String(), v, d)
	ok = true
	return
}

// GetCacheFromContext gets the cache object from the context. Should be used when wanting to customise
// expiration of cache items or to use the cache directly.
func GetCacheFromContext(ctx context.Context) (c *cache.Cache, ok bool) {
	ccc, ok := ctx.Value(cacheContextKey).(*cacheContextContent)
	if !ok {
		return
	}
	c = ccc.Cache
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
		logger.WithError(err).Error("error notifying on data changed")
		c.Cache.Remove(key.String())
		ok = false
	}
	return
}
