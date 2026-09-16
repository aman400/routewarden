import { defineConfig } from 'vitepress'
import versionData from '../version.json' with { type: 'json' }
import versionsRegistry from '../versions.json' with { type: 'json' }

export default defineConfig({
  title: 'RouteWarden',
  description: 'High-Performance Traefik Middleware for Sensitive Path Defense',
  base: '/routewarden/',
  cleanUrls: true,
  transformPageData(pageData) {
    // Provide version globally to markdown templates
    pageData.params = { ...pageData.params, version: versionData.version }
  },
  markdown: {
    config(md) {
      const originalRender = md.render.bind(md)
      md.render = (src, env) => {
        const replaced = src.replace(/\{\{version\}\}/g, versionData.version)
        return originalRender(replaced, env)
      }
    }
  },
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/routewarden/icon.svg' }],
    ['meta', { name: 'theme-color', content: '#6366f1' }]
  ],
  themeConfig: {
    logo: '/icon.svg',
    siteTitle: 'RouteWarden',
    nav: [
      {
        text: 'Guide',
        activeMatch: '^/guide/',
        items: [
          { text: 'Getting Started', link: '/guide/getting-started' },
          { text: 'System Architecture', link: '/guide/architecture' },
          { text: 'Local Development', link: '/guide/local-deployment' },
          { text: 'Testing & CI', link: '/guide/testing' }
        ]
      },
      {
        text: 'Reference',
        activeMatch: '^/reference/',
        items: [
          { text: 'Configuration Options', link: '/reference/configuration' },
          { text: 'Custom Path Patterns', link: '/reference/custom-paths' },
          { text: 'Anti-Evasion Security', link: '/reference/anti-evasion' },
          { text: 'Changelog & Migrations', link: '/reference/changelog' }
        ]
      },
      { text: 'Examples', link: '/examples/overview', activeMatch: '^/examples/' },
      {
        text: 'v0.2.x',
        activeMatch: '^/v0\\.',
        items: [
          ...versionsRegistry.versions.map(v => ({ text: v.text, link: v.link })),
          { text: 'Changelog & Breaking Changes', link: '/reference/changelog' },
          { text: 'Traefik Plugin Catalog', link: 'https://plugins.traefik.io' }
        ]
      }
    ],
    sidebar: {
      '/v0.1/': [
        {
          text: 'RouteWarden v0.1.x',
          collapsed: false,
          items: [
            { text: 'Overview & Setup (v0.1.x)', link: '/v0.1/guide/getting-started' },
            { text: 'Configuration (v0.1.x)', link: '/v0.1/reference/configuration' },
            { text: 'Switch to Latest (v0.2.x) ➔', link: '/guide/getting-started' }
          ]
        }
      ],
      '/': [
        {
          text: 'Getting Started',
          collapsed: false,
          items: [
            { text: 'Overview & Features', link: '/guide/getting-started' },
            { text: 'System Architecture', link: '/guide/architecture' },
            { text: 'Local Development & Deployment', link: '/guide/local-deployment' },
            { text: 'Testing & Verification', link: '/guide/testing' }
          ]
        },
        {
          text: 'Configuration & Security',
          collapsed: false,
          items: [
            { text: 'Configuration Reference', link: '/reference/configuration' },
            { text: 'Custom Path Patterns', link: '/reference/custom-paths' },
            { text: 'Anti-Evasion Security', link: '/reference/anti-evasion' },
            { text: 'Changelog & Migration', link: '/reference/changelog' }
          ]
        },
        {
          text: 'Examples Cookbook',
          collapsed: false,
          items: [
            { text: 'Examples Overview', link: '/examples/overview' },
            { text: '1. Basic Sensitive Files', link: '/examples/basic-sensitive-files' },
            { text: '2. Global EntryPoint Shield', link: '/examples/docker-compose-global' },
            { text: '3. Service-Level Docker Compose', link: '/examples/docker-compose-service' },
            { text: '4. IP / Subnet Whitelisting', link: '/examples/ip-whitelisting' },
            { text: '5. Captcha Challenge', link: '/examples/captcha' },
            { text: '6. Kubernetes IngressRoute', link: '/examples/kubernetes' }
          ]
        }
      ]
    },
    search: {
      provider: 'local'
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/aman400/routewarden' }
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2026 RouteWarden Contributors'
    }
  }
})
