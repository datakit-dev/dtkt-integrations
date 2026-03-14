package pkg

import (
	"context"

	fb "github.com/huandu/facebook/v2"
)

type (
	sessionCtxKey struct{}
)

func AddSessionToContext(ctx context.Context, session *fb.Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey{}, session)
}

func SessionFromContext(ctx context.Context) (*fb.Session, bool) {
	if session, ok := ctx.Value(sessionCtxKey{}).(*fb.Session); ok {
		return session, ok
	} else {
		return nil, false
	}
}
