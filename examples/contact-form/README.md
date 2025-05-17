# Contact Form Example with n8n-cli

This example demonstrates how to set up a contact form workflow in n8n and synchronize it using the n8n-cli and GitHub Actions.

## Overview

This example includes:

1. **Contact form workflow** - An n8n workflow that receives form submissions via webhook and sends email notifications
2. **HTML contact form** - A simple HTML contact form that submits data to the n8n webhook
3. **GitHub Actions workflow** - Automates the synchronization of workflows to your n8n instance

## Directory Structure

```
├── .github/
│   └── workflows/
│       └── sync-n8n.yml     # GitHub Actions workflow for syncing changes automatically to n8n
├── workflows/
│   └── Contact_Form.yaml    # n8n workflow definition
├── .env.example             # Example environment configuration
├── .gitignore               # Git ignore file
├── contact-form.html        # Sample HTML contact form
├── Taskfile.yaml            # Task definitions for common operations
└── README.md                # This documentation
```

## Setup Instructions

### 1. Set Up Your n8n Instance

If you don't already have an n8n instance, you can:

- Use the [n8n cloud](https://www.n8n.cloud/)
- Self-host using [Docker](https://docs.n8n.io/hosting/installation/docker/)
- Install [locally](https://docs.n8n.io/hosting/installation/npm/)

### 2. Obtain n8n API Key

1. Log in to your n8n instance
2. Go to Settings > API
3. Create a new API key with appropriate permissions

### 3. Configure GitHub Repository Secrets

If you're using GitHub Actions for automation, add these secrets to your repository:

1. Go to your GitHub repository > Settings > Secrets and variables > Actions
2. Add these secrets:
   - `N8N_API_KEY`: Your n8n API key
   - `N8N_INSTANCE_URL`: URL of your n8n instance (e.g., `https://your-instance.n8n.cloud`)

### 4. Sync Workflow to n8n

#### Option 1: Manual Sync

Use the n8n-cli to sync workflows manually:

```bash
# Option A: Using Taskfile (recommended)
# First make sure you've created a .env file with your credentials
task setup-env
# Edit the .env file with your actual credentials
nano .env
# Then run the sync task
task sync

# Option B: Direct CLI usage
# Set environment variables
export N8N_API_KEY=your_n8n_api_key
export N8N_INSTANCE_URL=https://your-instance.n8n.cloud
# Sync workflows
n8n workflows sync --directory workflows/
```

#### Option 2: Automatic Sync with GitHub Actions

Push changes to your repository and the GitHub Actions workflow will automatically sync your workflows to n8n.

### 5. Configure the Contact Form

After syncing, you need to:

1. Get the webhook URL from your n8n workflow:

   - Open the Contact Form workflow in n8n
   - Click on the Webhook node
   - Copy the webhook URL

2. Update the HTML form:
   - Open `contact-form.html`
   - Replace `YOUR_N8N_WEBHOOK_URL_HERE` with your actual webhook URL
   - Host the HTML form on your website

## Testing the Contact Form

1. Open the HTML form in a web browser
2. Fill out the form and submit
3. Check the configured email address for the notification
4. You should receive an email with the form submission details

## Customizing the Workflow

You can customize the workflow by:

1. Editing the `Contact_Form.yaml` file
2. Using Taskfile to sync changes: `task sync`  
   (or running `n8n workflows sync --directory workflows/` directly)
3. Or pushing changes to GitHub to trigger the automatic sync

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

## Notes on Email Configuration

The example workflow uses n8n's Email node. To properly configure email sending:

1. In your n8n instance, go to Settings > Credentials
2. Add SMTP credentials for sending emails
3. Update the Email node in your workflow to use these credentials

## Security Considerations

1. Always use HTTPS for your webhook
2. Consider adding a secret query parameter to your webhook URL
3. Implement CORS restrictions (the example allows only `https://example.com`)
4. Filter and validate all form inputs
5. Keep your n8n API key secure

## Troubleshooting

If you encounter issues:

1. Check webhook configuration in n8n
2. Verify CORS settings if submitting from a different domain
3. Check your email configurations in n8n
4. Examine the n8n logs for any errors
5. Test webhook using a tool like Postman

## Learn More

For more information, refer to:

- [n8n-cli Documentation](https://github.com/edenreich/n8n-cli)
- [n8n Documentation](https://docs.n8n.io/)
- [Webhook Node Documentation](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.webhook/)
