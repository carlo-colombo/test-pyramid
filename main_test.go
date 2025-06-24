package main_test

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2/dsl/core"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_myservice(t *testing.T) {
	var session, cleanup = buildAndRun(t)
	defer cleanup()

	t.Run("is started and listening on port 8080", func(t *testing.T) {
		assert.Eventually(t, func() bool {
			output := session.Err.Contents()
			return strings.Contains(string(output), "Starting server on :8080")
		}, 10*time.Second, 10*time.Millisecond, "server did not start in time")
	})

	t.Run("answers to the health endpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/health")

		assert.NoError(t, err, "cannot perform request")
		assert.Equal(t, 200, resp.StatusCode, "expected 200 status code")
	})

	t.Run("shuts down with interrupt signal", func(t *testing.T) {
		session.Interrupt()
		assert.Eventually(t, func() bool {
			select {
			case <-session.Exited:
				return true
			default:
				return false
			}
		}, 2*time.Second, 10*time.Millisecond, "process did not exit in time after interrupt")
	})
}

// build and run the test subjects
// returns a session that could be used to interact with the service (stdin) or
// to check the output. It also returns a cleanup function that should be deferred at the
// end of the test cases.
func buildAndRun(t *testing.T) (*gexec.Session, func()) {
	t.Helper()

	//change the following to your root module
	pathToBin, err := gexec.Build("github.com/carlo-colombo/test-pyramid", "-cover")

	require.NoError(t, err, "failed to build the service")

	cmd := exec.Command(pathToBin)

	cmd.Env = append(cmd.Env, os.Environ()...)

	session, err := gexec.Start(cmd,
		gexec.NewPrefixedWriter("[out] ", core.GinkgoWriter),
		gexec.NewPrefixedWriter("[err] ", core.GinkgoWriter))

	require.NoError(t, err, "failed to start service")

	return session, func() {
		gexec.Kill()
		gexec.CleanupBuildArtifacts()
	}
}
