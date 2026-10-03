package plant

import (
	"context"
	"time"
)

type locationKey struct{}

// ContextWithLocation carries the requesting user's time zone down to the use
// cases, so clock times such as the roadmap's "08:00" mean the user's 08:00.
func ContextWithLocation(ctx context.Context, loc *time.Location) context.Context {
	if loc == nil {
		return ctx
	}
	return context.WithValue(ctx, locationKey{}, loc)
}

// LocationFrom returns the user's time zone from ctx, or UTC when none was set.
func LocationFrom(ctx context.Context) *time.Location {
	if loc, ok := ctx.Value(locationKey{}).(*time.Location); ok && loc != nil {
		return loc
	}
	return time.UTC
}
