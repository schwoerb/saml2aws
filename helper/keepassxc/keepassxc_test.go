package keepassxc

import (
	"testing"

	"github.com/versent/saml2aws/v2/helper/credentials"
)

func TestSupportsCredentialStorage(t *testing.T) {
	helper := &KeePassXCHelper{}
	if !helper.SupportsCredentialStorage() {
		t.Error("SupportsCredentialStorage() = false, want true")
	}
}

func TestNewKeePassXCHelperNoConnection(t *testing.T) {
	// This will fail if KeePassXC is not running or browser integration is disabled
	config := Configuration{}

	_, err := NewKeePassXCHelper(config)
	// We expect an error since KeePassXC is likely not running in test environment
	if err == nil {
		t.Log("NewKeePassXCHelper() succeeded - KeePassXC must be running")
	} else {
		t.Logf("NewKeePassXCHelper() error = %v (expected in test environment)", err)
	}
}

func TestAddNoClient(t *testing.T) {
	helper := &KeePassXCHelper{}
	creds := &credentials.Credentials{
		ServerURL: "https://example.com",
		Username:  "testuser",
		Secret:    "testpass",
	}

	err := helper.Add(creds)
	if err == nil {
		t.Error("Add() expected error when client is not initialized, got nil")
	}
}

func TestGetNoClient(t *testing.T) {
	helper := &KeePassXCHelper{}
	_, _, err := helper.Get("https://example.com")
	if err == nil {
		t.Error("Get() expected error when client is not initialized, got nil")
	}
}

func TestDeleteNoClient(t *testing.T) {
	helper := &KeePassXCHelper{}
	err := helper.Delete("https://example.com")
	// Delete is graceful and won't error if client is nil
	if err != nil {
		t.Logf("Delete() returned error: %v", err)
	}
}

func TestClose(t *testing.T) {
	helper := &KeePassXCHelper{}
	err := helper.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// Integration test - only runs if KeePassXC is available and running
func TestKeePassXCIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := Configuration{}
	helper, err := NewKeePassXCHelper(config)
	if err != nil {
		t.Skipf("Skipping integration test - KeePassXC not available: %v", err)
		return
	}
	defer helper.Close()

	testCreds := &credentials.Credentials{
		ServerURL: "https://test.example.com/saml",
		Username:  "testuser",
		Secret:    "testpassword",
	}

	// Test Add
	t.Run("Add", func(t *testing.T) {
		err := helper.Add(testCreds)
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		username, password, err := helper.Get(testCreds.ServerURL)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if username != testCreds.Username {
			t.Errorf("Get() username = %q, want %q", username, testCreds.Username)
		}
		if password != testCreds.Secret {
			t.Errorf("Get() password = %q, want %q", password, testCreds.Secret)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := helper.Delete(testCreds.ServerURL)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
	})

	// Verify deletion
	t.Run("GetAfterDelete", func(t *testing.T) {
		_, _, err := helper.Get(testCreds.ServerURL)
		if err != credentials.ErrCredentialsNotFound {
			t.Errorf("Get() after delete error = %v, want ErrCredentialsNotFound", err)
		}
	})
}
