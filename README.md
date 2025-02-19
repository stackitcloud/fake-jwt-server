# Introduction

A straightforward mock-up of an identity provider (IDP) is designed to issue JWTs, enhancing the security of
communications between services.

![fake-jwt-server-illustration](./media/fake-jwt-illustration.png)

To safeguard service interactions, various methods are available such as mutual TLS, basic authentication, or bearer
authentication, to name a few. In bearer authentication, a token is required from an IDP, like Keycloak or Okta, through
the client credentials grant of OAuth 2.0. Once the JWT is acquired, it can be transmitted in the authorization header
when a service communicates with another. The recipient service can then verify the token using the IDP's public key.
During local development or integration testing, utilizing a real IDP may not be desirable. This is where the concept of
a fake-jwt-server is introduced. It's a basic version of an IDP that issues JWTs for OAuth flows and provides a public
key endpoint for token verification.

# Running the Server

To launch the server in a Docker container, execute the following command:

```bash
docker run -p 8008:8008 ghcr.io/stackitcloud/fake-jwt-server:v0.1.1
```

This command initializes the server on port 8008. The public key can be accessed
at http://localhost:8008/.well-known/jwks.json, and the OAuth token endpoint is available
at http://localhost:8008/token.

# Configuration

The server's settings can be adjusted using specified environment variables and flags.

| Environment Variable | Flag                   | Description                                                               |
|----------------------|------------------------|---------------------------------------------------------------------------|
| `PORT`               | `--port`               | The port the server listens on. Defaults to `8008`.                       |
| `ISSUER`             | `--issuer`             | The issuer of the tokens. Defaults to `test`.                             |
| `AUDIENCE`           | `--audience `          | The audience of the tokens. Defaults to `test `.                          |
| `SUBJECT`            | `--subject`            | The subject of the tokens. Defaults to `test`                             |
| `ID`                 | `--id`                 | The id of the tokens. Defaults to `test`.                                 |
| `EXPIRES_IN_MINUTES` | `--expires-in-minutes` | The expiration time of the JWT tokens in minutes. Defaults to `52560000`. |
| `GRAND_TYPE`         | `--grand-type`         | The grand type of the JWT tokens. Defaults to `client_credentials`.       |
| `EMAIL`              | `--email`              | The email of the JWT token. Defaults to `test@example.com`.               |

# Collaboration with Bruno


Bruno is our favourite request testing tool.

https://docs.usebruno.com/introduction/what-is-bruno

Therefore, an introduction to how the tokens can be integrated into Bruno.

The workflow is as follows Brono will perform a pre-request against the fake-jwt-server before each request and add the token as header to the actual request.

## Script

```javascript
const tokenUrl = 'http://localhost:8008/token';
try {
    let resp = await axios({
        method: 'POST',
        url: tokenUrl,
    });
    bru.setVar('ACCESS_TOKEN', resp.data.access_token);
} catch (error) {
    throw error;
}
```

## Integration

You can make settings for the entire collection.
The script above is stored in this as a pre-request script.
![bruno - collection script](../fake-jwt-server-fapo/media/bruno-collection-script.png)

The token is stored in the variable ACCESS_TOKEN in the script.

This must be added to the requests as a header.
![bruno - collection headers](../fake-jwt-server-fapo/media/bruno-collection-headers.png)

## Non Local Environment

The following script can be used to set the token depending on the environment.
I am not yet fully satisfied with this solution, so I will update the readme when new findings come to light.
```javascript
if (!bru.getEnvName("local")) {
    bru.setVar('ACCESS_TOKEN', "");
    return
}
```
