# Slack Integration

![Slack Integration](https://github.com/benwsapp/rlgl/blob/main/img/slack.png)

rlgl can automatically sync your work status to your Slack profile. When you update your `contributor.active` or `contributor.focus` fields, your Slack status will update in real-time.

## How It Works

- Works on **free Slack workspaces** (self-service status updates)
- Automatic sync when config changes
- Customizable emojis
- Non-blocking/async operation

## Setup

### 1. Create a Slack App

- Go to https://api.slack.com/apps
- Click **"Create New App"**
- Choose **"From scratch"**
- Give it a name (e.g., "rlgl Status Sync")
- Select your workspace
- Click **"Create App"**

### 2. Add OAuth Scopes

- In the left sidebar, click **"OAuth & Permissions"**
- Scroll down to **"Scopes"** section
- Under **"User Token Scopes"** (not Bot Token Scopes!), click **"Add an OAuth Scope"**
- Add: `users.profile:write`

### 3. Install to Workspace

- Scroll back up to **"OAuth Tokens for Your Workspace"**
- Click **"Install to Workspace"** (or "Reinstall to Workspace")
- Review the permissions - it will show "Update your profile information"
- Click **"Allow"**

### 4. Copy Your User Token

- After installation, you'll see **"User OAuth Token"**
- It starts with `xoxp-` (not `xoxb-` which is for bots)
- Copy this token

### 5. Update Your Config

Add the Slack configuration to your `rlgl.yaml`:

```yaml
name: "Your Name"
description: "Current work and availability"
user: "username"
contributor:
  active: true
  focus: "Working on feature X"
  queue:
    - "Task 1"
    - "Task 2"
slack:
  enabled: true
  user_token: "xoxp-1234567890-1234567890-1234567890-abcdef1234567890abcdef1234567890"
  status_emoji_active: ":large_green_circle:"
  status_emoji_inactive: ":red_circle:"
```

### 6. Run rlgl

Start your server and client as normal:

```bash
# Terminal 1: Start server
./rlgl serve

# Terminal 2: Start client
./rlgl client --client-id ben-macbook --config config/site.yaml
```

Your Slack status will now automatically sync whenever the config changes!

## Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `enabled` | Enable/disable Slack sync | `false` |
| `user_token` | Your Slack user token (starts with `xoxp-`) | `""` |
| `status_emoji_active` | Emoji when `active: true` | `:large_green_circle:` |
| `status_emoji_inactive` | Emoji when `active: false` | `:red_circle:` |

## Architecture

1. **Client pushes config** to server via WebSocket
2. **Server stores config** in memory
3. **If Slack is enabled**, server automatically calls Slack API to update your status:
   - Status text: Your `contributor.focus` field
   - Status emoji: Based on `contributor.active` (true/false)
4. **Your Slack status updates** in real-time

## Status Mapping

| rlgl Config | Slack Status |
|-------------|--------------|
| `active: true, focus: "Coding feature X"` | 🟢 Coding feature X |
| `active: false, focus: "Coffee break"` | 🔴 Coffee break |
| `active: false, focus: ""` | 🔴 Busy |

## Security

- **Keep your token secret** - don't commit it to git
- Add `rlgl.yaml` (or other config file) to your `.gitignore`
- The token only has permission to update **your own** profile
- You can revoke it anytime from the Slack App settings at https://api.slack.com/apps

## API Details

rlgl uses the Slack Web API `users.profile.set` method:
- Documentation: https://api.slack.com/methods/users.profile.set
- Requires user token with `users.profile:write` scope
- Status text limited to 100 characters (automatically truncated)
- Works on free Slack workspaces
