# Contact Form Example with n8n-cli

A complete example demonstrating how to build, deploy, and manage a contact form workflow using n8n-cli with automated GitHub Actions deployment.

## Table of Contents

- [🎯 What You'll Learn](#-what-youll-learn)
- [📦 What's Included](#-whats-included)
- [📁 Project Structure](#-project-structure)
- [🚀 Quick Start](#-quick-start)
  - [Prerequisites](#prerequisites)
  - [Step 1: Clone and Setup](#step-1-clone-and-setup)
  - [Step 2: Choose Your n8n Setup](#step-2-choose-your-n8n-setup)
  - [Step 3: Configure Environment](#step-3-configure-environment)
  - [Step 4: Deploy the Workflow](#step-4-deploy-the-workflow)
  - [Step 5: Configure the Contact Form](#step-5-configure-the-contact-form)
- [✅ Testing Your Setup](#-testing-your-setup)
  - [Local Testing with Mailhog](#local-testing-with-mailhog)
  - [Production Testing](#production-testing)
  - [Debugging Tips](#debugging-tips)
- [🎨 Customization Guide](#-customization-guide)
  - [Workflow Modifications](#workflow-modifications)
  - [Common Customizations](#common-customizations)
  - [Deployment After Changes](#deployment-after-changes)
- [🛠️ Development Commands](#️-development-commands)
  - [Quick Reference](#quick-reference)
  - [Docker Environment](#docker-environment)
  - [Workflow Management](#workflow-management)
  - [Development Tools](#development-tools)
- [📧 Email Configuration](#-email-configuration)
  - [Local Development (Mailhog)](#local-development-mailhog)
  - [Production Email Setup](#production-email-setup)
  - [Setting Credentials in n8n](#setting-credentials-in-n8n)
- [🔒 Security Best Practices](#-security-best-practices)
  - [Webhook Security](#webhook-security)
  - [CORS Configuration](#cors-configuration)
  - [API Key Management](#api-key-management)
  - [Input Validation](#input-validation)
- [🐛 Troubleshooting](#-troubleshooting)
  - [Common Issues and Solutions](#common-issues-and-solutions)
  - [Debug Commands](#debug-commands)
  - [Getting Help](#getting-help)
- [📚 Resources](#-resources)
  - [Documentation](#documentation)
  - [Tutorials](#tutorials)
  - [Community](#community)
- [📄 License](#-license)

## 🎯 What You'll Learn

- Setting up a webhook-triggered n8n workflow for form processing
- Configuring email notifications with form submission data
- Using n8n-cli for local development and workflow synchronization
- Implementing CI/CD with GitHub Actions for automatic deployments
- Testing email workflows locally with Mailhog

## 📦 What's Included

| Component | Description |
|-----------|-------------|
| **Contact Form Workflow** | Pre-built n8n workflow that receives webhook data and sends formatted emails |
| **HTML Contact Form** | Ready-to-use responsive contact form with client-side validation |
| **Docker Compose Setup** | Local development environment with n8n, Mailhog, and Nginx |
| **GitHub Actions** | Automated CI/CD pipeline for workflow deployment |
| **Taskfile** | Convenient commands for common development tasks |

## 📁 Project Structure

```
├── .github/
│   └── workflows/
│       └── sync-n8n.yml     # CI/CD pipeline for automatic n8n deployment
├── workflows/
│   └── Contact_Form.yaml    # n8n workflow definition (webhook → email)
├── docker-compose.yaml      # Local development environment setup
├── .env.example             # Environment variable template
├── .gitignore               # Version control exclusions
├── contact-form.html        # Frontend contact form with validation
├── Taskfile.yaml            # Development automation commands
└── README.md                # Project documentation
```

## 🚀 Quick Start

### Prerequisites

- [Docker](https://www.docker.com/get-started) and Docker Compose
- [n8n-cli](https://github.com/edenreich/n8n-cli#installation) installed
- [Task](https://taskfile.dev/installation/) (optional, for using Taskfile commands)
- Git for version control

### Step 1: Clone and Setup

```bash
# Clone this example
git clone <your-repo-url>
cd contact-form

# Create environment configuration
cp .env.example .env
# Edit .env with your credentials
```

### Step 2: Choose Your n8n Setup

#### Option A: Local Development with Docker (Recommended)

Perfect for local development and testing. Includes:

| Service | Purpose | Access URL |
|---------|---------|------------|
| n8n | Workflow automation platform | http://localhost:5678 |
| Mailhog | Email testing (catches all emails) | http://localhost:8025 |
| Nginx | Contact form web server | http://localhost:8080 |

```bash
# Start the development environment
task up
# Or without Task: docker compose up -d --build

# Services will be available at:
# - n8n: http://localhost:5678
# - Mailhog: http://localhost:8025
# - Contact form: http://localhost:8080

# View service logs
task logs        # n8n logs
task logs-mail   # Mailhog logs

# Stop everything
task down
```

**CLI Options for Docker:**

You can use the n8n-cli in Docker with two different approaches:

1. **Standard CLI (Downloaded)**: Use the `cli` service which downloads and installs the latest release:
   ```bash
   docker compose run --rm cli
   n8n workflows list
   ```

2. **Local CLI (Built from Source)**: Use the `cli-dev` service which builds the CLI from the source code in the parent repository:
   ```bash
   docker compose run --rm cli-dev
   n8n workflows list
   ```
   This is useful when developing or testing changes to the CLI itself.

**First-time setup:**
1. Access n8n at http://localhost:5678
2. Create your admin account
3. Navigate to Settings → API → Create API Key
4. Save the API key in your `.env` file

#### Option B: Use Existing n8n Instance

Connect to your existing n8n instance:

| Platform | Setup Guide |
|----------|-------------|
| n8n Cloud | [Sign up for free](https://www.n8n.cloud/) → Get API key from Settings |
| Self-hosted | Use your instance URL and [generate API key](https://docs.n8n.io/api/authentication/) |
| Local npm | Run `npx n8n` and access at http://localhost:5678 |

### Step 3: Configure Environment

#### Local Development

Edit your `.env` file:

```bash
N8N_API_KEY=your_api_key_here
N8N_INSTANCE_URL=http://localhost:5678  # or your cloud instance
```

#### GitHub Actions (For Automated Deployment)

1. Go to: **Repository Settings → Secrets and variables → Actions**
2. Add these repository secrets:

| Secret Name | Value Example |
|-------------|---------------|
| `N8N_API_KEY` | `n8n_api_abc123...` |
| `N8N_INSTANCE_URL` | `https://your-instance.n8n.cloud` |

### Step 4: Deploy the Workflow

#### Method 1: Manual Deployment (Development)

```bash
# Using Taskfile (recommended)
task sync

# Or using n8n-cli directly
n8n workflows sync --directory workflows/

# Preview changes without applying
task sync-dry-run
```

#### Method 2: Automated Deployment (Production)

Simply push to your repository:

```bash
git checkout -b chore/update-contact-form
# Make changes to workflows/Contact_Form.yaml if needed
git add .
git commit -m "chore: Update contact form workflow"
git push
gh pr create --title "chore: Update Contact Form Workflow"
# Someone reviews the PR and merges it
# After merging, GitHub Actions triggers deployment on workflow dispatch
```

GitHub Actions will automatically:
1. Validate the workflow syntax
2. Deploy to your n8n instance
3. Activate the workflow, if activate is enabled

### Step 5: Configure the Contact Form

#### Get Your Webhook URL

1. Open n8n dashboard → **Workflows** → **Contact Form**
2. Click the **Webhook** node
3. Copy the production URL (looks like: `https://your-instance.n8n.cloud/webhook/abc-123`)

#### Update and Deploy the Form

```bash
# Update the webhook URL in the HTML form
sed -i 's|YOUR_N8N_WEBHOOK_URL_HERE|your-actual-webhook-url|g' contact-form.html

# Test locally - form is available via nginx
# Opens form at http://localhost:8080
```

#### Form Hosting Options

- **Static hosting**: GitHub Pages, Netlify, Vercel
- **WordPress**: Embed as custom HTML block
- **Existing website**: Copy HTML into your site

## ✅ Testing Your Setup

### Local Testing with Mailhog

1. Submit a test form at http://localhost:8000
2. Check Mailhog at http://localhost:8025
3. Verify email formatting and data

### Production Testing

1. Submit form from your deployed URL
2. Check workflow execution in n8n dashboard
3. Verify email delivery to configured address

### Debugging Tips

```bash
# Check workflow logs
task logs

# Test webhook directly
curl -X POST your-webhook-url \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","message":"Hello"}'

# Verify workflow is active
n8n workflows list --active
```

## 🎨 Customization Guide

### Workflow Modifications

| What to Change | Where to Edit | Example |
|----------------|---------------|----------|
| Email template | `workflows/Contact_Form.yaml` → Email node | Add custom HTML template |
| Form fields | `contact-form.html` + Webhook node | Add phone number field |
| Notifications | Add nodes in n8n | Send to Slack, Discord, etc. |
| Data storage | Add Database node | Store in PostgreSQL, MongoDB |

### Common Customizations

```yaml
# Add spam protection (in Contact_Form.yaml)
- Add IF node to check for spam keywords
- Integrate with reCAPTCHA

# Multiple recipients
- Duplicate Email node
- Set different recipients

# Auto-response
- Add second Email node
- Send confirmation to submitter
```

### Deployment After Changes

```bash
# Local testing
task sync-dry-run  # Preview changes
task sync          # Apply changes

# Production
git commit -am "Update workflow"
git push  # Auto-deploys via GitHub Actions
```

## 🛠️ Development Commands

### Quick Reference

```bash
task help        # Show all available commands
```

### Docker Environment

| Command | Description | Use When |
|---------|-------------|----------|
| `task up` | Start all services | Beginning development |
| `task down` | Stop services | Done for the day |
| `task restart` | Restart services | Services acting up |
| `task clean` | Remove everything | Fresh start needed |
| `task logs` | View n8n logs | Debugging workflows |
| `task logs-mail` | View email logs | Testing email delivery |

### Workflow Management

| Command | Description | Use When |
|---------|-------------|----------|
| `task sync` | Deploy workflows | Ready to test changes |
| `task sync-dry-run` | Preview deployment | Before actual sync |
| `task refresh` | Pull from n8n | Get latest version |
| `task list` | List all workflows | Check deployment status |
| `task executions` | View execution history | Debug past workflow runs |

### Development Tools

| Command | Description |
|---------|-------------|
| `task setup-env` | Create .env template |

## 📧 Email Configuration

### Local Development (Mailhog)

Perfect for testing! Mailhog catches all emails locally:
- SMTP Server: `mailhog`
- Port: `1025`
- No authentication required
- View emails at: http://localhost:8025

**Setting up in n8n:**
1. Go to **Credentials** → **New** → **Email (SMTP)**
2. Name: `Mailhog Local`
3. Host: `mailhog`, Port: `1025`
4. Leave user/password empty, disable SSL/TLS
5. Save and update workflow to use this credential

**Important:** For production, replace with a real email provider (see below)

### Production Email Setup

#### Gmail
```yaml
SMTP Host: smtp.gmail.com
Port: 587
User: your-email@gmail.com
Password: App-specific password (not regular password)
SSL/TLS: STARTTLS
```

#### SendGrid
```yaml
SMTP Host: smtp.sendgrid.net
Port: 587
User: apikey
Password: Your SendGrid API key
```

#### Other Providers
- **Mailgun**: smtp.mailgun.org:587
- **Amazon SES**: email-smtp.[region].amazonaws.com:587
- **Postmark**: smtp.postmarkapp.com:587

### Setting Credentials in n8n

1. Go to **Credentials** → **New** → **Email (SMTP)**
2. Enter your SMTP settings
3. Test connection
4. Update workflow to use new credentials

## 🔒 Security Best Practices

### Webhook Security

✅ **DO:**
- Use HTTPS webhooks only
- Add authentication token: `webhook-url?token=secret123`
- Implement rate limiting in n8n
- Validate required fields

❌ **DON'T:**
- Expose webhook URLs in public repos
- Accept sensitive data without encryption
- Skip input validation

### CORS Configuration

The workflow includes CORS headers:
```javascript
// Currently allows: https://example.com
// Update in webhook node settings:
"Access-Control-Allow-Origin": "https://your-domain.com"
```

### API Key Management

```bash
# Never commit .env files
echo ".env" >> .gitignore

# Use GitHub secrets for CI/CD
# Rotate API keys regularly
# Use read-only keys when possible
```

### Input Validation

The workflow includes:
- Email format validation
- Required field checks
- HTML sanitization
- Length limits

## 🐛 Troubleshooting

### Common Issues and Solutions

| Problem | Solution |
|---------|----------|
| **Webhook returns 404** | Ensure workflow is active in n8n |
| **No email received** | Check Mailhog (local) or SMTP credentials (production) |
| **CORS error** | Update allowed origins in webhook node |
| **Form submission fails** | Check browser console for errors |
| **Workflow not syncing** | Verify API key and instance URL |
| **Docker won't start** | Check ports 5678, 8025, 1025 are available |

### Debug Commands

```bash
# Check if services are running
docker compose ps

# View detailed logs
docker compose logs -f n8n

# Test n8n API connection
curl -H "X-N8N-API-KEY: $N8N_API_KEY" \
  $N8N_INSTANCE_URL/api/v1/workflows

# Check webhook directly
curl -X POST your-webhook-url \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

### Getting Help

1. Check n8n execution logs for detailed error messages
2. Enable debug mode: `DEBUG=true task up`
3. Join [n8n Community Forum](https://community.n8n.io/)
4. Report issues on [GitHub](https://github.com/edenreich/n8n-cli/issues)

## 📚 Resources

### Documentation
- [n8n-cli Complete Guide](https://github.com/edenreich/n8n-cli)
- [n8n Official Docs](https://docs.n8n.io/)
- [Webhook Node Reference](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.webhook/)
- [Email Node Reference](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.emailsend/)

### Tutorials
- [Building Forms with n8n](https://docs.n8n.io/courses/level-one/chapter-5/)
- [Email Automation](https://n8n.io/blog/email-automation/)
- [Webhook Security](https://docs.n8n.io/hosting/security/)

### Community
- [n8n Community Forum](https://community.n8n.io/)
- [Discord Server](https://discord.gg/n8n)
- [GitHub Discussions](https://github.com/n8n-io/n8n/discussions)

## 📄 License

This example is provided under the MIT License. Feel free to use and modify for your projects!