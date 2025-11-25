#!/bin/bash
#
# EvilGophish Backup Script
# Version: 1.0.0
#
# Creates complete backups of EvilGophish infrastructure
#

set -e

# Configuration
BACKUP_DIR="/opt/backups/evilgophish"
INSTALL_DIR="/opt/evilgophish"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="evilgophish_backup_${DATE}"
RETENTION_DAYS=7

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root"
    exit 1
fi

# Create backup directory
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}"

log_info "Starting backup: ${BACKUP_NAME}"
log_info "Backup location: ${BACKUP_DIR}/${BACKUP_NAME}"

# Stop services for consistent backup
log_info "Stopping services..."
systemctl stop gophish evilginx3 evilfeed 2>/dev/null || true
sleep 2

# Backup Gophish
log_info "Backing up Gophish..."
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}/gophish"
cp "${INSTALL_DIR}/gophish/gophish.db" "${BACKUP_DIR}/${BACKUP_NAME}/gophish/" 2>/dev/null || log_warn "Database not found"
cp "${INSTALL_DIR}/gophish/config.json" "${BACKUP_DIR}/${BACKUP_NAME}/gophish/" 2>/dev/null || log_warn "Config not found"
cp "${INSTALL_DIR}/gophish/"*.crt "${BACKUP_DIR}/${BACKUP_NAME}/gophish/" 2>/dev/null || true
cp "${INSTALL_DIR}/gophish/"*.key "${BACKUP_DIR}/${BACKUP_NAME}/gophish/" 2>/dev/null || true
cp "${INSTALL_DIR}/gophish/"*.log "${BACKUP_DIR}/${BACKUP_NAME}/gophish/" 2>/dev/null || true

# Backup Evilginx3 configuration
log_info "Backing up Evilginx3..."
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}/evilginx3"
if [[ -d "$HOME/.evilginx" ]]; then
    cp -r "$HOME/.evilginx" "${BACKUP_DIR}/${BACKUP_NAME}/evilginx3/config"
fi
cp -r "${INSTALL_DIR}/evilginx3/phishlets" "${BACKUP_DIR}/${BACKUP_NAME}/evilginx3/" 2>/dev/null || true
cp -r "${INSTALL_DIR}/evilginx3/redirectors" "${BACKUP_DIR}/${BACKUP_NAME}/evilginx3/" 2>/dev/null || true

# Backup Apache configuration
log_info "Backing up Apache configuration..."
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}/apache"
cp -r /etc/apache2/sites-available "${BACKUP_DIR}/${BACKUP_NAME}/apache/"
cp -r /etc/apache2/custom-subs "${BACKUP_DIR}/${BACKUP_NAME}/apache/" 2>/dev/null || true
cp -r /etc/apache2/ssl "${BACKUP_DIR}/${BACKUP_NAME}/apache/" 2>/dev/null || true

# Backup SSL certificates
log_info "Backing up SSL certificates..."
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}/ssl"
if [[ -d /etc/letsencrypt ]]; then
    cp -r /etc/letsencrypt "${BACKUP_DIR}/${BACKUP_NAME}/ssl/"
fi

# Backup systemd services
log_info "Backing up systemd services..."
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}/systemd"
cp /etc/systemd/system/gophish.service "${BACKUP_DIR}/${BACKUP_NAME}/systemd/" 2>/dev/null || true
cp /etc/systemd/system/evilginx3.service "${BACKUP_DIR}/${BACKUP_NAME}/systemd/" 2>/dev/null || true
cp /etc/systemd/system/evilfeed.service "${BACKUP_DIR}/${BACKUP_NAME}/systemd/" 2>/dev/null || true

# Backup DNS configuration
log_info "Backing up DNS configuration..."
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}/dns"
cp /etc/hosts "${BACKUP_DIR}/${BACKUP_NAME}/dns/" 2>/dev/null || true
cp /etc/resolv.conf "${BACKUP_DIR}/${BACKUP_NAME}/dns/" 2>/dev/null || true

# Create metadata file
cat > "${BACKUP_DIR}/${BACKUP_NAME}/metadata.json" << EOF
{
    "backup_date": "$(date -Iseconds)",
    "hostname": "$(hostname)",
    "os_version": "$(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)",
    "evilgophish_version": "$(cd $INSTALL_DIR && git describe --tags 2>/dev/null || echo 'unknown')"
}
EOF

# Create archive
log_info "Creating archive..."
cd "${BACKUP_DIR}"
tar -czf "${BACKUP_NAME}.tar.gz" "${BACKUP_NAME}"
rm -rf "${BACKUP_NAME}"

# Calculate checksum
CHECKSUM=$(sha256sum "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" | cut -d' ' -f1)
echo "$CHECKSUM" > "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz.sha256"

# Restart services
log_info "Restarting services..."
systemctl start gophish evilginx3 evilfeed 2>/dev/null || true

# Cleanup old backups
log_info "Cleaning up old backups..."
find "${BACKUP_DIR}" -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete
find "${BACKUP_DIR}" -name "*.sha256" -mtime +${RETENTION_DAYS} -delete

# Calculate backup size
BACKUP_SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" | cut -f1)

log_info "Backup complete!"
echo ""
echo "=============================================="
echo "  Backup Summary"
echo "=============================================="
echo "  File: ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
echo "  Size: ${BACKUP_SIZE}"
echo "  Checksum: ${CHECKSUM:0:16}..."
echo "=============================================="
