# Provider Configuration
- [Provider Configuration](#provider-configuration)
  - [SMTP provider configuration](#smtp-provider-configuration)
  - [Gmail provider configuration](#gmail-provider-configuration)
## SMTP provider configuration
For example use the [mailsend](https://www.mailersend.com/) as a host.

Disclaimer: The free/trial account can only send the mail to the same mail as the mailsender owner's mail.

**SMTP Configuration**
- `SMTP_SERVER_HOST`: The hostname of the SMTP server
- `SMTP_SERVER_PORT`: The port of the SMTP server
- `SMTP_SENDER`: The sender email address
- `SMTP_PASSWORD`: The password for the email sender

## Gmail provider configuration
Follow Google OAuth 2.0 [tutorial](https://developers.google.com/identity/protocols/oauth2) for `CLIENT_ID` and `CLIENT_SECRET` then enable the Gmail API at the API Cloud Console
1. Go to APIs & Services → Library
2. Search for Gmail API
3. Click Enable

Don't for get to add the <u>**redirect uri**</u>.
Use the following url to get the authorization code first.
```
https://accounts.google.com/o/oauth2/v2/auth?client_id=your_client_id&redirect_uri=your_redirect_uri&response_type=code&scope=https://mail.google.com/&access_type=offline&prompt=consent
```

then use curl to get access and refresh token.
```curl
curl -X POST "https://oauth2.googleapis.com/token" 
    -H "Content-Type: application/x-www-form-urlencoded" 
    -d "grant_type=authorization_code" 
    -d "code=[Authorization Code]" 
    -d "client_id=[Client ID]" 
    -d "client_secret=[Client Secret]" 
    -d "redirect_uri=http://localhost:8000/google/login"
```

After you curl you will get the response with this format:
```
{
  "access_token": "your_access_token",
  "expires_in": 3599,
  "refresh_token": "your_refresh_token",
  "scope": "https://mail.google.com/",
  "token_type": "Bearer"
}
```
You need to redo the curl steps every time you start a server and the token is expired.

**Gmail Configuration**
- `GMAIL_OAUTH_CLIENT_ID`: The client id of your Google OAuth application
- `GMAIL_OAUTH_CLIENT_SECRET`: The client secret of your Google OAuth application
- `GMAIL_OAUTH_ACCESS_TOKEN`: The OAuth access token for your OAuth application for a gmail account
- `GMAIL_OAUTH_REFRESH_TOKEN`: The OAuth refresh token for your OAuth application for a gmail account

