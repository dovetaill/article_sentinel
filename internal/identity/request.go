package identity

import "context"

type RequestMetadata struct {
	RequestID string `json:"request_id"`
	SourceIP  string `json:"source_ip"`
}

type requestMetadataContextKey struct{}

func ContextWithRequestMetadata(ctx context.Context, metadata RequestMetadata) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	current, _ := RequestMetadataFromContext(ctx)
	if metadata.RequestID == "" {
		metadata.RequestID = current.RequestID
	}
	if metadata.SourceIP == "" {
		metadata.SourceIP = current.SourceIP
	}

	return context.WithValue(ctx, requestMetadataContextKey{}, metadata)
}

func RequestMetadataFromContext(ctx context.Context) (RequestMetadata, bool) {
	if ctx == nil {
		return RequestMetadata{}, false
	}
	metadata, ok := ctx.Value(requestMetadataContextKey{}).(RequestMetadata)
	if !ok {
		return RequestMetadata{}, false
	}
	if metadata.RequestID == "" && metadata.SourceIP == "" {
		return RequestMetadata{}, false
	}
	return metadata, true
}
