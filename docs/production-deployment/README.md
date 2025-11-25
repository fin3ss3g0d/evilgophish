# EvilGophish + Frameless-BitB Production Deployment

This directory contains comprehensive documentation and scripts for deploying EvilGophish with Frameless-BitB integration for authorized security testing.

## Directory Structure

```
production-deployment/
├── README.md                                    # This file
├── EVILGOPHISH-FRAMELESS-BITB-SETUP.md         # Complete setup guide
├── scripts/
│   ├── master-setup.sh                         # Automated installation
│   ├── verify-installation.sh                  # Installation verification
│   ├── backup.sh                               # Backup utility
│   ├── restore.sh                              # Restore utility
│   └── monitor.sh                              # Service monitoring
├── configs/
│   ├── nginx/
│   │   └── evilgophish.conf                    # Nginx reverse proxy config
│   ├── apache/
│   │   ├── bitb-vhost.conf                     # Apache VirtualHost config
│   │   └── ssl-hardening.conf                  # SSL security hardening
│   └── systemd/
│       ├── gophish.service                     # Gophish systemd service
│       ├── evilginx3.service                   # Evilginx3 systemd service
│       └── evilfeed.service                    # Evilfeed systemd service
└── templates/
    └── (authorization templates)
```

## Quick Start

### 1. Automated Installation

```bash
# Clone the repository
git clone https://github.com/theroseinc/evilgophish.git
cd evilgophish/docs/production-deployment/scripts

# Make scripts executable
chmod +x *.sh

# Run the master setup script
sudo ./master-setup.sh
```

### 2. Verify Installation

```bash
sudo ./verify-installation.sh
```

### 3. Access Services

- **Gophish Admin**: `https://127.0.0.1:3333` (use SSH tunnel)
- **Evilfeed Dashboard**: `http://127.0.0.1:1337` (use SSH tunnel)

## Documentation

The main setup guide is available in `EVILGOPHISH-FRAMELESS-BITB-SETUP.md` and covers:

1. **Prerequisites & Authorization** - Legal requirements and templates
2. **Environment Setup** - OS and dependency installation
3. **EvilGophish Deployment** - Core installation
4. **Frameless-BitB Integration** - Browser-in-Browser setup
5. **Production Hardening** - Security configurations
6. **Testing & Validation** - Verification procedures
7. **Backup & Recovery** - Data protection
8. **Troubleshooting** - Common issues and solutions

## Requirements

- Ubuntu 20.04 LTS or 22.04 LTS
- 2+ CPU cores, 2+ GB RAM, 20+ GB storage
- Root/sudo access
- Domain with DNS control
- Static public IP address

## Important Notice

This infrastructure is designed for **authorized security testing only**. Always ensure you have proper written authorization before deploying. See the authorization template in the main documentation.

## Scripts Reference

| Script | Purpose | Usage |
|--------|---------|-------|
| `master-setup.sh` | Complete automated installation | `sudo ./master-setup.sh` |
| `verify-installation.sh` | Verify installation status | `sudo ./verify-installation.sh` |
| `backup.sh` | Create backup archive | `sudo ./backup.sh` |
| `restore.sh` | Restore from backup | `sudo ./restore.sh <backup.tar.gz>` |
| `monitor.sh` | Monitor service health | `./monitor.sh` |

## Configuration Files

### Nginx (Alternative to Apache)

Copy `configs/nginx/evilgophish.conf` to `/etc/nginx/sites-available/` and update:
- Replace `YOUR_DOMAIN` with your phishing domain
- Update SSL certificate paths

### Apache

Copy `configs/apache/bitb-vhost.conf` to `/etc/apache2/sites-available/` and update:
- Replace `YOUR_DOMAIN` with your phishing domain
- Replace `YOUR_SERVER_IP` with your server's IP
- Update SSL certificate paths

### Systemd Services

Copy service files from `configs/systemd/` to `/etc/systemd/system/` then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable gophish evilginx3 evilfeed
```

## Support

For issues and questions:
- GitHub Issues: https://github.com/theroseinc/evilgophish/issues

## License

See the main repository LICENSE file.

## Disclaimer

This software is provided for authorized security testing, research, and educational purposes only. Users are responsible for ensuring compliance with all applicable laws and regulations.
