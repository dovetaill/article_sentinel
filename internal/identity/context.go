package identity

import "context"

type principalContextKey struct{}

func ContextWithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if principal, ok := ctx.Value(principalContextKey{}).(Principal); ok {
		return principal, true
	}

	actor, ok := ActorFromContext(ctx)
	if !ok {
		return Principal{}, false
	}

	return PrincipalFromActor(actor), true
}
