package changes

import "github.com/a-h/scache/data"

// Observable types have a data ID property on them.
type Observable interface {
	ObservableID() data.ID
}
