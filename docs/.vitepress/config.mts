import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'RouteWarden',
  description: 'High-Performance Traefik Middleware for Sensitive Path Defense',
  base: '/routewarden/',
  cleanUrls: true,
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/routewarden/icon.svg' }],
    ['meta', { name: 'theme-color', content: '#6366f1' }]
  ],
  themeConfig: {
    logo: '/icon.svg',
    siteTitle: 'RouteWarden',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Architecture', link: '/guide/architecture' },
      { text: 'Local Dev', link: '/guide/local-deployment' },
      { text: 'Testing', link: '/guide/testing' },
      { text: 'Configuration', link: '/reference/configuration' },
      { text: 'Anti-Evasion', link: '/reference/anti-evasion' },
      { text: 'Examples & Wiki', link: '/examples/overview' },
      {
        text: 'v0.2.0',
        items: [
          { text: 'Changelog', link: 'https://github.com/aman400/routewarden/blob/main/CHANGELOG.md' },
          { text: 'Traefik Plugin Catalog', link: 'https://plugins.traefik.io' }
        ]
      }
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Overview & Features', link: '/guide/getting-started' },
            { text: 'System Architecture', link: '/guide/architecture' },
            { text: 'Local Development & Deployment', link: '/guide/local-deployment' },
            { text: 'Testing & Verification', link: '/guide/testing' }
          ]
        },
        {
          text: 'Configuration & Security',
          items: [
            { text: 'Configuration Reference', link: '/reference/configuration' },
            { text: 'Anti-Evasion Security', link: '/reference/anti-evasion' }
          ]
        },
        {
          text: 'Examples Cookbook',
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
      ],
      '/reference/': [
        {
          text: 'Configuration & Security',
          items: [
            { text: 'Configuration Reference', link: '/reference/configuration' },
            { text: 'Anti-Evasion Security', link: '/reference/anti-evasion' }
          ]
        },
        {
          text: 'Back to Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Examples Overview', link: '/examples/overview' }
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Cookbook & Scenarios',
          items: [
            { text: 'All Examples', link: '/examples/overview' },
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
