package main_test

import (
	"fmt"
	"net"
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
	// Pick a random available port for most tests
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err, "could not get a free port for test")
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	os.Setenv("PORT", fmt.Sprintf("%d", port))

	var session, cleanup = buildAndRun(t)
	defer cleanup()

	t.Run("is started and listening on custom port", func(t *testing.T) {
		assert.Eventually(t, func() bool {
			output := session.Err.Contents()
			return strings.Contains(string(output), fmt.Sprintf("Starting server on :%d", port))
		}, 2*time.Second, 10*time.Millisecond, "server did not start in time")
	})

	t.Run("answers to the health endpoint", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
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

	t.Run("fails to start if port is busy", func(t *testing.T) {
		// Start a dummy listener on port 8080
		ln, err := net.Listen("tcp", ":8080")
		require.NoError(t, err, "could not listen on port 8080 for test setup")
		defer ln.Close()
		os.Setenv("PORT", "8080")

		// Attempt to start the service using buildAndRun
		session, cleanup := buildAndRun(t)
		defer cleanup()

		// The process should exit quickly with an error about the port
		assert.Eventually(t, func() bool {
			select {
			case <-session.Exited:
				return session.ExitCode() != 0
			default:
				return false
			}
		}, 2*time.Second, 10*time.Millisecond, "service did not exit as expected when port is busy")
		output := string(session.Err.Contents()) + string(session.Out.Contents())
		assert.Contains(t, output, "address already in use", "expected error about port being busy")
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
