import fs from 'node:fs'
import path from 'node:path'

/**
 * Automates snapshotting of the current documentation when creating a new release.
 * Usage:
 *   node scripts/snapshot-version.mjs <new-version>
 * Example:
 *   node scripts/snapshot-version.mjs v0.3.0
 */

const newVersion = process.argv[2]
if (!newVersion) {
  console.error('Usage: npm run docs:release <new-version>')
  console.error('Example: npm run docs:release v0.3.0')
  process.exit(1)
}

const cleanNewVersion = newVersion.startsWith('v') ? newVersion : `v${newVersion}`
const versionFilePath = path.resolve('docs/version.json')
const versionsRegistryPath = path.resolve('docs/versions.json')

if (!fs.existsSync(versionFilePath) || !fs.existsSync(versionsRegistryPath)) {
  console.error('Missing docs/version.json or docs/versions.json')
  process.exit(1)
}

const currentVersionData = JSON.parse(fs.readFileSync(versionFilePath, 'utf8'))
const currentVersion = currentVersionData.version
const registry = JSON.parse(fs.readFileSync(versionsRegistryPath, 'utf8'))

if (currentVersion === cleanNewVersion) {
  console.log(`Version is already set to ${cleanNewVersion}.`)
  process.exit(0)
}

console.log(`📦 Snapshotting current documentation (${currentVersion}) before releasing ${cleanNewVersion}...`)

// Major.Minor directory name (e.g. v0.2.1 -> v0.2)
const versionParts = currentVersion.replace(/^v/, '').split('.')
const minorSnapshotDirName = `v${versionParts[0]}.${versionParts[1]}`
const targetSnapshotDir = path.resolve(`docs/${minorSnapshotDirName}`)

// 1. Recursive copy function
function copyDir(src, dest, transformFile) {
  fs.mkdirSync(dest, { recursive: true })
  const entries = fs.readdirSync(src, { withFileTypes: true })

  for (const entry of entries) {
    const srcPath = path.join(src, entry.name)
    const destPath = path.join(dest, entry.name)

    if (entry.isDirectory()) {
      copyDir(srcPath, destPath, transformFile)
    } else if (entry.isFile()) {
      let content = fs.readFileSync(srcPath, 'utf8')
      if (transformFile) {
        content = transformFile(content, entry.name)
      }
      fs.writeFileSync(destPath, content, 'utf8')
    }
  }
}

// 2. Snapshot current guide, reference, and examples
const dirsToSnapshot = ['guide', 'reference', 'examples']
for (const dir of dirsToSnapshot) {
  const srcDir = path.resolve(`docs/${dir}`)
  const destDir = path.resolve(`docs/${minorSnapshotDirName}/${dir}`)
  if (fs.existsSync(srcDir)) {
    copyDir(srcDir, destDir, (content) => {
      // Replace dynamic template with the frozen version string
      let replaced = content.replace(/\{\{version\}\}/g, currentVersion)
      // Inject notice banner at top of markdown files if not already present
      if (!replaced.includes('Legacy Version Notice') && replaced.startsWith('# ')) {
        const firstLineEnd = replaced.indexOf('\n')
        const title = replaced.substring(0, firstLineEnd)
        const rest = replaced.substring(firstLineEnd)
        const banner = `\n\n::: warning Legacy Version Notice\nYou are viewing archived documentation for **${currentVersion}**. [Switch to Latest ➔](/guide/getting-started)\n:::`
        return `${title} (${currentVersion})${banner}${rest}`
      }
      return replaced
    })
    console.log(`  ✓ Archived docs/${dir} -> docs/${minorSnapshotDirName}/${dir}`)
  }
}

// 3. Update docs/versions.json
// Re-point previous latest to its archived snapshot
const updatedVersions = registry.versions.filter(v => v.tag !== cleanNewVersion && v.tag !== currentVersion)

// Add new latest version at top
const newRegistryVersions = [
  {
    text: `${cleanNewVersion} (Latest)`,
    link: '/guide/getting-started',
    tag: cleanNewVersion
  },
  {
    text: `${currentVersion}`,
    link: `/${minorSnapshotDirName}/guide/getting-started`,
    tag: currentVersion
  },
  ...updatedVersions
]

registry.current = cleanNewVersion
registry.versions = newRegistryVersions
fs.writeFileSync(versionsRegistryPath, JSON.stringify(registry, null, 2) + '\n', 'utf8')
console.log(`  ✓ Updated docs/versions.json`)

// 4. Update docs/version.json to new version
fs.writeFileSync(versionFilePath, JSON.stringify({ version: cleanNewVersion }, null, 2) + '\n', 'utf8')
console.log(`  ✓ Updated docs/version.json to ${cleanNewVersion}`)

// 5. Run sync-version to update README.md and examples
import('./sync-version.mjs')

console.log(`\n🎉 Release snapshot complete! You are now authoring documentation for ${cleanNewVersion}.`)
