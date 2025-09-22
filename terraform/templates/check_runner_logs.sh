#!/bin/bash

# Enhanced script to check GitHub Actions runner logs
# Usage: ./check_runner_logs.sh <runner-ip-address> [environment]

if [ $# -eq 0 ]; then
    echo "Usage: $0 <runner-ip-address> [environment]"
    echo "Example: $0 123.456.789.012 prod"
    echo ""
    echo "To find your runner IP address:"
    echo "1. Go to OVH Cloud Console"
    echo "2. Navigate to Compute > Instances"
    echo "3. Look for your runner instance and copy its IP address"
    echo ""
    echo "Or use: make status (from terraform directory)"
    echo ""
    echo "Alternative: SSH into the runner and use the built-in utility scripts:"
    echo "  runner-monitor: Check runner status and health"
    echo "  runner-info: Display comprehensive system information"
    exit 1
fi

RUNNER_IP="$1"
ENVIRONMENT="${2:-prod}"
SSH_KEY="./runner_private_key.pem"

echo "🔍 Checking GitHub Actions runner logs on $RUNNER_IP ($ENVIRONMENT environment)"
echo "=================================================================="

# Check if SSH key exists
if [ ! -f "$SSH_KEY" ]; then
    echo "❌ SSH private key not found: $SSH_KEY"
    echo "Please run: make ssh-key (from terraform directory)"
    exit 1
fi

echo "📋 Runner Status Check"
echo "---------------------"
ssh -o ConnectTimeout=10 -o StrictHostKeyChecking=no -i "$SSH_KEY" ubuntu@"$RUNNER_IP" '
echo "=== System Information ==="
echo "Hostname: $(hostname)"
echo "Uptime: $(uptime)"
echo "Date: $(date)"
echo "Environment: '$ENVIRONMENT'"
echo ""

echo "=== GitHub Runner Service Status ==="
sudo systemctl status actions.runner.* --no-pager -l
echo ""

echo "=== GitHub Runner Process Check ==="
ps aux | grep -E "(run.sh|Runner)" | grep -v grep
echo ""

echo "=== Recent GitHub Runner Logs ==="
sudo journalctl -u actions.runner.* --no-pager -n 20
echo ""

echo "=== Health Check Logs ==="
if [ -f "/var/log/runner-health-check.log" ]; then
    echo "Last 10 health check entries:"
    tail -10 /var/log/runner-health-check.log
else
    echo "No health check log found"
fi
echo ""

echo "=== Runner Configuration ==="
if [ -f "/home/runner/.runner" ]; then
    echo "Runner configuration file exists:"
    cat /home/runner/.runner
else
    echo "No runner configuration file found"
fi
echo ""

echo "=== Runner Directory Contents ==="
if [ -d "/home/runner" ]; then
    ls -la /home/runner/
else
    echo "Runner directory not found"
fi
echo ""

echo "=== Docker Status ==="
sudo systemctl status docker --no-pager -l
echo ""

echo "=== Network Connectivity Test ==="
echo "Testing connection to GitHub:"
curl -s -o /dev/null -w "HTTP Status: %{http_code}, Time: %{time_total}s\n" https://api.github.com
echo ""

echo "=== System Resources ==="
echo "CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk "{print \$2}" | cut -d"%" -f1)%"
echo "Memory Usage: $(free | awk "NR==2{printf \"%.1f%%\", \$3*100/\$2}")"
echo "Disk Usage: $(df -h / | awk "NR==2{print \$5}")"
echo ""

echo "=== Runner Status Page ==="
echo "Status page should be available at: http://'$RUNNER_IP'/"
'

echo ""
echo "🔗 Useful Commands:"
echo "------------------"
echo "SSH into runner: ssh -i $SSH_KEY ubuntu@$RUNNER_IP"
echo "Check runner status: ssh -i $SSH_KEY ubuntu@$RUNNER_IP 'sudo systemctl status actions.runner.*'"
echo "View live logs: ssh -i $SSH_KEY ubuntu@$RUNNER_IP 'sudo journalctl -u actions.runner.* -f'"
echo "Restart runner: ssh -i $SSH_KEY ubuntu@$RUNNER_IP 'sudo systemctl restart actions.runner.*'"
echo "Status page: http://$RUNNER_IP/"
echo "Environment: $ENVIRONMENT"
