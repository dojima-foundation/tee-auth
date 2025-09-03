# 🚀 Manual Setup Guide - GitHub Actions Runners on OVH Cloud

## 📋 Prerequisites ✅ COMPLETE
- ✅ OVH Cloud account configured
- ✅ Terraform infrastructure deployed
- ✅ SSH keys generated
- ✅ OpenStack user created

## 🎯 Step-by-Step Instructions

### Step 1: Get Your SSH Public Key
```bash
# Run this command to get your SSH public key
terraform output ssh_public_key
```

**Copy the entire output** - you'll need this for Step 4.

### Step 2: Access OVH Cloud Console
1. **Go to**: https://www.ovh.com/manager/public-cloud/
2. **Login** with your OVH account credentials
3. **Select your project**: `5729bde2da4e41c8b8157376c56c6899` (Prod-dojima-cluster)

### Step 3: Navigate to Instances
1. **Click**: "Compute" in the left sidebar
2. **Click**: "Instances"
3. **Click**: "Create an instance"

### Step 4: Configure Instance 1 (github-runner-1)

#### Basic Configuration:
- **Name**: `github-runner-1`
- **Region**: `SGP1` (Singapore)
- **Image**: `Ubuntu 22.04`
- **Flavor**: `b2-7` (2 vCPUs, 7GB RAM)

#### SSH Key:
1. **Click**: "Add a new SSH key"
2. **Name**: `github-runner-key`
3. **Public key**: Paste the SSH public key from Step 1
4. **Click**: "Add this SSH key"

#### Advanced Configuration:
1. **Click**: "Advanced" section
2. **User data**: Copy and paste the entire content from `templates/user_data.sh`
3. **Click**: "Create the instance"

### Step 5: Configure Instance 2 (github-runner-2)

**Repeat Step 4** with these changes:
- **Name**: `github-runner-2`
- **User data**: Same script but change `RUNNER_NAME="runner-2"` in the script

### Step 6: Wait for Instances to Start
- **Status**: Wait until both instances show "ACTIVE"
- **Time**: Usually 2-5 minutes per instance

### Step 7: Verify in GitHub
1. **Go to**: https://github.com/luffybhaagi/tee-auth/settings/actions/runners
2. **Check**: Look for `runner-1` and `runner-2`
3. **Status**: Should show "Online" after a few minutes

### Step 8: Test Your Runners
1. **Create a test workflow** in your repository
2. **Check**: Runners should appear in the job queue
3. **Monitor**: Jobs should execute successfully

## 🔧 Troubleshooting

### If Runners Don't Appear:
1. **Check instance logs** in OVH Cloud Console
2. **SSH into instance**: `ssh -i runner_private_key.pem ubuntu@<instance-ip>`
3. **Check service**: `sudo systemctl status github-runner`
4. **Check logs**: `sudo journalctl -u github-runner`

### If SSH Connection Fails:
1. **Verify SSH key**: Check the public key was added correctly
2. **Check instance status**: Ensure instance is ACTIVE
3. **Try different SSH options**: `ssh -o StrictHostKeyChecking=no -i runner_private_key.pem ubuntu@<instance-ip>`

### If GitHub Token Issues:
1. **Verify token permissions**: Token needs `repo` and `admin:org` scopes
2. **Check token validity**: Ensure token hasn't expired
3. **Regenerate token** if needed

## 📊 Monitoring

### Status Pages:
- **Runner 1**: `http://<instance-1-ip>/runner-status.html`
- **Runner 2**: `http://<instance-2-ip>/runner-status.html`

### SSH Access:
```bash
# Get instance IPs from OVH Cloud Console
ssh -i runner_private_key.pem ubuntu@<instance-ip>

# Check runner status
sudo systemctl status github-runner

# View logs
sudo journalctl -u github-runner -f
```

### Health Checks:
- **Automatic**: Every 5 minutes via cron
- **Manual**: `sudo systemctl restart github-runner` if needed

## 💰 Cost Management

### Current Setup:
- **2 instances**: b2-7 flavor
- **Estimated cost**: ~$20-30/month total
- **Region**: SGP1 (Singapore)

### Scaling Options:
- **Scale up**: Change to larger flavors (b2-15, b2-30)
- **Scale out**: Add more instances
- **Scale down**: Use smaller flavors for cost savings

## 🛡️ Security Notes

### What's Configured:
- ✅ **Firewall**: UFW enabled with SSH and HTTP access
- ✅ **SSH Keys**: Secure key-based authentication
- ✅ **User isolation**: Runner runs as dedicated user
- ✅ **Service isolation**: Runner service with proper permissions

### Best Practices:
- **Keep SSH key secure**: Don't share `runner_private_key.pem`
- **Monitor access**: Check SSH logs regularly
- **Update regularly**: Keep Ubuntu and Docker updated
- **Backup configuration**: Save your Terraform state

## 🎉 Success Indicators

You'll know everything is working when:

1. ✅ **Instances**: Both show "ACTIVE" in OVH Cloud Console
2. ✅ **GitHub**: Runners appear in repository settings
3. ✅ **Status**: Runners show "Online" status
4. ✅ **Jobs**: Workflows execute successfully
5. ✅ **Monitoring**: Status pages are accessible

## 📞 Need Help?

### OVH Cloud Issues:
- **Documentation**: https://docs.ovh.com/gb/en/public-cloud/
- **Support**: Contact OVH Cloud support

### GitHub Actions Issues:
- **Documentation**: https://docs.github.com/en/actions/hosting-your-own-runners
- **Community**: GitHub Community forums

### Local Issues:
- **Check logs**: Instance logs in OVH Console
- **SSH debug**: Connect to instances for troubleshooting
- **Terraform**: Check state with `terraform show`

---

## 🚀 **Ready to Start?**

Your infrastructure is complete. Follow the steps above to create your GitHub Actions runners!

**Estimated time**: 15-30 minutes
**Difficulty**: Easy (step-by-step instructions)
**Result**: 2 fully functional GitHub Actions runners
