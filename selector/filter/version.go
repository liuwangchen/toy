package filter

import (
	"context"

	"github.com/liuwangchen/toy/selector"
)

// Version is version filter.
func Version(version string) selector.Filter {
	return func(_ context.Context, nodes []selector.Node) []selector.Node {
		newNodes := nodes[:0]
		for _, n := range nodes {
			if n.Version() == version {
				newNodes = append(newNodes, n)
			}
		}
		return newNodes
	}
}
