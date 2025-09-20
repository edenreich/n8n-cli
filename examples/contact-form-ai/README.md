# AI-Enhanced Contact Form Example with n8n-cli

An advanced example demonstrating how to build, deploy, and manage an AI-powered contact form workflow using n8n-cli with automated GitHub Actions deployment and intelligent message processing.

## Table of Contents

- [🎯 What You'll Learn](#-what-youll-learn)
- [📦 What's Included](#-whats-included)
- [📁 Project Structure](#-project-structure)
- [🚀 Quick Start](#-quick-start)
  - [Prerequisites](#prerequisites)
  - [Step 1: Clone and Setup](#step-1-clone-and-setup)
  - [Step 2: Choose Your n8n Setup](#step-2-choose-your-n8n-setup)
  - [Step 3: Configure Environment](#step-3-configure-environment)
  - [Step 4: Set Up AI Credentials](#step-4-set-up-ai-credentials)
  - [Step 5: Deploy the Workflow](#step-5-deploy-the-workflow)
  - [Step 6: Configure the Contact Form](#step-6-configure-the-contact-form)
- [✅ Testing Your Setup](#-testing-your-setup)
  - [Local Testing with Mailhog](#local-testing-with-mailhog)
  - [Testing AI Processing](#testing-ai-processing)
  - [Production Testing](#production-testing)
  - [Debugging Tips](#debugging-tips)
- [🤖 AI Features](#-ai-features)
  - [Message Analysis](#message-analysis)
  - [Sentiment Detection](#sentiment-detection)
  - [Auto-Categorization](#auto-categorization)
  - [Response Suggestions](#response-suggestions)
- [🎨 Customization Guide](#-customization-guide)
  - [Workflow Modifications](#workflow-modifications)
  - [AI Model Configuration](#ai-model-configuration)
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
  - [API Key Management](#api-key-management)
  - [AI Data Privacy](#ai-data-privacy)
  - [Input Validation](#input-validation)
- [🐛 Troubleshooting](#-troubleshooting)
  - [Common Issues and Solutions](#common-issues-and-solutions)
  - [AI-Specific Issues](#ai-specific-issues)
  - [Debug Commands](#debug-commands)
  - [Getting Help](#getting-help)
- [📚 Resources](#-resources)
  - [Documentation](#documentation)
  - [Tutorials](#tutorials)
  - [Community](#community)
- [📄 License](#-license)

## 🎯 What You'll Learn

- Setting up an AI-enhanced webhook-triggered n8n workflow
- Integrating Groq AI for intelligent message processing
- Implementing sentiment analysis and auto-categorization
- Generating AI-powered response suggestions
- Using n8n-cli for local development and workflow synchronization
- Implementing CI/CD with GitHub Actions for automatic deployments
- Testing AI workflows locally with Mailhog

## 📦 What's Included

| Component | Description |
|-----------|-------------|
| **AI Contact Form Workflow** | Advanced n8n workflow with Groq AI integration for intelligent processing |
| **HTML Contact Form** | Modern responsive form with priority selection and AI indicators |
| **Docker Compose Setup** | Local development environment with n8n, Mailhog, and Nginx |
| **GitHub Actions** | Automated CI/CD pipeline for workflow deployment |
| **Taskfile** | Convenient commands for common development tasks |
| **AI Processing** | Sentiment analysis, categorization, and response suggestions |

## 📁 Project Structure

```
├── .github/
│   └── workflows/
│       └── sync-n8n.yml        # CI/CD pipeline for automatic n8n deployment
├── workflows/
│   └── Contact_Form_AI.yaml    # AI-enhanced n8n workflow (webhook → AI → email)
├── docker-compose.yaml         # Local development environment setup
├── .env.example                # Environment variable template
├── .gitignore                  # Version control exclusions
├── contact-form-ai.html        # Frontend form with AI features
├── Taskfile.yaml               # Development automation commands
└── README.md                   # Project documentation
```

## 🚀 Quick Start

### Prerequisites

- [Docker](https://www.docker.com/get-started) and Docker Compose
- [n8n-cli](https://github.com/edenreich/n8n-cli#installation) installed
- [Task](https://taskfile.dev/installation/) (optional, for using Taskfile commands)
- [Groq API Key](https://console.groq.com/) for AI processing
- Git for version control

### Step 1: Clone and Setup

```bash
# Clone this example
git clone <your-repo-url>
cd contact-form-ai

# Create environment configuration
cp .env.example .env
```

### Step 2: Choose Your n8n Setup

#### Option A: Local Development (Recommended for Testing)

```bash
# Start n8n, mailhog, and nginx locally
task up

# n8n will be available at http://localhost:5678
# Mailhog will be available at http://localhost:8025
# AI Contact form will be available at http://localhost:8080
```

#### Option B: Use n8n Cloud

1. Sign up at [n8n.cloud](https://n8n.cloud)
2. Update `.env` with your cloud instance URL
3. Generate an API key from Settings → API

#### Option C: Self-Hosted n8n

Follow the [n8n self-hosting guide](https://docs.n8n.io/hosting/)

### Step 3: Configure Environment

Edit `.env` file with your credentials:

```bash
# For local development
N8N_API_KEY=
N8N_INSTANCE_URL=http://localhost:5678

# For production
# N8N_API_KEY=your_actual_api_key
# N8N_INSTANCE_URL=https://your-instance.n8n.cloud
```

### Step 4: Set Up AI Credentials

#### In n8n Interface:

1. Navigate to **Settings → Credentials**
2. Click **New Credential**
3. Select **Groq API**
4. Enter your Groq API key
5. Name it "Groq AI" (or update the workflow if using a different name)
6. Save the credential

#### Getting a Groq API Key:

1. Visit [console.groq.com](https://console.groq.com/)
2. Sign up or log in
3. Navigate to API Keys
4. Create a new API key
5. Copy and save it securely

### Step 5: Deploy the Workflow

```bash
# Preview what will be synced (dry run)
task sync-dry-run

# Deploy the workflow to n8n
task sync

# Verify deployment
task list
```

### Step 6: Configure the Contact Form

1. Get your webhook URL from n8n:
   ```bash
   # List workflows to see webhook URLs
   task list
   ```

2. Update the form HTML:
   ```javascript
   // In contact-form-ai.html, update line ~150:
   const webhookUrl = 'YOUR_WEBHOOK_URL_HERE';
   ```

3. Test the form:
   ```bash
   # Use the included nginx server
   task up
   # Open http://localhost:8080
   ```

## ✅ Testing Your Setup

### Local Testing with Mailhog

1. Start the local environment:
   ```bash
   task up
   ```

2. Submit a test message through the form

3. Check Mailhog at http://localhost:8025 to see:
   - Email with AI-generated summary
   - Sentiment analysis results
   - Auto-categorization
   - Suggested responses

### Testing AI Processing

Submit different types of messages to test AI capabilities:

- **Positive feedback**: "Your service is amazing! I'm very happy with the results."
- **Support request**: "I'm having trouble logging in to my account."
- **Sales inquiry**: "I'd like to know more about your pricing plans."
- **Bug report**: "The checkout button doesn't work on mobile devices."

### Production Testing

1. Ensure workflow is active:
   ```bash
   task list  # Check status column
   ```

2. Monitor executions:
   ```bash
   task executions
   ```

3. Check execution details in n8n UI under "Executions"

### Debugging Tips

```bash
# Check n8n logs
task logs

# Check mailhog logs
task logs-mail

# View recent workflow executions
task executions

# Restart services if needed
task restart
```

## 🤖 AI Features

### Message Analysis

The workflow uses Groq's LLMs to analyze incoming messages:

- **Summary Generation**: Concise overview of the message
- **Key Points Extraction**: Main topics and concerns
- **Intent Detection**: What the sender wants to achieve

### Sentiment Detection

Automatically identifies the emotional tone:

- 😊 **Positive**: Happy, satisfied, grateful
- 😐 **Neutral**: Informational, factual
- 😟 **Negative**: Frustrated, disappointed, angry

### Auto-Categorization

Messages are automatically sorted into categories:

- **Support**: Technical issues, help requests
- **Sales**: Pricing, features, purchasing
- **Feedback**: Reviews, suggestions, complaints
- **General**: Other inquiries

### Response Suggestions

AI generates appropriate response templates based on:

- Message content and context
- Detected sentiment
- Category and priority
- Common response patterns

## 🎨 Customization Guide

### Workflow Modifications

1. Open the workflow in n8n editor
2. Common modifications:
   - Change AI model (default: mixtral-8x7b-32768)
   - Adjust prompt templates
   - Add custom categories
   - Modify email template

### AI Model Configuration

In the Groq node, you can use different models:

```javascript
// Available models:
- mixtral-8x7b-32768 (default, balanced)
- llama2-70b-4096 (more creative)
- gemma-7b-it (faster, lighter)
```

### Common Customizations

1. **Custom Categories**:
   Edit the AI prompt in the workflow to include your categories

2. **Priority Rules**:
   Modify the priority logic based on keywords or sentiment

3. **Email Templates**:
   Customize the HTML email template in the Send Email node

4. **Additional AI Analysis**:
   Add more Groq nodes for specialized processing

### Deployment After Changes

```bash
# After modifying workflows locally
task refresh  # Pull changes from n8n

# Or after modifying in n8n UI
task sync     # Push changes to n8n
```

## 🛠️ Development Commands

### Quick Reference

```bash
task help         # Show all available commands
task up           # Start local environment
task down         # Stop local environment
task sync         # Deploy workflows to n8n
task refresh      # Pull workflows from n8n
task list         # List all workflows
task executions   # View execution history
```

### Docker Environment

```bash
# Container management
task up          # Start n8n and mailhog
task down        # Stop containers
task restart     # Restart all services
task clean       # Remove containers and volumes

# Logs and debugging
task logs        # View n8n logs
task logs-mail   # View mailhog logs
```

### Workflow Management

```bash
# Synchronization
task sync           # Deploy workflows
task sync-dry-run   # Preview changes
task refresh        # Download workflows
task refresh-all    # Download all workflows

# Monitoring
task list           # List workflows with status
task executions     # View execution history
```

### Development Tools

```bash
# Environment setup
task setup-env   # Create .env from template
```

## 📧 Email Configuration

### Local Development (Mailhog)

Mailhog is pre-configured in docker-compose.yaml:

- SMTP Server: `mailhog:1025`
- Web Interface: `http://localhost:8025`
- No authentication required

### Production Email Setup

Configure your preferred email service in n8n:

1. **Gmail**:
   - Enable 2FA and create app password
   - Use OAuth2 for better security

2. **SendGrid**:
   - Get API key from SendGrid dashboard
   - Use SendGrid node instead of Email node

3. **Custom SMTP**:
   - Configure host, port, security settings
   - Add credentials in n8n

### Setting Credentials in n8n

1. Navigate to **Settings → Credentials**
2. Click **New Credential**
3. Select your email service type
4. Configure authentication
5. Test the connection
6. Update workflow to use new credentials

## 🔒 Security Best Practices

### Webhook Security

1. **Use HTTPS in production**:
   Ensure webhook URLs use HTTPS

2. **Implement rate limiting**:
   Configure n8n rate limits

3. **Add webhook authentication**:
   Use webhook passwords or headers

### API Key Management

1. **Never commit credentials**:
   Use `.env` files (already in .gitignore)

2. **Rotate keys regularly**:
   Update API keys periodically

3. **Use GitHub Secrets**:
   Store production credentials securely

### AI Data Privacy

1. **Sanitize sensitive data**:
   Remove PII before sending to AI

2. **Use appropriate models**:
   Choose models based on data sensitivity

3. **Log retention policies**:
   Configure appropriate data retention

### Input Validation

The workflow includes validation for:

- Email format verification
- Message length limits
- Required field checking
- XSS prevention in form

## 🐛 Troubleshooting

### Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| Workflow not triggering | Check webhook URL is correct and workflow is active |
| Emails not sending | Verify email credentials in n8n |
| AI processing fails | Check Groq API key and quota |
| Form submission fails | Check CORS settings and webhook URL |
| Container won't start | Ensure ports 5678, 8025 are free |

### AI-Specific Issues

1. **"Invalid API Key" error**:
   - Verify Groq API key in n8n credentials
   - Check key hasn't expired

2. **"Rate limit exceeded"**:
   - Check Groq API quotas
   - Implement request throttling

3. **Poor AI responses**:
   - Try different models
   - Refine prompt templates
   - Add more context

### Debug Commands

```bash
# Check service status
docker-compose ps

# View detailed logs
task logs | grep ERROR

# Test webhook manually
curl -X POST YOUR_WEBHOOK_URL \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","message":"Test message"}'

# Check n8n API connectivity
curl -H "X-N8N-API-KEY: $N8N_API_KEY" \
  $N8N_INSTANCE_URL/api/v1/workflows
```

### Getting Help

1. Check [n8n documentation](https://docs.n8n.io)
2. Visit [n8n community forum](https://community.n8n.io)
3. Review [Groq documentation](https://console.groq.com/docs)
4. Open an issue in this repository

## 📚 Resources

### Documentation

- [n8n Official Docs](https://docs.n8n.io)
- [n8n-cli Documentation](https://github.com/edenreich/n8n-cli)
- [Groq API Reference](https://console.groq.com/docs/api-reference)
- [Task Documentation](https://taskfile.dev)

### Tutorials

- [n8n Workflow Basics](https://docs.n8n.io/workflows/)
- [Webhook Triggers](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.webhook/)
- [AI Integration Guide](https://docs.n8n.io/integrations/builtin/cluster-nodes/sub-nodes/n8n-nodes-langchain.lmchatgroq/)

### Community

- [n8n Community Forum](https://community.n8n.io)
- [n8n Discord](https://discord.gg/n8n)
- [Groq Discord](https://discord.gg/groq)

## 📄 License

This example is provided as-is for educational purposes. Feel free to use and modify for your own projects.