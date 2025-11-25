#!/bin/bash
#
# EvilGophish + Frameless-BitB Installation Verification Script
# Version: 1.0.0
#

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Installation paths
INSTALL_DIR="/opt/evilgophish"
BITB_DIR="/opt/frameless-bitb"

echo ""
echo "=============================================="
echo "  EvilGophish Installation Verification"
echo "=============================================="
echo ""

# Track results
PASSED=0
FAILED=0
WARNINGS=0

check_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

check_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

check_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

# =============================================================================
# Binary Checks
# =============================================================================
echo "=== Binary Checks ==="

if [[ -x "$INSTALL_DIR/evilginx3/evilginx3" ]]; then
    check_pass "Evilginx3 binary exists and is executable"
else
    check_fail "Evilginx3 binary missing or not executable"
fi

if [[ -x "$INSTALL_DIR/gophish/gophish" ]]; then
    check_pass "Gophish binary exists and is executable"
else
    check_fail "Gophish binary missing or not executable"
fi

if [[ -x "$INSTALL_DIR/evilfeed/evilfeed" ]]; then
    check_pass "Evilfeed binary exists and is executable"
else
    check_warn "Evilfeed binary missing (may be disabled)"
fi

# =============================================================================
# Configuration Checks
# =============================================================================
echo ""
echo "=== Configuration Checks ==="

if [[ -f "$INSTALL_DIR/gophish/config.json" ]]; then
    check_pass "Gophish config.json exists"

    # Validate JSON
    if jq empty "$INSTALL_DIR/gophish/config.json" 2>/dev/null; then
        check_pass "Gophish config.json is valid JSON"
    else
        check_fail "Gophish config.json is invalid JSON"
    fi
else
    check_fail "Gophish config.json missing"
fi

if [[ -d "$HOME/.evilginx" ]]; then
    check_pass "Evilginx3 config directory exists"
else
    check_warn "Evilginx3 config directory not yet created (created on first run)"
fi

# =============================================================================
# Database Checks
# =============================================================================
echo ""
echo "=== Database Checks ==="

if [[ -f "$INSTALL_DIR/gophish/gophish.db" ]]; then
    check_pass "Gophish database exists"

    # Check integrity
    if sqlite3 "$INSTALL_DIR/gophish/gophish.db" "PRAGMA integrity_check;" | grep -q "ok"; then
        check_pass "Database integrity check passed"
    else
        check_fail "Database integrity check failed"
    fi

    # Check tables
    TABLE_COUNT=$(sqlite3 "$INSTALL_DIR/gophish/gophish.db" ".tables" | wc -w)
    if [[ $TABLE_COUNT -ge 10 ]]; then
        check_pass "Database has $TABLE_COUNT tables"
    else
        check_warn "Database has only $TABLE_COUNT tables (expected 10+)"
    fi
else
    check_warn "Gophish database not yet created (created on first run)"
fi

# =============================================================================
# Service Checks
# =============================================================================
echo ""
echo "=== Service Checks ==="

for service in apache2 gophish evilginx3 evilfeed; do
    if systemctl is-enabled --quiet $service 2>/dev/null; then
        check_pass "$service service is enabled"
    else
        if [[ "$service" == "evilfeed" ]]; then
            check_warn "$service service not enabled (may be disabled)"
        else
            check_fail "$service service is not enabled"
        fi
    fi

    if systemctl is-active --quiet $service 2>/dev/null; then
        check_pass "$service service is running"
    else
        if [[ "$service" == "evilfeed" ]]; then
            check_warn "$service service not running (may be disabled)"
        else
            check_fail "$service service is not running"
        fi
    fi
done

# =============================================================================
# Network Checks
# =============================================================================
echo ""
echo "=== Network Checks ==="

for port in 53 80 443 3333; do
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        check_pass "Port $port is listening"
    else
        check_fail "Port $port is not listening"
    fi
done

# Port 1337 (Evilfeed - optional)
if netstat -tuln 2>/dev/null | grep -q ":1337 "; then
    check_pass "Port 1337 (Evilfeed) is listening"
