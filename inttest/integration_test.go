package inttest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jeffreybolle/inttest-example-go/pkg/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var (
	firstName = "Jeffrey"
	lastName  = "Bolle"
	dob       = time.Date(1985, 9, 22, 0, 0, 0, 0, time.UTC)
)

func Test(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	creditScoreMock, tearDownMockCreditScoreService := startMockCreditScoreService(t, ctx, 8000)
	defer tearDownMockCreditScoreService()

	tearDownService := startService(t, ctx)
	defer tearDownService()

	client, tearDownClient := createClient(t, ctx)
	defer tearDownClient()

	t.Run("create_user", func(t *testing.T) {
		createResp, err := client.CreateUser(ctx, &api.CreateUserRequest{
			FirstName:   firstName,
			LastName:    lastName,
			DateOfBirth: dob,
		})
		require.NoError(t, err)
		getResp, err := client.GetUser(ctx, &api.GetUserRequest{
			ID: createResp.ID,
		})
		require.NoError(t, err)
		require.Equal(t, getResp.FirstName, firstName)
		require.Equal(t, getResp.LastName, lastName)
		require.Equal(t, getResp.DateOfBirth, dob)
	})
	t.Run("credit_score", func(t *testing.T) {
		createResp, err := client.CreateUser(ctx, &api.CreateUserRequest{
			FirstName:   firstName,
			LastName:    lastName,
			DateOfBirth: dob,
		})
		require.NoError(t, err)
		creditScoreMock.SetNextScore("0.87")
		getResp, err := client.GetUser(ctx, &api.GetUserRequest{
			ID: createResp.ID,
		})
		require.NoError(t, err)
		require.Equal(t, getResp.CreditScore, 0.87)
		creditScoreMock.SetNextScore("0.34")
		getResp, err = client.GetUser(ctx, &api.GetUserRequest{
			ID: createResp.ID,
		})
		require.NoError(t, err)
		require.Equal(t, getResp.CreditScore, 0.34)
	})
}

func createClient(t *testing.T, ctx context.Context) (api.APIClient, func()) {
	conn, err := grpc.DialContext(ctx, ":9000", grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)
	client := api.NewAPIClient(conn)
	return client, func() {
		_ = conn.Close()
	}
}

func startService(t *testing.T, ctx context.Context) func() {
	assertPortAvailable(t, 9000)
	assertPortAvailable(t, 9001)

	wd, err := os.Getwd()
	require.NoError(t, err)
	wd = filepath.Dir(wd)

	// build service
	command := exec.Command("go", "build", "-o", binaryName, "cmd/main.go")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = wd
	err = command.Run()
	require.NoError(t, err)

	// start service
	command = exec.Command(filepath.Join(wd, binaryName))
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Dir = wd
	command.Env = append(command.Env, "SERVICE_PORT=9000")
	command.Env = append(command.Env, "HEALTH_CHECK_PORT=9001")
	command.Env = append(command.Env, "DATASTORE_EMULATOR_HOST=0.0.0.0:8282")
	command.Env = append(command.Env, "GCP_PROJECT_ID=example")
	command.Env = append(command.Env, "CREDIT_SCORE_URL=http://localhost:8000/api/score")
	err = command.Start()
	require.NoError(t, err)

	// wait until service is up
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:9001/live", nil)
		require.NoError(t, err)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				// stop process if possible
				_ = command.Process.Signal(stopSignal)
				_, _ = command.Process.Wait()
				// clean up binary
				err = os.Remove(filepath.Join(wd, binaryName))
				require.NoError(t, err)
				t.Fatalf("context timed out before service was available")
			}
		}
		if response.StatusCode == http.StatusOK {
			break
		}
	}

	return func() {
		// stop service
		_ = command.Process.Signal(stopSignal)
		_, _ = command.Process.Wait()

		// clean up binary
		_ = os.Remove(filepath.Join(wd, binaryName))
	}
}

func assertPortAvailable(t *testing.T, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		_ = conn.Close()
		t.Fatalf("port %d is already in-use", port)
	}
}
