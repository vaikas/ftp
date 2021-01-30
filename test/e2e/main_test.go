// +build e2e

package e2e

import (
	"os"
	"testing"

	testlib "knative.dev/eventing/test/lib"
	"knative.dev/pkg/system"
)

var setup = testlib.Setup
var tearDown = testlib.TearDown

func TestMain(m *testing.M) {
	exit := m.Run()

	testlib.ExportLogs(testlib.SystemLogsDir, system.Namespace())

	os.Exit(exit)
}
