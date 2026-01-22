package commands

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/versent/saml2aws/v2/helper/credentials"
	"github.com/versent/saml2aws/v2/helper/keepassxc"
	"github.com/versent/saml2aws/v2/helper/wincred"
)

func init() {
	// Check if KeePassXC should be used
	useKeePassXC := os.Getenv("SAML2AWS_KEEPASSXC_ENABLED")

	// If credentials file exists and env var is not explicitly set to false, default to enabled
	shouldUseKeePassXC := false
	if useKeePassXC == "true" || useKeePassXC == "1" {
		shouldUseKeePassXC = true
	} else if useKeePassXC != "false" && useKeePassXC != "0" && keepassxc.CredentialsFileExists() {
		// Credentials file exists and env var not explicitly disabled
		shouldUseKeePassXC = true
		logrus.Debug("Found KeePassXC credentials file, defaulting to enabled")
	}

	if shouldUseKeePassXC {
		kpxcConfig := keepassxc.Configuration{}

		keepassxcHelper, err := keepassxc.NewKeePassXCHelper(kpxcConfig)
		if err != nil {
			logrus.WithField("err", err).Warn("Failed to initialize KeePassXC helper - trying fallback to wincred")
		} else {
			credentials.CurrentHelper = keepassxcHelper
			logrus.Debug("Using KeePassXC credential helper")
			return
		}
	}

	// Fallback to Windows Credential Manager
	credentials.CurrentHelper = &wincred.Wincred{}
}
