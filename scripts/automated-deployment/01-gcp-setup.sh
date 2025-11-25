#!/bin/bash
#
# EvilGophish GCP Setup - Automated Script
# This script automates the GCP project, DNS, and VM deployment
#

set -e  # Exit on any error

echo "=========================================="
echo "  EvilGophish Automated GCP Setup"
echo "=========================================="
echo ""

# Configuration
DOMAIN="exodustraderai.com"
REGION="us-central1"
ZONE="us-central1-c"
VM_NAME="evilgophish-server"
MACHINE_TYPE="e2-standard-2"
IP_NAME="evilgophish-ip"
DNS_ZONE_NAME="exodustraderai-zone"
FIREWALL_RULE="evilgophish-allow-all"

# Subdomains to create
SUBDOMAINS=("login" "www" "portal" "admin" "feed" "api")

echo "Configuration:"
echo "  Domain: $DOMAIN"
echo "  Region: $REGION"
echo "  Zone: $ZONE"
echo "  Subdomains: ${SUBDOMAINS[*]}"
echo ""

# Step 1: Create or select project
echo "=========================================="
echo "Step 1: Project Setup"
echo "=========================================="
echo ""

read -p "Do you want to create a NEW project? (y/n): " CREATE_NEW

if [[ "$CREATE_NEW" == "y" ]]; then
    PROJECT_ID="evilgophish-prod-$(date +%s)"
    echo "Creating new project: $PROJECT_ID"
    gcloud projects create $PROJECT_ID --name="EvilGophish Production"
    gcloud config set project $PROJECT_ID
    echo "✓ Project created: $PROJECT_ID"
else
    echo "Available projects:"
    gcloud projects list
    echo ""
    read -p "Enter your project ID: " PROJECT_ID
    gcloud config set project $PROJECT_ID
    echo "✓ Using project: $PROJECT_ID"
fi

echo ""
echo "IMPORTANT: Link billing account to project"
echo "Available billing accounts:"
gcloud billing accounts list
echo ""
read -p "Enter billing account ID: " BILLING_ACCOUNT
gcloud billing projects link $PROJECT_ID --billing-account=$BILLING_ACCOUNT
echo "✓ Billing linked"

# Step 2: Enable APIs
echo ""
echo "=========================================="
echo "Step 2: Enabling Required APIs"
echo "=========================================="
echo ""

gcloud services enable compute.googleapis.com
gcloud services enable dns.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
echo "✓ APIs enabled"
echo "Waiting 30 seconds for APIs to initialize..."
sleep 30

# Step 3: Create DNS Zone
echo ""
echo "=========================================="
echo "Step 3: Creating Google Cloud DNS Zone"
echo "=========================================="
echo ""

# Check if zone already exists
if gcloud dns managed-zones describe $DNS_ZONE_NAME &>/dev/null; then
    echo "⚠ DNS zone already exists"
else
    gcloud dns managed-zones create $DNS_ZONE_NAME \
        --dns-name="$DOMAIN." \
        --description="DNS zone for EvilGophish" \
        --visibility=public
    echo "✓ DNS zone created"
fi

# Get nameservers
echo ""
echo "=========================================="
echo "IMPORTANT: UPDATE YOUR DOMAIN NAMESERVERS"
echo "=========================================="
echo ""
echo "Go to your domain registrar and update nameservers to:"
echo ""
gcloud dns managed-zones describe $DNS_ZONE_NAME --format="value(nameServers)" | tr ';' '\n'
echo ""
read -p "Press ENTER after you've updated your nameservers..."

# Step 4: Reserve External IP
echo ""
echo "=========================================="
echo "Step 4: Reserving Static External IP"
echo "=========================================="
echo ""

# Check if IP already exists
if gcloud compute addresses describe $IP_NAME --region=$REGION &>/dev/null; then
    echo "⚠ IP address already exists"
    EXTERNAL_IP=$(gcloud compute addresses describe $IP_NAME --region=$REGION --format="get(address)")
else
    gcloud compute addresses create $IP_NAME \
        --region=$REGION \
        --network-tier=PREMIUM
    EXTERNAL_IP=$(gcloud compute addresses describe $IP_NAME --region=$REGION --format="get(address)")
    echo "✓ External IP reserved"
fi

echo ""
echo "Your External IP: $EXTERNAL_IP"
echo ""

# Step 5: Create DNS A Records
echo ""
echo "=========================================="
echo "Step 5: Creating DNS A Records"
echo "=========================================="
echo ""

# Start transaction
gcloud dns record-sets transaction start --zone=$DNS_ZONE_NAME 2>/dev/null || {
    gcloud dns record-sets transaction abort --zone=$DNS_ZONE_NAME 2>/dev/null
    gcloud dns record-sets transaction start --zone=$DNS_ZONE_NAME
}

