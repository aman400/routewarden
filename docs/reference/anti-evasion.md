# Anti-Evasion Security Engine

Attackers frequently employ URL obfuscation techniques to disguise sensitive paths from simple reverse proxy filters. RouteWarden implements an anti-evasion normalization pipeline before evaluating regex patterns.

---

## Evasion Techniques Defeated

### 1. Multi-Layer URL Encoding
Attackers encode dots and slashes multiple times to bypass naive single-pass decoders.
- **Payload**: `/%252e%252e/.env` or `/%252eenv`
- **Defense**: RouteWarden unescapes URLs recursively until the string stabilizes, successfully reducing `%252e` to `.` and catching the hidden payload.

### 2. Semicolon Matrix Parameters
Some application frameworks (Java/Spring, PHP, Tomcat) interpret semicolons as path parameters or session IDs (`/app;jsessionid=.../.env`). Reverse proxies that only inspect paths before the semicolon may incorrectly forward the request.
- **Payload**: `/public;param=1/.env`
- **Defense**: RouteWarden analyzes both the raw candidate path and stripped segment permutations.

### 3. Backslash Path Separators
On Windows/IIS servers or backend services that treat `\` identically to `/`, attackers use backslashes to bypass forward-slash based regexes.
- **Payload**: `/static\..\.env` or `\.env`
- **Defense**: RouteWarden converts backslashes to forward slashes during path normalization.

### 4. Null Byte Injection
Attackers append null bytes to confuse string termination in C-based runtimes or legacy file parsers.
- **Payload**: `/.env%00.html`
- **Defense**: Encoded and raw null bytes (`\x00`, `%00`) are stripped during candidate expansion.

### 5. Directory Traversal Canonicalization
Complex dot-segment sequences attempting to climb out of public folders (`/assets/../../.env`) are collapsed via Go's `path.Clean` logic.
