import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import os from 'node:os'
import { syncVersion } from '../scripts/sync-version.mjs'
import { snapshotVersion } from '../scripts/snapshot-version.mjs'

/**
 * Creates a mock isolated workspace for testing scripts
 */
function createMockWorkspace() {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'routewarden-scripts-test-'))

  // Setup docs structure
  fs.mkdirSync(path.join(tempDir, 'docs/guide'), { recursive: true })
  fs.mkdirSync(path.join(tempDir, 'docs/reference'), { recursive: true })
  fs.mkdirSync(path.join(tempDir, 'docs/examples'), { recursive: true })

  fs.writeFileSync(
    path.join(tempDir, 'docs/version.json'),
    JSON.stringify({ version: 'v0.2.1' }, null, 2)
  )

  fs.writeFileSync(
    path.join(tempDir, 'docs/versions.json'),
    JSON.stringify(
      {
        current: 'v0.2.x',
        versions: [
          { text: 'v0.2.x (Latest)', link: '/guide/getting-started', tag: 'v0.2.x' },
          { text: 'v0.1.x', link: '/v0.1/guide/getting-started', tag: 'v0.1.x' }
        ]
      },
      null,
      2
    )
  )

  fs.writeFileSync(
    path.join(tempDir, 'docs/guide/getting-started.md'),
    '# Getting Started\n\nInstall version: {{version}}\n'
  )

  // Setup package.json
  fs.writeFileSync(
    path.join(tempDir, 'package.json'),
    JSON.stringify({ name: 'routewarden', version: '0.2.1' }, null, 2)
  )

  // Setup README.md
  fs.writeFileSync(
    path.join(tempDir, 'README.md'),
    '--experimental.plugins.routewarden.version=v0.2.1\nversion: v0.2.1\n'
  )

  // Setup sample example
  fs.mkdirSync(path.join(tempDir, 'examples/01-basic-sensitive-files'), { recursive: true })
  fs.writeFileSync(
    path.join(tempDir, 'examples/01-basic-sensitive-files/docker-compose.yml'),
    '--experimental.plugins.routewarden.version=v0.2.1\n'
  )

  return tempDir
}

test('syncVersion updates versions in files when target version changes', (t) => {
  const ws = createMockWorkspace()
  t.after(() => fs.rmSync(ws, { recursive: true, force: true }))

  // Change version in docs/version.json to v0.3.0
  fs.writeFileSync(
    path.join(ws, 'docs/version.json'),
    JSON.stringify({ version: 'v0.3.0' }, null, 2)
  )

  const result = syncVersion({ rootDir: ws })

  assert.equal(result.targetVersion, 'v0.3.0')
  assert.ok(result.updatedFiles.includes('README.md'))
  assert.ok(result.updatedFiles.includes('package.json'))
  assert.ok(result.updatedFiles.includes('examples/01-basic-sensitive-files/docker-compose.yml'))

  // Verify README
  const readme = fs.readFileSync(path.join(ws, 'README.md'), 'utf8')
  assert.match(readme, /--experimental\.plugins\.routewarden\.version=v0\.3\.0/)
  assert.match(readme, /version: v0\.3\.0/)

  // Verify package.json
  const pkg = JSON.parse(fs.readFileSync(path.join(ws, 'package.json'), 'utf8'))
  assert.equal(pkg.version, '0.3.0')

  // Running sync again without version changes should be idempotent
  const secondRun = syncVersion({ rootDir: ws })
  assert.equal(secondRun.updatedFiles.length, 0)
})

test('syncVersion throws error when docs/version.json is missing and no version passed', (t) => {
  const ws = createMockWorkspace()
  t.after(() => fs.rmSync(ws, { recursive: true, force: true }))

  fs.rmSync(path.join(ws, 'docs/version.json'))
  assert.throws(() => syncVersion({ rootDir: ws }), /docs\/version\.json not found/)
})

test('snapshotVersion archives previous minor version and updates registry and files', (t) => {
  const ws = createMockWorkspace()
  t.after(() => fs.rmSync(ws, { recursive: true, force: true }))

  const res = snapshotVersion('v0.3.0', { rootDir: ws })

  assert.equal(res.previousVersion, 'v0.2.1')
  assert.equal(res.newVersion, 'v0.3.0')
  assert.equal(res.minorSnapshotDirName, 'v0.2')
  assert.equal(res.skipped, false)

  // Verify snapshot dir was created
  const archivedDoc = path.join(ws, 'docs/v0.2/guide/getting-started.md')
  assert.ok(fs.existsSync(archivedDoc))
  const archivedContent = fs.readFileSync(archivedDoc, 'utf8')
  assert.match(archivedContent, /Legacy Version Notice/)
  assert.match(archivedContent, /Install version: v0\.2\.1/)

  // Verify docs/versions.json updated with v0.3.x series
  const registry = JSON.parse(fs.readFileSync(path.join(ws, 'docs/versions.json'), 'utf8'))
  assert.equal(registry.current, 'v0.3.x')
  assert.equal(registry.versions[0].tag, 'v0.3.x')
  assert.equal(registry.versions[0].text, 'v0.3.x (Latest)')
  assert.equal(registry.versions[1].tag, 'v0.2.x')
  assert.equal(registry.versions[1].link, '/v0.2/guide/getting-started')

  // Verify docs/version.json updated
  const versionData = JSON.parse(fs.readFileSync(path.join(ws, 'docs/version.json'), 'utf8'))
  assert.equal(versionData.version, 'v0.3.0')

  // Verify README and package.json were updated through sync
  const readme = fs.readFileSync(path.join(ws, 'README.md'), 'utf8')
  assert.match(readme, /--experimental\.plugins\.routewarden\.version=v0\.3\.0/)
  const pkg = JSON.parse(fs.readFileSync(path.join(ws, 'package.json'), 'utf8'))
  assert.equal(pkg.version, '0.3.0')
})

test('snapshotVersion skips when target version matches current version', (t) => {
  const ws = createMockWorkspace()
  t.after(() => fs.rmSync(ws, { recursive: true, force: true }))

  const res = snapshotVersion('v0.2.1', { rootDir: ws })
  assert.equal(res.skipped, true)
})

test('snapshotVersion throws when version argument is missing', (t) => {
  const ws = createMockWorkspace()
  t.after(() => fs.rmSync(ws, { recursive: true, force: true }))

  assert.throws(() => snapshotVersion('', { rootDir: ws }), /New version argument is required/)
})
