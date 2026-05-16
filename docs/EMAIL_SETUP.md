# Email Configuration Guide

## Mailtrap Setup

### Step 1: Get API Token
1. Go to [Mailtrap.io](https://mailtrap.io)
2. Sign up / Login
3. Go to **Settings** → **API Tokens**
4. Copy your **API Token**

### Step 2: Configure Environment
Add to your `.env` file:

```env
MAILTRAP_API_TOKEN="your_api_token_here"
MAILTRAP_FROM_EMAIL="your-email@mailtrap.io"
MAILTRAP_FROM_NAME="CWD Forum"
```
