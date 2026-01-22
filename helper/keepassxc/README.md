# KeePassXC Credential Helper

This helper provides cross-platform credential storage integration with KeePassXC using the browser integration protocol.

**Supported Platforms**: Linux, macOS, Windows

## Prerequisites

1. **KeePassXC** must be installed on your system
2. **KeePassXC must be running** with browser integration enabled
3. A KeePassXC database must be unlocked

### Installation

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install keepassxc
```

**Linux (Fedora/RHEL):**
```bash
sudo dnf install keepassxc
```

**Linux (Arch):**
```bash
sudo pacman -S keepassxc
```

**macOS:**
```bash
brew install keepassxc
```

**Windows:**
Download from https://keepassxc.org/download/

## Configuration

### Step 1: Enable Browser Integration in KeePassXC

1. Open KeePassXC
2. Go to **Tools** → **Settings** → **Browser Integration**
3. Check **Enable browser integration**
4. Make sure **Enable integration for these browsers** includes the option you need

### Step 2: Enable the Helper

**Important:** KeePassXC can only be enabled via environment variable. It is **not** configurable through the `.saml2aws` configuration file.

Set the following environment variable to enable KeePassXC credential storage:

```bash
export SAML2AWS_KEEPASSXC_ENABLED=true
```

To make this permanent, add it to your shell profile (e.g., `~/.bashrc` or `~/.zshrc`).

#### Auto-Enable After First Use

After the first successful connection to KeePassXC, saml2aws saves association credentials to a file based on your config file location. The credentials file name is created by appending `_keepassxc` to your config file path:

- Default: `~/.saml2aws_keepassxc`
- Custom config: If you use `--config /path/to/myconfig` or set `SAML2AWS_CONFIGFILE=/path/to/myconfig`, the credentials will be saved to `/path/to/myconfig_keepassxc`

Once this file exists, KeePassXC will be **automatically enabled** by default, even without setting the environment variable.

To disable KeePassXC when the credentials file exists:

```bash
export SAML2AWS_KEEPASSXC_ENABLED=false
```

This auto-enable behavior:
- Simplifies setup after initial configuration
- Eliminates need for the user to re-approve the association on each run
- Can be explicitly disabled if needed
- Respects your config file location for proper multi-profile support

## Usage

Once configured, saml2aws will automatically use KeePassXC to store and retrieve credentials.

### First-time setup

1. Create a KeePassXC database if you don't have one (via GUI or CLI)

2. **Start KeePassXC and unlock your database**

3. **Enable browser integration** (see Configuration step 1 above)

4. Set the environment variable:
   ```bash
   export SAML2AWS_KEEPASSXC_ENABLED=true
   ```

5. Run saml2aws as normal:
   ```bash
   saml2aws login
   ```

6. On first use, KeePassXC will prompt you to allow saml2aws to access the database

### How it works

- Uses KeePassXC's browser integration protocol (same protocol used by browser extensions)
- Credentials are stored as login entries in your KeePassXC database
- Entries are automatically grouped under "saml2aws"
- The URL field stores the ServerURL
- The username and password are stored in standard fields
- Full credential metadata is stored as JSON in a custom field for future compatibility
- **KeePassXC must be running** for credential operations

### Security

- Your KeePassXC database is protected by your master password
- Credentials are encrypted using KeePassXC's strong encryption
- No credentials are stored in plain text

## Troubleshooting

### "failed to connect to KeePassXC"

**Cause:** KeePassXC is not running or browser integration is disabled.

**Solution:**
1. Start KeePassXC
2. Unlock your database
3. Enable browser integration: **Tools** → **Settings** → **Browser Integration**

### "failed to associate with KeePassXC"

**Cause:** saml2aws needs permission to access KeePassXC.

**Solution:**
When prompted by KeePassXC, click "Allow" to grant saml2aws access to your database.

### Credentials not found

**Causes:**
- KeePassXC is not running
- Database is locked
- No credentials have been stored yet

**Solution:**
1. Ensure KeePassXC is running and database is unlocked
2. Run `saml2aws configure` and `saml2aws login` to store credentials

### KeePassXC not running in WSL

The browser integration protocol requires KeePassXC to be running. In WSL:
- Option 1: Run KeePassXC on Windows and use the Windows socket (may require additional setup)
- Option 2: Run KeePassXC within WSL using an X server
- Option 3: Use the fallback keyring helper instead

## Configuration Notes

### Environment Variable Only

KeePassXC credential storage is enabled **exclusively** via the `SAML2AWS_KEEPASSXC_ENABLED` environment variable. Unlike other saml2aws settings that can be configured in the `.saml2aws` INI file (such as `username`, `provider`, `mfa`, etc.), the KeePassXC helper cannot be enabled through the configuration file.

This is by design, as the credential helper selection happens during application initialization, before the configuration file is loaded.

### Per-Profile Configuration

Since the environment variable is global, all saml2aws profiles will use KeePassXC when it's enabled. If you need different credential storage methods for different profiles, you can:

1. Use separate shell sessions with different environment variable settings
2. Temporarily unset the variable for specific commands: `unset SAML2AWS_KEEPASSXC_ENABLED && saml2aws login --profile=dev`
3. Use wrapper scripts to set/unset the variable as needed

## Fallback Behavior

If KeePassXC initialization fails, saml2aws will automatically fall back to the default Linux keyring helper (libsecret/kwallet/pass).
