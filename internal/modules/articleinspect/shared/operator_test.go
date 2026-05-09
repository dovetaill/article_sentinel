package shared

import (
	"context"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

func TestOperatorResolverUsesActorAndRequestMetadata(t *testing.T) {
	actor := identity.NewActor(23, "jwt-user", "reviewer", "active")
	ctx := identity.ContextWithActor(context.Background(), actor)
	ctx = identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(actor))
	ctx = identity.ContextWithRequestMetadata(ctx, identity.RequestMetadata{RequestID: "req-123", SourceIP: "203.0.113.10"})

	operator := ResolveOperator(ctx)
	if operator.ID != 23 || operator.Name != "jwt-user" || operator.Role != "reviewer" {
		t.Fatalf("ResolveOperator() identity = %+v, want actor fields", operator)
	}
	if operator.RequestID != "req-123" || operator.SourceIP != "203.0.113.10" {
		t.Fatalf("ResolveOperator() audit metadata = %+v, want request id and ip", operator)
	}
}
