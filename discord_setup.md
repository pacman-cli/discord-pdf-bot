## Discord Bot Setup Guide

To get your bot up and running on Discord, follow these steps:

### 1. Create a Discord Application
1. Go to the [Discord Developer Portal](https://discord.com/developers/applications).
2. Click the **New Application** button in the top right.
3. Give your application a name (e.g., "PDF Bot") and click **Create**.

### 2. Configure the Bot
1. In the left sidebar, click on **Bot**.
2. Scroll down to the **Privileged Gateway Intents** section.
3. Since your bot might need to read messages or server info depending on how commands are set up in the future, it's a good idea to enable:
   - **Server Members Intent** (if it tracks user roles deeply)
   - **Message Content Intent** (if it ever reads non-slash commands)
   *Note: Slash commands generally don't require privileged intents, but it's safe to turn these on while developing.*
4. Click **Save Changes**.

### 3. Get Your Bot Token
1. On the same **Bot** page, click the **Reset Token** button to generate your token.
2. Click **Yes, do it!**
3. **Copy the token** and save it somewhere secure. *You will need this for Render.*

### 4. Invite the Bot to Your Server
1. In the left sidebar, go to **OAuth2 -> URL Generator**.
2. Under **Scopes**, select `bot` and `applications.commands`.
3. Under **Bot Permissions**, select the permissions your bot needs (e.g., `Send Messages`, `Read Messages/View Channels`, `Attach Files`, `Embed Links`).
4. Copy the generated URL at the bottom of the page.
5. Paste the URL into your browser, select the server you want to add the bot to, and click **Authorize**.

### 5. Get Your Guild (Server) ID
To sync commands specifically to your server (which happens instantly unlike global commands):
1. In Discord, open your **User Settings** (the gear icon).
2. Go to **Advanced** and turn on **Developer Mode**.
3. Right-click on your server's icon in the left server list and select **Copy Server ID**.
4. Save this ID. *You will need this for Render.*
