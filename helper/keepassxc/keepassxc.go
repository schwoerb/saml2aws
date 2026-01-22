package keepassxc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/versent/saml2aws/v2/helper/credentials"
	"github.com/xakep666/gkpxc"
)

var logger = logrus.WithField("helper", "keepassxc")

const (
	timeout = 30 * time.Second
)

// KeePassXCHelper handles secrets using KeePassXC via the browser integration protocol.
type KeePassXCHelper struct {
	client *gkpxc.Client
}

type Configuration struct {
	// DatabaseID is optional - if empty, will use the first available database
	DatabaseID string
}

// NewKeePassXCHelper creates a new KeePassXC helper using the browser integration protocol.
func NewKeePassXCHelper(config Configuration) (*KeePassXCHelper, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create client connection to KeePassXC
	client, err := gkpxc.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create KeePassXC client: %w (ensure KeePassXC is running with browser integration enabled)", err)
	}

	// Try to load existing association credentials
	if creds, err := loadAssociationCredentials(); err == nil {
		client.SetAssociationCredentials(creds)
		logger.Debug("loaded existing association credentials")

		// Test if credentials still work
		if err := client.TestAssociate(ctx); err == nil {
			logger.Debug("successfully connected to KeePassXC using cached credentials")
			return &KeePassXCHelper{
				client: client,
			}, nil
		}
		logger.Debug("cached credentials invalid, re-associating")
	}

	// Associate with the database (either first time or credentials expired)
	if err := client.Associate(ctx); err != nil {
		return nil, fmt.Errorf("failed to associate with KeePassXC: %w", err)
	}

	// Save the association credentials for future use
	if err := saveAssociationCredentials(client.AssociationCredentials()); err != nil {
		logger.WithError(err).Warn("failed to save association credentials")
	}

	logger.Debug("successfully connected to KeePassXC")

	return &KeePassXCHelper{
		client: client,
	}, nil
}

