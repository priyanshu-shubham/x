import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'x',
  description: 'A simple command runner. Optionally supercharged with AI.',

  
  head: [
    ['meta', { name: 'theme-color', content: '#10B981' }],
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Reference', link: '/reference/exec-steps' },
      { text: 'Examples', link: '/examples/' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is x?', link: '/guide/' },
            { text: 'Quick Start', link: '/guide/quick-start' },
            { text: 'Installation', link: '/guide/installation' }
          ]
        },
        {
          text: 'Tutorials',
          items: [
            { text: 'Your First Command', link: '/guide/first-command' },
            { text: 'Building a Python Fixer', link: '/guide/python-fixer' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'Step Types',
          items: [
            { text: 'exec', link: '/reference/exec-steps' },
            { text: 'llm', link: '/reference/llm-steps' },
            { text: 'agentic', link: '/reference/agentic-steps' },
            { text: 'subcommand', link: '/reference/subcommand-steps' }
          ]
        },
        {
          text: 'Configuration',
          items: [
            { text: 'Variables', link: '/reference/variables' },
            { text: 'Config Files', link: '/reference/config-files' }
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Overview', link: '/examples/' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/priyanshu-shubham/x' }
    ],

    search: {
      provider: 'local'
    },

    footer: {
      message: 'Released under the MIT License.',
    }
  }
})
