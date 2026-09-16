# Configuration Reference (v0.1.0)

::: warning Legacy Version Notice
You are currently viewing the documentation for **v0.1.0**.  
For the newest features (such as `allowedIps` CIDR/IP whitelisting), visit [Latest Configuration](/reference/configuration).
:::

## Options in v0.1.0

| Key | Type | Default | Description |
|---|---|---|---|
| `enabled` | `bool` | `true` | Enables or disables the middleware. |
| `enableDefaultPatterns` | `bool` | `true` | Enables built-in sensitive file blocking. |
| `pathPatterns` | `[]string` | `[]` | List of custom regular expressions to block. |
| `blockPatterns` | `[]string` | `[]` | Alias for `pathPatterns`. |
| `allowPatterns` | `[]string` | *(Safe defaults)* | List of regex patterns to explicitly allow. |
| `checkQuery` | `bool` | `false` | Also inspects raw query strings. |
| `statusCode` | `int` | `403` | HTTP status code when request is blocked. |
| `response` | `object` | `{}` | Configures custom response modes (`json`, `html`, `captcha`, `redirect`, `text`). |