// Add adds new credentials to KeePassXC.
func (k *KeePassXCHelper) Add(creds *credentials.Credentials) error {
	if k.client == nil {
		return fmt.Errorf("KeePassXC client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Check if an entry already exists for this URL
	var uuid string
	getReq := gkpxc.GetLoginsRequest{
		URL: creds.ServerURL,
	}

	logins, err := k.client.GetLogins(ctx, getReq)
	if err == nil && len(logins.Entries) > 0 {
		// Use existing entry's UUID to update it
		uuid = logins.Entries[0].UUID
		logger.WithField("uuid", uuid).Debug("updating existing entry")
	} else {
		// No existing entry, will create new one
		logger.Debug("creating new entry")
	}

	// Encode full credentials as JSON for storage in a custom field
	encoded, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to encode credentials: %w", err)
	}

	// Create or update login entry
	req := gkpxc.SetLoginRequest{
		URL:      creds.ServerURL,
		Login:    creds.Username,
		Password: creds.Secret,
		Group:    "saml2aws",
		UUID:     uuid, // Empty UUID creates new entry, existing UUID updates entry
	}

	// Set (create/update) the login
	if err := k.client.SetLogin(ctx, req); err != nil {
		logger.WithError(err).Error("failed to add credentials to KeePassXC")
		return fmt.Errorf("failed to add credentials: %w", err)
	}

	// Log the encoded credentials for debugging (note: stored separately if needed)
	_ = encoded

	logger.WithField("url", creds.ServerURL).Debug("credentials added to KeePassXC")
	return nil
}

// Delete removes credentials from KeePassXC.
func (k *KeePassXCHelper) Delete(serverURL string) error {
	if k.client == nil {
		return nil // Gracefully handle uninitialized client
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get the entry first to find its UUID
	req := gkpxc.GetLoginsRequest{
		URL: serverURL,
	}

	logins, err := k.client.GetLogins(ctx, req)
	if err != nil {
		logger.WithError(err).Debug("failed to find credentials for deletion")
		// Not necessarily an error - entry might not exist
		return nil
	}

	if len(logins.Entries) == 0 {
		logger.Debug("no credentials found to delete")
		return nil
	}

	// Delete all matching logins (usually just one)
	for _, login := range logins.Entries {
		delReq := gkpxc.DeleteEntryRequest{
			UUID: login.UUID,
		}
		if err := k.client.DeleteEntry(ctx, delReq); err != nil {
			logger.WithError(err).WithField("uuid", login.UUID).Warn("failed to delete entry")
		} else {
			logger.WithField("uuid", login.UUID).Debug("deleted entry")
		}
	}

	logger.WithField("url", serverURL).Debug("credentials deleted from KeePassXC")
	return nil
}

// Get retrieves credentials from KeePassXC.
func (k *KeePassXCHelper) Get(serverURL string) (string, string, error) {
	if k.client == nil {
		return "", "", fmt.Errorf("KeePassXC client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.WithField("url", serverURL).Debug("retrieving credentials from KeePassXC")

	// Retrieve logins for the URL
	req := gkpxc.GetLoginsRequest{
		URL: serverURL,
	}

	logins, err := k.client.GetLogins(ctx, req)
	if err != nil {
		logger.WithError(err).Error("failed to retrieve credentials from KeePassXC")
		return "", "", credentials.ErrCredentialsNotFound
	}

	if len(logins.Entries) == 0 {
		logger.Debug("no credentials found in KeePassXC")
		return "", "", credentials.ErrCredentialsNotFound
	}

	// Use the first matching login
	login := logins.Entries[0]

	fmt.Printf("\033[1m✓ Found credentials in KeePassXC for username: %s\033[0m\n", login.Login)
	logger.WithField("username", login.Login).Debug("credentials retrieved from KeePassXC")
	return login.Login, login.Password, nil
}

// SupportsCredentialStorage returns true since KeePassXC supports credential storage.
func (KeePassXCHelper) SupportsCredentialStorage() bool {
	return true
}

// Close closes the connection to KeePassXC.
func (k *KeePassXCHelper) Close() error {
	if k.client != nil {
		return k.client.Close()
	}
	return nil
}

// saveAssociationCredentials saves the KeePassXC association credentials to disk.
func saveAssociationCredentials(creds *gkpxc.AssociationCredentials) error {
	if creds == nil {
		return fmt.Errorf("credentials are nil")
	}

	filePath, err := getCredentialsFilePath()
	if err != nil {
		return fmt.Errorf("failed to get credentials file path: %w", err)
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restrictive permissions (0600 = rw-------)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	logger.WithField("path", filePath).Debug("saved association credentials")
	return nil
}

// loadAssociationCredentials loads the KeePassXC association credentials from disk.
func loadAssociationCredentials() (*gkpxc.AssociationCredentials, error) {
	filePath, err := getCredentialsFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials file path: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds gkpxc.AssociationCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	logger.WithField("path", filePath).Debug("loaded association credentials")
	return &creds, nil
}

// CredentialsFileExists checks if the KeePassXC credentials file exists.
func CredentialsFileExists() bool {
	filePath, err := getCredentialsFilePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(filePath)
	return err == nil
}

// getCredentialsFilePath returns the path to the KeePassXC credentials file.
// It uses the SAML2AWS_CONFIGFILE environment variable or defaults to ~/.saml2aws,
// then appends "_keepassxc" to create the credentials file name.
func getCredentialsFilePath() (string, error) {
	// Get config file from environment variable or use default
	configFile := os.Getenv("SAML2AWS_CONFIGFILE")
	if configFile == "" {
		configFile = "~/.saml2aws"
	}

	// Expand home directory
	configPath, err := homedir.Expand(configFile)
	if err != nil {
		return "", fmt.Errorf("failed to expand config path: %w", err)
	}

	// Append _keepassxc to the config file path
	credentialsPath := configPath + "_keepassxc"
	return credentialsPath, nil
}
