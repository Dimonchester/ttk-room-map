<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

async function logout() {
  await auth.logout()
  await router.push({ name: 'login' })
}
</script>

<template>
  <main class="mx-auto flex min-h-full max-w-md flex-col justify-center gap-3 p-6">
    <h1 class="text-3xl font-bold text-brand">Мои карты</h1>
    <p class="text-neutral-600">Заглушка, список карт появится на шаге 3.</p>
    <p class="text-sm">{{ auth.user?.email }}</p>
    <button class="rounded-lg bg-brand px-4 py-2 font-semibold text-white active:bg-brand-dark" @click="logout">
      Выйти
    </button>
  </main>
</template>
