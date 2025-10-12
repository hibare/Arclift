# Arclift

[![Go Version](https://img.shields.io/github/go-mod/go-version/hibare/arclift)](https://go.dev/)
[![License](https://img.shields.io/github/license/hibare/arclift)](https://github.com/hibare/arclift/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/hibare/arclift)](https://github.com/hibare/arclift/releases)

A scheduled backup solution for directories to S3-compatible storage with optional encryption and Discord notifications.

## Features

- 📦 **S3-Compatible Storage** - Backup to Amazon S3, MinIO, or any S3-compatible storage
- 🗜️ **Archive Support** - Compress directories as tar.gz archives before upload
- 🔐 **GPG Encryption** - Optional GPG encryption for secure backups
- ⏰ **Scheduled Backups** - Cron-based automatic backup scheduling
- 🧹 **Automatic Retention** - Configurable retention policy with automatic purging of old backups
- 🔔 **Discord Notifications** - Get notified on backup success, failure, or deletion errors
- ⚙️ **Flexible Configuration** - Configure via YAML file or environment variables
- 🔄 **Version Checking** - Automatic update availability notifications

## Installation

### Binary

Download the latest binary from the [releases page](https://github.com/hibare/arclift/releases).

### From Source

```bash
git clone https://github.com/hibare/arclift.git
cd arclift
make init
go build -o arclift
```

### Debian/Ubuntu

#### Using APT

Add the Hibare repository and install Arclift:

```bash
# Download and import the GPG key
sudo wget -O /usr/share/keyrings/hibare-keyring.gpg https://apt.hibare.in/gpg.key

# Add the repository
echo "deb [signed-by=/usr/share/keyrings/hibare-keyring.gpg] https://apt.hibare.in/ * *" | sudo tee /etc/apt/sources.list.d/hibare.list

# Update package list and install
sudo apt update
sudo apt install arclift
```

#### Using Ansible

```yaml
- name: Install Arclift
  hosts: all
  become: true
  tasks:
    - name: Download and import the Hibare repository
      get_url:
        url: https://apt.hibare.in/gpg.key
        dest: /usr/share/keyrings/hibare-keyring.gpg
        mode: "0644"

    - name: Add Hibare repository source file to sources list
      copy:
        dest: /etc/apt/sources.list.d/hibare.list
        content: "deb [signed-by=/usr/share/keyrings/hibare-keyring.gpg] https://apt.hibare.in/ * *"

    - name: Update package list
      apt:
        update_cache: yes

    - name: Install Arclift
      apt:
        name: arclift
        state: present

    - name: Copy configuration file to /etc/arclift
      copy:
        src: config/arclift-config.yaml
        dest: /etc/arclift/config.yaml

    - name: Restart Arclift service
      service:
        name: arclift
        state: restarted

    - name: Ensure Arclift service is enabled and started
      systemd:
        name: arclift
        state: started
        enabled: yes
      register: arclift_service
```

## Quick Start

1. **Initialize Configuration**

   ```bash
   arclift config init -c /path/to/config.yaml
   ```

   This creates a default configuration file at the specified location.

2. **Edit Configuration**

   Edit the generated config file with your S3 credentials and backup settings.

3. **Run Backup Service**

   ```bash
   arclift -c /path/to/config.yaml
   ```

   This starts the backup scheduler based on your cron configuration.

## Configuration

### Configuration File

The configuration file uses YAML format and supports the following structure:

```yaml
s3:
  endpoint: "" # S3 endpoint URL (leave empty for AWS S3)
  region: "us-east-1" # S3 region
  access-key: "" # S3 access key
  secret-key: "" # S3 secret key
  bucket: "" # S3 bucket name
  prefix: "" # Prefix for backup keys

backup:
  dirs:
    - /path/to/backup1
    - /path/to/backup2
  hostname: "my-host" # Hostname identifier for backups
  retention-count: 30 # Number of backups to retain
  date-time-layout: "20060102150405" # Datetime format for backup keys
  cron: "0 0 * * *" # Backup schedule (daily at midnight)
  archive-dirs: false # Archive directories as tar.gz
  encryption:
    enabled: false # Enable GPG encryption (requires archive-dirs: true)
    gpg:
      key-server: "keyserver.ubuntu.com"
      key-id: "" # GPG key ID for encryption

notifiers:
  enabled: false
  discord:
    enabled: false
    webhook: "" # Discord webhook URL

logger:
  level: "info" # Log level: debug, info, warn, error
  mode: "json" # Log mode: json, text
```

### Environment Variables

All configuration options can be set via environment variables with the prefix `ARCLIFT_`:

```bash
export ARCLIFT_S3_ENDPOINT="http://localhost:9000"
export ARCLIFT_S3_ACCESS_KEY="admin"
export ARCLIFT_S3_SECRET_KEY="admin123"
export ARCLIFT_S3_BUCKET="my-bucket"
export ARCLIFT_BACKUP_CRON="0 0 * * *"
```

## Usage

### Run Backup Scheduler

Start the backup service with scheduled backups:

```bash
arclift -c /path/to/config.yaml
```

### Manual Backup

Perform a one-time backup:

```bash
arclift backup -c /path/to/config.yaml
```

Or using the subcommand:

```bash
arclift backup add -c /path/to/config.yaml
```

### List Backups

List all available backups:

```bash
arclift backup list -c /path/to/config.yaml
```

### Purge Old Backups

Manually purge old backups based on retention policy:

```bash
arclift backup purge -c /path/to/config.yaml
```

### Configuration Management

Initialize a new configuration file:

```bash
arclift config init -c /path/to/config.yaml
```

## Systemd Service

Arclift includes systemd service integration for running as a system service.

### Service File

Located at `scripts/arclift.service`:

```ini
[Unit]
Description=Arclift Backup Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
ExecStart=/usr/bin/arclift

[Install]
WantedBy=multi-user.target
```

### Installation Scripts

- **Post-Install** (`scripts/postinstall.sh`): Initializes config and enables the service
- **Pre-Remove** (`scripts/preremove.sh`): Stops and disables the service before removal

## Development

### Prerequisites

- Go 1.25.2 or higher
- Docker and Docker Compose (for development environment)
- golangci-lint (installed via `make install-golangci-lint`)
- pre-commit (for git hooks)

### Setup Development Environment

```bash
# Initialize development tools
make init

# Start MinIO for local testing
make dev

# Run tests
make test

# Clean up
make clean
```

### Testing with MinIO

The development environment includes MinIO for local S3-compatible storage testing:

```bash
make dev
```

This starts:

- MinIO server on port 9000 (API) and 9001 (console)
- Creates a test bucket named `test-bucket`
- Sets up test directories and files

Access MinIO console at `http://localhost:9001` with credentials:

- Username: `admin`
- Password: `admin123`

## How It Works

1. **Scheduler Initialization**: On startup, Arclift initializes a cron scheduler based on the configured schedule
2. **Backup Process**:
   - For each configured directory:
     - If `archive-dirs` is enabled: Creates a tar.gz archive
     - If encryption is enabled: Encrypts the archive using GPG
     - Uploads to S3 with a timestamped key
     - Sends success/failure notifications
3. **Retention Management**: After each backup, old backups exceeding the retention count are automatically purged
4. **Version Checking**: Daily checks for new versions and notifies if updates are available

## Backup Key Structure

Backups are stored in S3 with the following key structure:

```txt
<prefix>/<hostname>/<timestamp>
```

- **prefix**: Configured S3 prefix
- **hostname**: Machine hostname or configured identifier
- **timestamp**: Formatted datetime of backup creation
