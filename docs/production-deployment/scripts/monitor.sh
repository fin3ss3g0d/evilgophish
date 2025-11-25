#!/bin/bash
#
# EvilGophish Monitoring Script
# Version: 1.0.0
#
# Monitors service health, resource usage, and alerts on issues
#

# Configuration
INSTALL_DIR="/opt/evilgophish"
LOG_FILE="/var/log/evilgophish-monitor.log"
ALERT_EMAIL=""  # Set to enable email alerts

# Thresholds
CPU_THRESHOLD=80
MEMORY_THRESHOLD=80
DISK_THRESHOLD=90

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Logging function
log() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Alert function
alert() {
    local message=$1
    log "ALERT" "$message"

    if [[ -n "$ALERT_EMAIL" ]]; then
        echo "$message" | mail -s "EvilGophish Alert" "$ALERT_EMAIL"
    fi
}

# Check service status
check_service() {
    local service=$1
    if systemctl is-active --quiet "$service"; then
        echo -e "${GREEN}[OK]${NC} $service is running"
        return 0
    else
        echo -e "${RED}[FAIL]${NC} $service is not running"
        alert "Service $service is not running"
        return 1
    fi
}

# Check port status
check_port() {
    local port=$1
    local name=$2
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        echo -e "${GREEN}[OK]${NC} Port $port ($name) is listening"
        return 0
    else
        echo -e "${RED}[FAIL]${NC} Port $port ($name) is not listening"
        alert "Port $port ($name) is not listening"
        return 1
    fi
}

# Check CPU usage
check_cpu() {
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print int($2 + $4)}')
    if [[ $cpu_usage -lt $CPU_THRESHOLD ]]; then
        echo -e "${GREEN}[OK]${NC} CPU usage: ${cpu_usage}%"
    else
        echo -e "${YELLOW}[WARN]${NC} CPU usage: ${cpu_usage}% (threshold: ${CPU_THRESHOLD}%)"
        alert "High CPU usage: ${cpu_usage}%"
    fi
}

# Check memory usage
check_memory() {
    local mem_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [[ $mem_usage -lt $MEMORY_THRESHOLD ]]; then
        echo -e "${GREEN}[OK]${NC} Memory usage: ${mem_usage}%"
    else
        echo -e "${YELLOW}[WARN]${NC} Memory usage: ${mem_usage}% (threshold: ${MEMORY_THRESHOLD}%)"
        alert "High memory usage: ${mem_usage}%"
    fi
}

# Check disk usage
check_disk() {
    local disk_usage=$(df -h / | awk 'NR==2{print int($5)}')
    if [[ $disk_usage -lt $DISK_THRESHOLD ]]; then
        echo -e "${GREEN}[OK]${NC} Disk usage: ${disk_usage}%"
    else
        echo -e "${RED}[FAIL]${NC} Disk usage: ${disk_usage}% (threshold: ${DISK_THRESHOLD}%)"
        alert "High disk usage: ${disk_usage}%"
    fi
}

# Check database size
check_database() {
    if [[ -f "$INSTALL_DIR/gophish/gophish.db" ]]; then
        local db_size=$(du -h "$INSTALL_DIR/gophish/gophish.db" | cut -f1)
        echo -e "${GREEN}[OK]${NC} Database size: $db_size"
    else
        echo -e "${YELLOW}[WARN]${NC} Database not found"
    fi
}

# Check SSL certificate expiration
check_ssl() {
    local cert_file="/etc/apache2/ssl/bitb.crt"
    if [[ -f "$cert_file" ]]; then
        local expiry=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2)
        local expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null || echo 0)
        local now_epoch=$(date +%s)
        local days_left=$(( (expiry_epoch - now_epoch) / 86400 ))

        if [[ $days_left -gt 30 ]]; then
            echo -e "${GREEN}[OK]${NC} SSL certificate valid for $days_left days"
        elif [[ $days_left -gt 7 ]]; then
            echo -e "${YELLOW}[WARN]${NC} SSL certificate expires in $days_left days"
            alert "SSL certificate expires in $days_left days"
        else
            echo -e "${RED}[FAIL]${NC} SSL certificate expires in $days_left days"
            alert "CRITICAL: SSL certificate expires in $days_left days"
        fi
    else
        echo -e "${YELLOW}[WARN]${NC} SSL certificate not found at $cert_file"
    fi
}

# Check recent events
check_events() {
    if [[ -f "$INSTALL_DIR/gophish/gophish.db" ]]; then
        local recent_events=$(sqlite3 "$INSTALL_DIR/gophish/gophish.db" \
            "SELECT COUNT(*) FROM events WHERE time > datetime('now', '-1 hour');" 2>/dev/null || echo 0)
        echo -e "${GREEN}[INFO]${NC} Events in last hour: $recent_events"
    fi
}

# Check log file sizes
check_logs() {
    for logdir in gophish evilginx3 evilfeed; do
        if [[ -d "/var/log/$logdir" ]]; then
            local log_size=$(du -sh "/var/log/$logdir" 2>/dev/null | cut -f1)
            echo -e "${GREEN}[OK]${NC} Log size ($logdir): $log_size"
        fi
    done
}

# Main monitoring function
main() {
    echo ""
    echo "=============================================="
    echo "  EvilGophish System Monitor"
    echo "  $(date '+%Y-%m-%d %H:%M:%S')"
    echo "=============================================="
    echo ""

    echo "=== Service Status ==="
    check_service "apache2"
    check_service "gophish"
    check_service "evilginx3"
    check_service "evilfeed"

    echo ""
    echo "=== Port Status ==="
    check_port 53 "DNS"
    check_port 80 "HTTP"
    check_port 443 "HTTPS"
    check_port 3333 "Gophish Admin"
    check_port 1337 "Evilfeed"

    echo ""
    echo "=== System Resources ==="
    check_cpu
    check_memory
    check_disk

    echo ""
    echo "=== Application Status ==="
    check_database
    check_ssl
    check_events
    check_logs

    echo ""
    echo "=============================================="

    log "INFO" "Monitoring check completed"
}

# Run monitoring
main
