# AI-Enhanced Contact Form Example with n8n-cli

This example demonstrates how to set up an AI-enhanced contact form workflow in n8n and synchronize it using the n8n-cli and GitHub Actions.

## Overview

This example includes:

1. **AI-Enhanced Contact Form Workflow** - An n8n workflow that:

   - Receives form submissions via webhook
   - Processes the submissions with AI to generate:
     - A summary of the contact message
     - Sentiment analysis
     - Message categorization
     - Response suggestions
   - Sends detailed email notifications with AI insights
   - Returns a success response to the submitter

2. **HTML Contact Form** - A modern HTML form with AI-specific features:

   - Message categorization
   - Priority selection
   - AI processing indicator

3. **GitHub Actions Workflow** - Automates the synchronization of workflows to your n8n instance

## Directory Structure

```
├── .github/
│   └── workflows/
│       └── sync-n8n.yml     # GitHub Actions workflow for automation
├── workflows/
│   └── AI_Contact_Form.yaml # AI-enhanced n8n workflow definition
├── .env.example             # Example environment configuration
├── contact-form-ai.html     # AI-enhanced HTML contact form
├── Taskfile.yaml            # Task definitions for common operations
└── README.md                # This documentation
```

## Setup Instructions

### 1. Set Up Your n8n Instance

If you don't already have an n8n instance, you can:

- Use the [n8n cloud](https://www.n8n.cloud/)
- Self-host using [Docker](https://docs.n8n.io/hosting/installation/docker/)
- Install [locally](https://docs.n8n.io/hosting/installation/npm/)

### 2. Set Up OpenAI Credentials in n8n

The workflow uses OpenAI for AI processing, so you need to set up credentials:

1. In your n8n instance, go to **Settings > Credentials**
2. Click **New Credential**
3. Select **OpenAI API**
4. Enter your OpenAI API key
5. Save the credential

### 3. Obtain n8n API Key

1. Log in to your n8n instance
2. Go to **Settings > API**
3. Create a new API key with appropriate permissions

### 4. Configure GitHub Repository Secrets

If you're using GitHub Actions for automation, add these secrets to your repository:

1. Go to your GitHub repository > Settings > Secrets and variables > Actions
2. Add these secrets:
   - `N8N_API_KEY`: Your n8n API key
   - `N8N_INSTANCE_URL`: URL of your n8n instance (e.g., `https://your-instance.n8n.cloud`)

### 5. Sync Workflow to n8n

#### Option A: Using Taskfile (recommended)

```bash
# First make sure you've created a .env file with your credentials
cd examples/contact-form-ai
task setup-env
# Edit the .env file with your actual credentials
nano .env
# Then run the sync task
task sync
```

#### Option B: Direct CLI Usage

```bash
# Set environment variables
export N8N_API_KEY=your_n8n_api_key
export N8N_INSTANCE_URL=https://your-instance.n8n.cloud
# Sync workflows
n8n workflows sync --directory examples/contact-form-ai/workflows/
```

### 6. Configure the Contact Form

After syncing, you need to:

1. Get the webhook URL from your n8n workflow:

   - Open the AI Contact Form workflow in n8n
   - Click on the Webhook node
   - Copy the webhook URL

2. Update the HTML form:
   - Open `contact-form-ai.html`
   - Replace `YOUR_N8N_WEBHOOK_URL_HERE` with your actual webhook URL
   - Host the HTML form on your website

### 7. Update Email Settings

Make sure to update the email settings in the workflow to use your preferred email address:

1. In your n8n instance, open the AI Contact Form workflow
2. Click on the Email node
3. Update the "From Email" and "To Email" fields

## Using Taskfile

This example uses [Taskfile](https://taskfile.dev/) for common operations. Available tasks:

```bash
# List all available tasks
task help

# Sync workflows to n8n
task sync

# Preview sync without making changes
task sync-dry-run

# Refresh local workflows from n8n instance
task refresh

# List workflows from n8n instance
task list

# Start a local web server to preview the HTML form
task preview

# Create a template .env file
task setup-env
```

## How It Works

1. **Contact Form Submission**: The user fills out the HTML form, selecting a category and priority level.

2. **Webhook Receiver**: The n8n workflow receives the submission via a webhook.

3. **AI Processing**: The OpenAI node analyzes the message and provides:

   - A concise summary
   - Message categorization
   - Sentiment analysis
   - Suggested response

4. **Email Notification**: An email is sent to the admin with:

   - Original message details
   - AI-generated summary and analysis
   - Suggested response to the customer

5. **Webhook Response**: A success message is returned to the user's browser.

## Customizing the AI Prompt

You can customize how the AI processes the contact form messages by editing the prompt in the OpenAI node:

1. In the n8n workflow, click on the "AI Summary" node
2. Modify the "Prompt" field to change how the AI analyzes the messages
3. Update the output format as needed

## Security Considerations

1. **API Key Protection**: Keep your OpenAI and n8n API keys secure
2. **HTTPS**: Always use HTTPS for your webhook
3. **CORS**: The example restricts origins to a specific domain
4. **Input Validation**: The form includes basic validation
5. **Rate Limiting**: Consider implementing rate limiting for the form

## Troubleshooting

If you encounter issues:

1. Check both webhook and OpenAI API configurations in n8n
2. Verify CORS settings if submitting from a different domain
3. Check your email configurations in n8n
4. Review OpenAI API quotas if AI processing fails
5. Examine the n8n logs for any errors

## Learn More

For more information, refer to:

- [n8n-cli Documentation](https://github.com/edenreich/n8n-cli)
- [n8n Documentation](https://docs.n8n.io/)
- [OpenAI Documentation](https://platform.openai.com/docs)
- [Webhook Node Documentation](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.webhook/)
- [OpenAI Node Documentation](https://docs.n8n.io/integrations/builtin/app-nodes/n8n-nodes-base.openai/)
