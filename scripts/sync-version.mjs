import fs from 'node:fs'
import path from 'node:path'

// Read target version from single source of truth
const versionFile = path.resolve('docs/version.json')
if (!fs.existsSync(versionFile)) {
  console.error(`Error: ${versionFile} not found`)
  process.exit(1)
}

const { version } = JSON.parse(fs.readFileSync(versionFile, 'utf8'))
console.log(`🔄 Syncing RouteWarden version: ${version}`)

// Files to sync with regex replacements
const filesToSync = [
  'README.md',
  'examples/01-basic-sensitive-files/docker-compose.yml',
  'examples/02-global-entrypoint-shield/docker-compose.yml',
  'examples/03-ip-whitelist-vpn/docker-compose.yml',
  'examples/04-captcha-challenge/docker-compose.yml',
  'examples/05-kubernetes-ingressroute/README.md'
]

let updatedCount = 0

for (const relPath of filesToSync) {
  const filePath = path.resolve(relPath)
  if (!fs.existsSync(filePath)) {
    continue
  }

  let content = fs.readFileSync(filePath, 'utf8')
  const original = content

  // Match CLI flag: --experimental.plugins.routewarden.version=vX.Y.Z
  content = content.replace(
    /(--experimental\.plugins\.routewarden\.version=)v?[0-9]+\.[0-9]+\.[0-9]+/g,
    `$1${version}`
  )

  // Match YAML property: version: vX.Y.Z
  content = content.replace(
    /(version:\s+)v?[0-9]+\.[0-9]+\.[0-9]+/g,
    `$1${version}`
  )

  if (content !== original) {
    fs.writeFileSync(filePath, content, 'utf8')
    console.log(`  ✓ Updated ${relPath}`)
    updatedCount++
  } else {
    console.log(`  - Up to date: ${relPath}`)
  }
}

// Also ensure package.json version matches (without leading 'v')
const pkgPath = path.resolve('package.json')
if (fs.existsSync(pkgPath)) {
  const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'))
  const semver = version.replace(/^v/, '')
  if (pkg.version !== semver) {
    pkg.version = semver
    fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + '\n', 'utf8')
    console.log(`  ✓ Updated package.json to ${semver}`)
  }
}

console.log(`✨ Successfully synced version to ${updatedCount} file(s)!`)
