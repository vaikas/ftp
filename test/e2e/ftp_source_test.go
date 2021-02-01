// +build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/vaikas/ftp/test/e2e/config/source"
	"github.com/vaikas/ftp/test/e2e/config/sourceproducer"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"


	_ "knative.dev/pkg/system/testing"
)

func DirectSourceTest() *feature.Feature {
	f := new(feature.Feature)

	f.Setup("install FTP source", source.Install())
	f.Alpha("FTP source").Must("goes ready", AllGoReady)

	f.Setup("install producer", sourceproducer.Install())
	f.Alpha("FTP source").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t *testing.T) {
				eventshub.StoreFromContext(ctx, "recorder").AssertAtLeast(5)
			})

	return f
}

