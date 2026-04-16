/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import react from '@vitejs/plugin-react';
import { defineConfig, transformWithEsbuild } from 'vite';
import pkg from '@douyinfe/vite-plugin-semi';
import path from 'path';
import { codeInspectorPlugin } from 'code-inspector-plugin';
const { vitePluginSemi } = pkg;

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@douyinfe/semi-ui/dist/css/semi.css': path.resolve(
        __dirname,
        'node_modules/@douyinfe/semi-ui/dist/css/semi.css',
      ),
    },
  },
  plugins: [
    codeInspectorPlugin({
      bundler: 'vite',
    }),
    {
      name: 'treat-js-files-as-jsx',
      async transform(code, id) {
        if (!/src\/.*\.js$/.test(id)) {
          return null;
        }

        // Use the exposed transform from vite, instead of directly
        // transforming with esbuild
        return transformWithEsbuild(code, id, {
          loader: 'jsx',
          jsx: 'automatic',
        });
      },
    },
    react(),
    vitePluginSemi({
      cssLayer: true,
    }),
  ],
  optimizeDeps: {
    force: true,
    esbuildOptions: {
      loader: {
        '.js': 'jsx',
        '.json': 'json',
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-core': ['react', 'react-dom', 'react-router-dom'],
          'semi-ui': ['@douyinfe/semi-icons', '@douyinfe/semi-ui'],
          tools: ['axios', 'history', 'marked'],
          'react-components': [
            'react-dropzone',
            'react-fireworks',
            'react-telegram-login',
            'react-toastify',
            'react-turnstile',
          ],
          i18n: [
            'i18next',
            'react-i18next',
            'i18next-browser-languagedetector',
          ],
        },
      },
    },
  },
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        bypass: (req, res, proxyOptions) => {
          if (req.url === '/api/status') {
            res.end(JSON.stringify({
              success: true,
              data: {
                version: 'dev',
                system_name: 'SpaceHPC',
                logo: '/logo.png',
                demo_site_enabled: false,
                docs_link: 'https://docs.newapi.pro',
                setup: true,
                footer_html: '',
                server_address: 'http://localhost:3000',
                HeaderNavModules: JSON.stringify({ pricing: false }),
              }
            }));
            return true;
          }
          if (req.url === '/api/notice') {
            res.end(JSON.stringify({ success: true, data: '' }));
            return true;
          }
          if (req.url === '/api/home_page_content') {
            res.end(JSON.stringify({ success: true, data: '' }));
            return true;
          }
          if (req.url === '/api/user/self') {
            res.end(JSON.stringify({
              success: true,
              data: {
                id: 1,
                username: 'dev',
                role: 100,
                group: 'default',
                setting: JSON.stringify({ language: 'zh-CN' }),
                sidebar_modules: {},
              }
            }));
            return true;
          }
        },
      },
      '/mj': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/pg': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
});
