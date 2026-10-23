import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import basicSsl from '@vitejs/plugin-basic-ssl'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    // basicSsl: камера и датчики ориентации на телефоне работают только по HTTPS
    plugins: [vue(), tailwindcss(), basicSsl()],
    resolve: {
      alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
    },
    server: {
      host: true, // доступ с телефона по IP в локальной сети
      proxy: {
        // same-origin: фронт и API на одном origin, CORS не нужен, cookie работают как в проде
        '/api': { target: env.API_TARGET || 'http://localhost:8080' },
      },
    },
  }
})
