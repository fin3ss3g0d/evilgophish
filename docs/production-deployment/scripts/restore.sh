#!/bin/bash
#
# EvilGophish Restore Script
# Version: 1.0.0
#
# Restores EvilGophish from a backup archive
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Installation paths
INSTALL_DIR="/opt/evilgophish"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check arguments
if [[ -z "$1" ]]; then
    echo "Usage: $0 <backup_file.tar.gz>"
    echo ""
    echo "Example: $0 /opt/backups/evilgophish/evilgophish_backup_20240115_120000.tar.gz"
    exit 1
fi

BACKUP_FILE="$1"
RESTORE_DIR="/tmp/evilgophish_restore_$$"

# Check root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root"
    exit 1
fi

# Check backup file exists
if [[ ! -f "$BACKUP_FILE" ]]; then
    log_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Verify checksum if available
CHECKSUM_FILE="${BACKUP_FILE}.sha256"
if [[ -f "$CHECKSUM_FILE" ]]; then
    log_info "Verifying backup integrity..."
    if sha256sum -c "$CHECKSUM_FILE" >/dev/null 2>&1; then
        log_info "Checksum verification passed"
    else
        log_error "Checksum verification failed"
        exit 1
    fi
fi

echo ""
echo "=============================================="
echo "  EvilGophish Restore"
echo "=============================================="
echo ""
echo "Backup file: $BACKUP_FILE"
echo ""

read -p "This will overwrite current configuration. Continue? (y/n): " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    log_info "Restore cancelled"
    exit 0
fi

log_info "Starting restoration..."

# Stop services
log_info "Stopping services..."
systemctl stop gophish evilginx3 evilfeed apache2 2>/dev/null || true
sleep 2

# Extract backup
log_info "Extracting backup archive..."
mkdir -p "$RESTORE_DIR"
tar -xzf "$BACKUP_FILE" -C "$RESTORE_DIR"
BACKUP_NAME=$(ls "$RESTORE_DIR")

# Display backup metadata
if [[ -f "$RESTORE_DIR/$BACKUP_NAME/metadata.json" ]]; then
    log_info "Backup metadata:"
    jq . "$RESTORE_DIR/$BACKUP_NAME/metadata.json"
    echo ""
fi

# Restore Gophish
if [[ -d "$RESTORE_DIR/$BACKUP_NAME/gophish" ]]; then
    log_info "Restoring Gophish..."

    # Backup current database
    if [[ -f "$INSTALL_DIR/gophish/gophish.db" ]]; then
        cp "$INSTALL_DIR/gophish/gophish.db" "$INSTALL_DIR/gophish/gophish.db.pre-restore"
    fi

    cp "$RESTORE_DIR/$BACKUP_NAME/gophish/gophish.db" "$INSTALL_DIR/gophish/" 2>/dev/null || true
    cp "$RESTORE_DIR/$BACKUP_NAME/gophish/config.json" "$INSTALL_DIR/gophish/" 2>/dev/null || true
    cp "$RESTORE_DIR/$BACKUP_NAME/gophish/"*.crt "$INSTALL_DIR/gophish/" 2>/dev/null || true
    cp "$RESTORE_DIR/$BACKUP_NAME/gophish/"*.key "$INSTALL_DIR/gophish/" 2>/dev/null || true

    chown -R gophish:gophish "$INSTALL_DIR/gophish"
fi

# Restore Evilginx3 configuration
if [[ -d "$RESTORE_DIR/$BACKUP_NAME/evilginx3" ]]; then
    log_info "Restoring Evilginx3 configuration..."

    if [[ -d "$RESTORE_DIR/$BACKUP_NAME/evilginx3/config" ]]; then
        rm -rf "$HOME/.evilginx"
        cp -r "$RESTORE_DIR/$BACKUP_NAME/evilginx3/config" "$HOME/.evilginx"
    fi

    if [[ -d "$RESTORE_DIR/$BACKUP_NAME/evilginx3/phishlets" ]]; then
        cp -r "$RESTORE_DIR/$BACKUP_NAME/evilginx3/phishlets/"* "$INSTALL_DIR/evilginx3/phishlets/" 2>/dev/null || true
    fi
fi

# Restore Apache configuration
if [[ -d "$RESTORE_DIR/$BACKUP_NAME/apache" ]]; then
    log_info "Restoring Apache configuration..."

    if [[ -d "$RESTORE_DIR/$BACKUP_NAME/apache/sites-available" ]]; then
        cp -r "$RESTORE_DIR/$BACKUP_NAME/apache/sites-available/"* /etc/apache2/sites-available/ 2>/dev/null || true
    fi

    if [[ -d "$RESTORE_DIR/$BACKUP_NAME/apache/custom-subs" ]]; then
        mkdir -p /etc/apache2/custom-subs
        cp -r "$RESTORE_DIR/$BACKUP_NAME/apache/custom-subs/"* /etc/apache2/custom-subs/ 2>/dev/null || true
    fi

    if [[ -d "$RESTORE_DIR/$BACKUP_NAME/apache/ssl" ]]; then
        mkdir -p /etc/apache2/ssl
        cp -r "$RESTORE_DIR/$BACKUP_NAME/apache/ssl/"* /etc/apache2/ssl/ 2>/dev/null || true
        chmod 600 /etc/apache2/ssl/*.key 2>/dev/null || true
    fi
fi

# Restore SSL certificates
if [[ -d "$RESTORE_DIR/$BACKUP_NAME/ssl/letsencrypt" ]]; then
    log_info "Restoring Let's Encrypt certificates..."
    rm -rf /etc/letsencrypt
    cp -r "$RESTORE_DIR/$BACKUP_NAME/ssl/letsencrypt" /etc/
fi

# Restore systemd services
if [[ -d "$RESTORE_DIR/$BACKUP_NAME/systemd" ]]; then
    log_info "Restoring systemd services..."
    cp "$RESTORE_DIR/$BACKUP_NAME/systemd/"*.service /etc/systemd/system/ 2>/dev/null || true
    systemctl daemon-reload
fi

# Restore DNS configuration
if [[ -d "$RESTORE_DIR/$BACKUP_NAME/dns" ]]; then
    log_info "Restoring DNS configuration..."

    # Only restore if files exist in backup
    if [[ -f "$RESTORE_DIR/$BACKUP_NAME/dns/hosts" ]]; then
        cp "$RESTORE_DIR/$BACKUP_NAME/dns/hosts" /etc/hosts
    fi
fi

# Cleanup
log_info "Cleaning up..."
rm -rf "$RESTORE_DIR"

# Start services
log_info "Starting services..."
systemctl start apache2
sleep 2
systemctl start gophish
sleep 3
systemctl start evilginx3
systemctl start evilfeed 2>/dev/null || true

# Verify services
log_info "Verifying services..."
for service in apache2 gophish evilginx3; do
    if systemctl is-active --quiet $service; then
        log_info "$service is running"
    else
        log_warn "$service failed to start"
    fi
done

echo ""
echo "=============================================="
echo "  Restore Complete"
echo "=============================================="
echo ""
log_info "Please verify your configuration and test the setup"
log_info "Pre-restore database backup: $INSTALL_DIR/gophish/gophish.db.pre-restore"
