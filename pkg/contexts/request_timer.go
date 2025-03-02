package contexts

import (
	"context"
	"errors"
	"time"
)

// Set the start time of the request
func SetStart(ctx context.Context, time time.Time) context.Context {
	return context.WithValue(ctx, contextKeyRequestTime, time)
}

// Get the start time of the request
func GetStartTime(ctx context.Context) (time.Time, error) {
	start, ok := ctx.Value(contextKeyRequestTime).(time.Time)
	if !ok {
		return time.Time{}, errors.New("Failed to get start time of request")
	}
	return start, nil
}
