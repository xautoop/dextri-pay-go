package integration_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xautoop/dextri-pay-go/channels"
	"github.com/xautoop/dextri-pay-go/client"
)

func TestSandboxChannelsSmoke(t *testing.T) {
	if os.Getenv("DEXTRI_PAY_RUN_SANDBOX") != "1" {
		t.Skip("set DEXTRI_PAY_RUN_SANDBOX=1 to run the live Sandbox smoke test")
	}

	pay, err := client.New(client.Config{
		BaseURL: requiredEnv(t, "DEXTRI_PAY_SANDBOX_BASE_URL"),
		Credentials: client.Credentials{
			AppID:  requiredEnv(t, "DEXTRI_PAY_SANDBOX_APP_ID"),
			KeyID:  requiredEnv(t, "DEXTRI_PAY_SANDBOX_KEY_ID"),
			Secret: requiredEnv(t, "DEXTRI_PAY_SANDBOX_APP_SECRET"),
		},
		UserAgent: "dextri-pay-go-sandbox-smoke",
	}, client.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, response, err := pay.Channels.List(ctx, channels.ListParams{})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected Sandbox response: %#v", response)
	}
}

func requiredEnv(t *testing.T, name string) string {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		t.Fatalf("%s is required when DEXTRI_PAY_RUN_SANDBOX=1", name)
	}
	return value
}
