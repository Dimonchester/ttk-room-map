import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import { router } from './router'
import { ApiError } from './api/errors'
import './styles/main.css'

const app = createApp(App)

app.use(createPinia()) // pinia должен быть подключён до router: guard использует store
app.use(VueQueryPlugin, {
  queryClientConfig: {
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        refetchOnWindowFocus: false,
        // 4xx — ошибка запроса, повторять бессмысленно; сеть и 5xx — повторяем
        retry: (failureCount, error) =>
          !(error instanceof ApiError && error.status >= 400 && error.status < 500) && failureCount < 2,
      },
    },
  },
})
app.use(router)

app.mount('#app')
