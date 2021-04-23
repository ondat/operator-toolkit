package composite

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a composite client, composed of cached and unclient clients. It
// can be used to query objects from cache and fall back to use the raw client
// and get the object directly from the API server.
type Client struct {
	client.Client
	Options

	uncached client.Client
}

// ClientOption is used to configure the client.
type Options struct {
	// RawListing is used to perform raw listing operations, uncached.
	RawListing bool
}

// NewClient creates and returns a composite Client.
func NewClient(cached client.Client, uncached client.Client, opts Options) *Client {
	return &Client{
		Client:   cached,
		Options:  opts,
		uncached: uncached,
	}
}

// Get first fetches the object using the cached client. If the object is not
// found in the cached client, it retries using the uncached client.
func (c *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	if cErr := c.Client.Get(ctx, key, obj); cErr != nil {
		// If not found in the cache, try with the uncached client.
		if apierrors.IsNotFound(cErr) {
			return c.uncached.Get(ctx, key, obj)
		}
		return cErr
	}
	return nil
}

// List lists the objects based on the client configuration. If RawListing is
// true, it uses the uncached client to list, else it uses the cached client.
func (c *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if c.RawListing {
		return c.uncached.List(ctx, list, opts...)
	}
	return c.Client.List(ctx, list, opts...)
}
