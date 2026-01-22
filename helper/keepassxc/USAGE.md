# KeePassXC Helper - Quick Start Guide

## Overview

The KeePassXC helper provides secure credential storage for saml2aws using KeePassXC's browser integration protocol. This protocol is the same one used by browser extensions to communicate with KeePassXC, providing secure and reliable credential management.

## Quick Setup

### 1. Install KeePassXC

**Ubuntu/Debian:**
```bash
sudo apt-get update && sudo apt-get install -y keepassxc
```

**Fedora/RHEL:**
```bash
sudo dnf install keepassxc
```

**Arch Linux:**
```bash
sudo pacman -S keepassxc
```

### 2. Create and Open a Database

If you don't already have a KeePassXC database:

1. Launch KeePassXC GUI
2. Create a new database: **Database** → **New Database**
3. Set a strong master password
4. Keep the database unlocked

### 3. Enable Browser Integration

In KeePassXC:
1. Go to **Tools** → **Settings** → **Browser Integration**
2. Check **Enable browser integration**
3. Leave other settings at default

### 4. Configure saml2aws

**Important:** KeePassXC can only be enabled via environment variable. It is **not** configurable through the `.saml2aws` configuration file.

Set the environment variable to enable KeePassXC:

```bash
export SAML2AWS_KEEPASSXC_ENABLED=true
```

To make this permanent, add it to your shell profile:

**For bash (~/.bashrc):**
```bash
echo 'export SAML2AWS_KEEPASSXC_ENABLED=true' >> ~/.bashrc
source ~/.bashrc
```

**For zsh (~/.zshrc):**
```bash
echo 'export SAML2AWS_KEEPASSXC_ENABLED=true' >> ~/.zshrc
source ~/.zshrc
```

### 5. Use saml2aws Normally

Now when you use saml2aws, credentials will be stored in KeePassXC:

```bash
saml2aws configure
saml2aws login
```

**Important:** 
- KeePassXC must be running with an unlocked database
- On first use, KeePassXC will prompt you to allow "saml2aws" access - click **Allow**
- After the first successful connection, association credentials are saved to `<config_file>_keepassxc` (e.g., `~/.saml2aws_keepassxc`)
- The credentials file location follows your config file (set via `--config` or `SAML2AWS_CONFIGFILE`)
- Once this file exists, KeePassXC will be automatically enabled (no need to set the environment variable)
- To disable auto-enable, set `SAML2AWS_KEEPASSXC_ENABLED=false`

## How It Works

### Storage Format

- **Group**: Entries are stored in a "saml2aws" group for organization
- **URL**: The ServerURL is stored in the URL field
- **Username**: Stored in the standard username field
- **Password**: Stored as the entry password
- **Metadata**: Full credential JSON stored in a custom field for future compatibility

### Security Features

- ✅ Strong encryption (AES-256)
- ✅ Master password protection
- ✅ Secure browser integration protocol
- ✅ Association requires explicit user approval
- ✅ Industry-standard password manager
- ✅ Same protocol used by browser extensions

## Common Workflows

### First Login

```bash
# Configure your IDP
saml2aws configure

# Login (credentials will be stored in KeePassXC)
saml2aws login
```

### Subsequent Logins

```bash
# Credentials are retrieved from KeePassXC automatically
saml2aws login
```

You'll only be prompted for your KeePassXC master password, not your IDP credentials.

### Managing Credentials

All credentials are managed through KeePassXC:

1. Open KeePassXC
2. Look for the "saml2aws" group
3. View, edit, or delete entries as needed using the KeePassXC GUI

Credentials are automatically organized by URL for easy identification.

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SAML2AWS_KEEPASSXC_ENABLED` | Enable KeePassXC helper (cannot be set in `.saml2aws` config file) | `true` or `1` |

## Advantages Over Other Options

### vs. Disabling Keychain
- ✅ Credentials are securely stored (no re-entering on each login)
- ✅ Industry-standard encryption

### vs. Pass
- ✅ Simpler setup (no GPG key management)
- ✅ GUI available for credential management
- ✅ Better cross-platform support

### vs. libsecret/gnome-keyring
- 🟡 Requires KeePassXC to be running
- ✅ More user-friendly interface
- ✅ Better visibility into stored credentials
- 🟡 May require X11/Wayland for GUI (but works great in WSL with Windows KeePassXC)

## Integration with Existing KeePassXC Database

If you already use KeePassXC, simply enable the helper:

```bash
export SAML2AWS_KEEPASSXC_ENABLED=true
```

Your saml2aws credentials will be stored in a dedicated "saml2aws" group alongside your other passwords, keeping everything organized.

## Troubleshooting

### Issue: "failed to connect to KeePassXC"

**Solution:** 
1. Make sure KeePassXC is running
2. Check that a database is unlocked
3. Enable browser integration in KeePassXC settings

### Issue: "failed to associate with KeePassXC"

**Solution:** KeePassXC will show a popup asking for permission. Click "Allow" to grant saml2aws access.

### Issue: Connection works but credentials not found

**Solution:** 
1. Make sure you've run `saml2aws configure` and `saml2aws login` at least once to store credentials
2. Check that the database is unlocked
3. Look in KeePassXC GUI for the "saml2aws" group

### Issue: Falls back to keyring

**Solution:** Verify the environment variable is set:
```bash
echo $SAML2AWS_KEEPASSXC_ENABLED
```

If empty, set it to `true` and restart your shell.

## Advanced: Using with Multiple Databases

You can use different KeePassXC databases for different environments by switching which database is unlocked in KeePassXC:

```bash
# Simply switch databases in KeePassXC GUI
# The helper will use whichever database is currently unlocked
export SAML2AWS_KEEPASSXC_ENABLED=true
saml2aws login --profile=prod

# Or temporarily disable for testing
unset SAML2AWS_KEEPASSXC_ENABLED
saml2aws login --profile=dev
```

## Security Best Practices

1. **Use a strong master password** for your KeePassXC database
2. **Back up your database** regularly
3. **Lock your database** when not in use (KeePassXC can auto-lock)
4. **Review access permissions** in KeePassXC settings periodically
5. **Consider using a key file** or hardware key in addition to password for enhanced security
6. **Keep KeePassXC updated** for the latest security fixes

## Support

For issues specific to the KeePassXC helper, please file an issue on the saml2aws GitHub repository with the `[keepassxc]` tag in the title.
