# RouteWarden Release & Version Update Guide

This guide explains how versioning works in RouteWarden and how to publish patch updates, minor releases, and new documentation series snapshots.

---

## Architecture & Single Source of Truth

RouteWarden uses [docs/version.json](file:///Users/aman/git/routewarden/docs/version.json) as the **single source of truth** for versioning:

```json
{
  "version": "v0.2.2"
}
```

Whenever this version changes, RouteWarden's automated tooling synchronizes it across:
- Root `package.json` (`version: "0.2.2"`)
- Root `README.md` Traefik CLI flags and YAML configurations
- All `examples/**/docker-compose.yml` Traefik plugin flags (`--experimental.plugins.routewarden.version=...`)
- All VitePress markdown pages using `{{version}}` dynamic interpolation

---

## 1. Patch & Maintenance Updates (e.g. `v0.2.2` ➔ `v0.2.3`)

Patch releases and minor non-breaking fixes do not require creating a new documentation snapshot. The documentation continues to serve the `v0.2.x` series.

### Step-by-Step:
1. **Update `docs/version.json`**:
   ```json
   {
     "version": "v0.2.3"
   }
   ```
2. **Run Version Sync**:
   ```bash
   npm run sync-version
   ```
3. **Verify Changes**:
   ```bash
   npm test
   git diff
   ```
4. **Commit and Tag**:
   ```bash
   git commit -am "chore: release v0.2.3"
   git tag v0.2.3
   git push origin main --tags
   ```

---

## 2. New Major or Minor Series Release (e.g. `v0.2.x` ➔ `v0.3.0`)

When releasing a new version series that introduces breaking changes or significant features warranting an archived version of past documentation, use `npm run docs:release`.

### Command:
```bash
npm run docs:release v0.3.0
```

### What Happens Automatically:
1. **Archives Documentation**:
   - Copies `docs/guide`, `docs/reference`, and `docs/examples` into a frozen snapshot directory `docs/v0.2/`.
   - Injects a `Legacy Version Notice` banner with a one-click link back to Latest at the top of each archived markdown file.
2. **Updates Version Registry (`docs/versions.json`)**:
   - Re-points the previous series (`v0.2.x`) to the archived `/v0.2/guide/getting-started`.
   - Promotes the new series (`v0.3.x (Latest)`) to `/guide/getting-started`.
3. **Updates `docs/version.json`**:
   - Sets target version to `v0.3.0`.
4. **Cascades Version Synchronization**:
   - Replaces plugin version strings in `package.json`, `README.md`, and all `docker-compose.yml` examples.
5. **VitePress Live Preview**:
   - The top navbar dropdown immediately allows users to switch between `v0.3.x (Latest)` and legacy `v0.2.x` snapshots.

---

## 3. Testing & CI Verification

Before pushing any release:

```bash
# Run script unit tests and Go plugin tests
npm run test:all

# Validate VitePress docs build and dead-link check
npm run docs:build
```
