# Example 04: Captcha Challenge (hCaptcha & Cloudflare Turnstile)

This scenario demonstrates using RouteWarden to challenge requests to sensitive endpoints with **hCaptcha** and **Cloudflare Turnstile**.

## Is a `captcha.html` File Required?

> [!NOTE]
> **No external `captcha.html` file is needed!**  
> RouteWarden has a **built-in, self-contained HTML template** compiled directly into the binary with modern responsive styling and dark mode.
>
> When `response.mode: captcha` is set, RouteWarden dynamically injects the appropriate provider SDK script (`https://js.hcaptcha.com/1/api.js` or `https://challenges.cloudflare.com/turnstile/v0/api.js`) and renders the widget seamlessly.
>
> *(Optional: If you want to supply your own custom layout, you can pass an HTML template string into `response.captcha.template`).*

---

## Running the Example

```bash
docker compose up -d
```

---

## Verification

### 1. Test hCaptcha Service
```bash
curl -i -H "Host: hcaptcha.localhost" http://localhost/login
# Returns: HTTP/1.1 403 Forbidden with hCaptcha JS and <div class="h-captcha">
```

Or open in your browser: [http://localhost/login](http://localhost/login) with `Host: hcaptcha.localhost`.

### 2. Test Cloudflare Turnstile Service
```bash
curl -i -H "Host: turnstile.localhost" http://localhost/portal
# Returns: HTTP/1.1 403 Forbidden with Turnstile JS and <div class="cf-turnstile">
```
