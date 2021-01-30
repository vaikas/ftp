// +build tools

package hack

// This package imports things required by this repository, to force `go mod` to see them as dependencies
import (
	_ "knative.dev/eventing/test/test_images/recordevents"
)
