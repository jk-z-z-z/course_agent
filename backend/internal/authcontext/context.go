package authcontext

import "context"

type userIDKey struct{}

func WithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(userIDKey{}).(uint64)
	return userID, ok
}
