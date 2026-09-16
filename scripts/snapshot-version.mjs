import fs from 'node:fs'
import path from 'node:path'
import { syncVersion } from './sync-version.mjs'

/**
 * Recursive copy function
 */
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

/**
 * Automates snapshotting of the documentation when releasing a new version.
 * @param {string} newVersion - Target release version (e.g., 'v0.3.0')
 * @param {Object} [options]
 * @param {string} [options.rootDir]
 * @returns {{ previousVersion: string, newVersion: string, minorSnapshotDirName: string }}
 */
export function snapshotVersion(newVersion, options = {}) {
  if (!newVersion) {
    throw new Error('New version argument is required')
  }

  const cleanNewVersion = newVersion.startsWith('v') ? newVersion : `v${newVersion}`
  const rootDir = options.rootDir || process.cwd()

  const versionFilePath = path.join(rootDir, 'docs/version.json')
  const versionsRegistryPath = path.join(rootDir, 'docs/versions.json')

  if (!fs.existsSync(versionFilePath) || !fs.existsSync(versionsRegistryPath)) {
    throw new Error(`Missing ${versionFilePath} or ${versionsRegistryPath}`)
  }

  const currentVersionData = JSON.parse(fs.readFileSync(versionFilePath, 'utf8'))
  const currentVersion = currentVersionData.version
  const registry = JSON.parse(fs.readFileSync(versionsRegistryPath, 'utf8'))

  if (currentVersion === cleanNewVersion) {
    return {
      previousVersion: currentVersion,
      newVersion: cleanNewVersion,
      minorSnapshotDirName: '',
      skipped: true
    }
  }

  // Major.Minor directory name (e.g. v0.2.1 -> v0.2)
  const versionParts = currentVersion.replace(/^v/, '').split('.')
  const minorSnapshotDirName = `v${versionParts[0]}.${versionParts[1]}`
  const targetSnapshotDir = path.join(rootDir, `docs/${minorSnapshotDirName}`)

  // 1. Snapshot current guide, reference, and examples
  const dirsToSnapshot = ['guide', 'reference', 'examples']
  for (const dir of dirsToSnapshot) {
    const srcDir = path.join(rootDir, `docs/${dir}`)
    const destDir = path.join(rootDir, `docs/${minorSnapshotDirName}/${dir}`)
    if (fs.existsSync(srcDir)) {
      copyDir(srcDir, destDir, (content) => {
        let replaced = content.replace(/\{\{version\}\}/g, currentVersion)
        if (!replaced.includes('Legacy Version Notice') && replaced.startsWith('# ')) {
          const firstLineEnd = replaced.indexOf('\n')
          const title = replaced.substring(0, firstLineEnd)
          const rest = replaced.substring(firstLineEnd)
          const banner = `\n\n::: warning Legacy Version Notice\nYou are viewing archived documentation for **${currentVersion}**. [Switch to Latest ➔](/guide/getting-started)\n:::`
          return `${title} (${currentVersion})${banner}${rest}`
        }
        return replaced
      })
    }
  }

  // 2. Update docs/versions.json
  const newSeries = `v${cleanNewVersion.replace(/^v/, '').split('.').slice(0, 2).join('.')}.x`
  const prevSeries = `v${versionParts[0]}.${versionParts[1]}.x`

  const updatedVersions = registry.versions.filter(
    (v) => v.tag !== newSeries && v.tag !== prevSeries
  )

  const newRegistryVersions = [
    {
      text: `${newSeries} (Latest)`,
      link: '/guide/getting-started',
      tag: newSeries
    },
    {
      text: `${prevSeries}`,
      link: `/${minorSnapshotDirName}/guide/getting-started`,
      tag: prevSeries
    },
    ...updatedVersions
  ]

  registry.current = newSeries
  registry.versions = newRegistryVersions
  fs.writeFileSync(versionsRegistryPath, JSON.stringify(registry, null, 2) + '\n', 'utf8')

  // 3. Update docs/version.json to new version
  fs.writeFileSync(versionFilePath, JSON.stringify({ version: cleanNewVersion }, null, 2) + '\n', 'utf8')

  // 4. Sync version across examples and README
  syncVersion({ rootDir, version: cleanNewVersion })

  return {
    previousVersion: currentVersion,
    newVersion: cleanNewVersion,
    minorSnapshotDirName,
    skipped: false
  }
}

// CLI execution
if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(new URL(import.meta.url).pathname)) {
  const versionArg = process.argv[2]
  if (!versionArg) {
    console.error('Usage: npm run docs:release <new-version>')
    console.error('Example: npm run docs:release v0.3.0')
    process.exit(1)
  }
  const result = snapshotVersion(versionArg)
  if (result.skipped) {
    console.log(`Version is already set to ${result.newVersion}.`)
  } else {
    console.log(`\n🎉 Release snapshot complete! You are now authoring documentation for ${result.newVersion}.`)
  }
}
