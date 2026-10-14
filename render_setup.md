## Render Deployment Guide

The Discord bot code has been successfully linked to Render as a Web Service. However, to get it fully functional, there are two important things to complete.

### 1. Update the Code to Pass Health Checks (Already Done Locally)

Render Web Services expect an HTTP server to be running and listening on the `$PORT` provided by Render to pass its deployment health check. Since this is a Discord bot, we patched `cmd/bot/main.go` to add a dummy HTTP server.

*We've already committed this change locally.* All you need to do is push it to your GitHub repository to trigger a new build on Render:
`git push origin main`

### 2. Set Environment Variables in Render

The bot requires the Discord Token and Guild ID to start.

1. Go to your [Render Dashboard](https://dashboard.render.com).
2. Click on the `discord-pdf-bot` Web Service.
3. In the left sidebar, click on **Environment**.
4. Add the following **Environment Variables**:
   - `DISCORD_BOT_TOKEN`: Paste the token you copied from the Discord Developer Portal.
   - `GUILD_ID`: Paste the ID of your Discord server.
   - `ADMIN_ROLE`: Give it the name of the role in your server that should have admin access (e.g., `PDF Admin`).
5. Save changes. This will automatically trigger a new deployment.

### 3. Add Persistent Storage (Disk)

Since SQLite and the PDF files are stored locally, a standard Web Service will lose this data on every restart. You should add a Persistent Disk:

1. In the Render Dashboard for your service, click on **Disks**.
2. Add a new disk:
   - **Name**: `bot-data`
   - **Mount Path**: `/opt/render/project/src/data` (for SQLite db)
3. Save changes.

*Note: You may need to upgrade from the Free tier to attach disks on Render.*
