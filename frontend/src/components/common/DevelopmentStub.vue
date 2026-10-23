<script setup lang="ts">
import { Construction } from "lucide-vue-next";

defineProps<{
  title: string;
  description?: string;
}>();
</script>

<template>
  <main
    class="mx-auto flex min-h-screen max-w-lg flex-col items-center justify-center p-6 text-center"
  >
    <div
      class="bg-white border border-gray-200 shadow-sm rounded-2xl p-8 w-full flex flex-col items-center"
    >
      <div
        class="w-12 h-12 bg-amber-50 text-amber-600 rounded-xl flex items-center justify-center mb-4"
      >
        <Construction class="w-6 h-6" />
      </div>
      <h1 class="text-2xl font-bold text-gray-900 mb-2">{{ title }}</h1>
      <p class="text-gray-500 text-sm mb-6">
        {{
          description ||
          "Ещё в разработке. Этот функционал появится в следующих шагах."
        }}
      </p>
      <router-link
        to="/"
        class="bg-gray-900 hover:bg-black text-white px-5 py-2.5 rounded-xl text-sm font-medium transition-colors"
      >
        Вернуться к списку карт
      </router-link>
    </div>
  </main>
</template>
