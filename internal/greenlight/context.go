package greenlight

import "context"

type ctxKey string

const (
	userIDCtxKey ctxKey = "userID"
)

func NewContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

func UserIDFromContext(ctx context.Context) int64 {
	userID, _ := ctx.Value(userIDCtxKey).(int64)
	return userID
}