# Add A records for each subdomain
for SUBDOMAIN in "${SUBDOMAINS[@]}"; do
    FQDN="${SUBDOMAIN}.${DOMAIN}."

    # Check if record already exists
    if gcloud dns record-sets list --zone=$DNS_ZONE_NAME --name=$FQDN --type=A 2>/dev/null | grep -q "$FQDN"; then
        echo "⚠ A record for $FQDN already exists, skipping"
    else
        gcloud dns record-sets transaction add $EXTERNAL_IP \
            --name=$FQDN \
            --ttl=300 \
            --type=A \
            --zone=$DNS_ZONE_NAME
        echo "✓ Added A record: $FQDN -> $EXTERNAL_IP"
    fi
done

# Execute transaction
gcloud dns record-sets transaction execute --zone=$DNS_ZONE_NAME 2>/dev/null || {
    echo "⚠ Transaction already executed or no changes needed"
    gcloud dns record-sets transaction abort --zone=$DNS_ZONE_NAME 2>/dev/null || true
}

echo "✓ DNS records created"

# Verify DNS records
echo ""
echo "DNS Records:"
gcloud dns record-sets list --zone=$DNS_ZONE_NAME

# Step 6: Create Firewall Rule
echo ""
echo "=========================================="
echo "Step 6: Creating Firewall Rules"
echo "=========================================="
echo ""

if gcloud compute firewall-rules describe $FIREWALL_RULE &>/dev/null; then
    echo "⚠ Firewall rule already exists"
else
    gcloud compute firewall-rules create $FIREWALL_RULE \
        --direction=INGRESS \
        --priority=1000 \
        --network=default \
        --action=ALLOW \
        --rules=tcp:22,tcp:53,udp:53,tcp:80,tcp:443,tcp:3333 \
        --source-ranges=0.0.0.0/0 \
        --target-tags=evilgophish-server
    echo "✓ Firewall rule created"
fi

# Step 7: Create VM Instance
echo ""
echo "=========================================="
echo "Step 7: Creating VM Instance"
echo "=========================================="
echo ""

if gcloud compute instances describe $VM_NAME --zone=$ZONE &>/dev/null; then
    echo "⚠ VM already exists"
    read -p "Do you want to DELETE and RECREATE the VM? (y/n): " RECREATE_VM
    if [[ "$RECREATE_VM" == "y" ]]; then
        echo "Deleting existing VM..."
        gcloud compute instances delete $VM_NAME --zone=$ZONE --quiet
        sleep 10
    else
        echo "Using existing VM"
        VM_EXTERNAL_IP=$(gcloud compute instances describe $VM_NAME --zone=$ZONE --format="get(networkInterfaces[0].accessConfigs[0].natIP)")
        echo "VM External IP: $VM_EXTERNAL_IP"
        echo ""
        echo "=========================================="
        echo "  GCP Setup Complete!"
        echo "=========================================="
        echo ""
        echo "VM Name: $VM_NAME"
        echo "External IP: $EXTERNAL_IP"
        echo "SSH Command: gcloud compute ssh $VM_NAME --zone=$ZONE"
        echo ""
        echo "Next: Run the VM installation script"
        echo "  ./02-vm-install.sh"
        exit 0
    fi
fi

gcloud compute instances create $VM_NAME \
    --zone=$ZONE \
    --machine-type=$MACHINE_TYPE \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --boot-disk-size=50GB \
    --boot-disk-type=pd-standard \
    --tags=evilgophish-server \
    --address=$IP_NAME \
    --metadata=startup-script='#!/bin/bash
apt-get update
apt-get install -y git curl wget build-essential'

echo "✓ VM created: $VM_NAME"
echo ""
echo "Waiting 60 seconds for VM to fully boot..."
sleep 60

# Step 8: Save configuration
echo ""
echo "=========================================="
echo "Step 8: Saving Configuration"
echo "=========================================="
echo ""

cat > /tmp/evilgophish-config.sh << EOF
# EvilGophish Configuration
export PROJECT_ID="$PROJECT_ID"
export DOMAIN="$DOMAIN"
export EXTERNAL_IP="$EXTERNAL_IP"
export REGION="$REGION"
export ZONE="$ZONE"
export VM_NAME="$VM_NAME"
EOF

echo "✓ Configuration saved to /tmp/evilgophish-config.sh"

# Final output
echo ""
echo "=========================================="
echo "  GCP Setup Complete!"
echo "=========================================="
echo ""
echo "Project ID: $PROJECT_ID"
echo "Domain: $DOMAIN"
echo "External IP: $EXTERNAL_IP"
echo "VM Name: $VM_NAME"
echo "Zone: $ZONE"
echo ""
echo "DNS Nameservers (update at your registrar):"
gcloud dns managed-zones describe $DNS_ZONE_NAME --format="value(nameServers)" | tr ';' '\n'
echo ""
echo "Next Steps:"
echo "  1. Verify DNS propagation: dig login.$DOMAIN +short"
echo "  2. SSH to VM: gcloud compute ssh $VM_NAME --zone=$ZONE"
echo "  3. Run installation script on VM"
echo ""
echo "SSH Command to copy:"
echo "  gcloud compute ssh $VM_NAME --zone=$ZONE"
echo ""