else
    check_warn "Port 1337 (Evilfeed) not listening (may be disabled)"
fi

# =============================================================================
# SSL/TLS Checks
# =============================================================================
echo ""
echo "=== SSL/TLS Checks ==="

if [[ -f /etc/apache2/ssl/bitb.crt ]]; then
    check_pass "SSL certificate exists"

    # Check expiration
    EXPIRY=$(openssl x509 -in /etc/apache2/ssl/bitb.crt -noout -enddate 2>/dev/null | cut -d= -f2)
    EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s 2>/dev/null || echo 0)
    NOW_EPOCH=$(date +%s)
    DAYS_LEFT=$(( (EXPIRY_EPOCH - NOW_EPOCH) / 86400 ))

    if [[ $DAYS_LEFT -gt 30 ]]; then
        check_pass "SSL certificate valid for $DAYS_LEFT more days"
    elif [[ $DAYS_LEFT -gt 0 ]]; then
        check_warn "SSL certificate expires in $DAYS_LEFT days"
    else
        check_fail "SSL certificate has expired"
    fi
else
    check_fail "SSL certificate missing"
fi

if [[ -f /etc/apache2/ssl/bitb.key ]]; then
    check_pass "SSL private key exists"

    # Check permissions
    PERMS=$(stat -c %a /etc/apache2/ssl/bitb.key)
    if [[ "$PERMS" == "600" ]]; then
        check_pass "SSL key has correct permissions (600)"
    else
        check_warn "SSL key permissions are $PERMS (should be 600)"
    fi
else
    check_fail "SSL private key missing"
fi

# =============================================================================
# Apache Checks
# =============================================================================
echo ""
echo "=== Apache Checks ==="

if apache2ctl configtest 2>&1 | grep -q "Syntax OK"; then
    check_pass "Apache configuration syntax is valid"
else
    check_fail "Apache configuration has syntax errors"
fi

if [[ -f /etc/apache2/sites-enabled/bitb.conf ]]; then
    check_pass "BitB site is enabled"
else
    check_fail "BitB site is not enabled"
fi

if [[ -d /etc/apache2/custom-subs ]]; then
    check_pass "Custom substitutions directory exists"
else
    check_fail "Custom substitutions directory missing"
fi

# =============================================================================
# Frameless-BitB Checks
# =============================================================================
echo ""
echo "=== Frameless-BitB Checks ==="

if [[ -d /var/www/bitb ]]; then
    check_pass "BitB web root exists"
else
    check_fail "BitB web root missing"
fi

for dir in home primary secondary; do
    if [[ -d /var/www/bitb/$dir ]]; then
        check_pass "BitB $dir directory exists"
    else
        check_fail "BitB $dir directory missing"
    fi
done

# =============================================================================
# Firewall Checks
# =============================================================================
echo ""
echo "=== Firewall Checks ==="

if ufw status | grep -q "Status: active"; then
    check_pass "UFW firewall is active"

    for port in 22 53 80 443; do
        if ufw status | grep -q "$port"; then
            check_pass "Port $port is allowed in firewall"
        else
            check_warn "Port $port may not be allowed in firewall"
        fi
    done
else
    check_warn "UFW firewall is not active"
fi

# =============================================================================
# Log Directory Checks
# =============================================================================
echo ""
echo "=== Log Directory Checks ==="

for logdir in gophish evilginx3 evilfeed; do
    if [[ -d /var/log/$logdir ]]; then
        check_pass "/var/log/$logdir exists"
    else
        check_warn "/var/log/$logdir missing"
    fi
done

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "=============================================="
echo "  Verification Summary"
echo "=============================================="
echo ""
echo -e "  ${GREEN}Passed:${NC}   $PASSED"
echo -e "  ${RED}Failed:${NC}   $FAILED"
echo -e "  ${YELLOW}Warnings:${NC} $WARNINGS"
echo ""

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}All critical checks passed!${NC}"
    exit 0
else
    echo -e "${RED}Some checks failed. Please review and fix issues.${NC}"
    exit 1
fi
