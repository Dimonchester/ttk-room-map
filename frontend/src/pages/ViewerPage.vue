<script setup lang="ts">
defineProps<{ slug: string }>()
</script>

<template>
  <main class="mx-auto flex min-h-full max-w-md flex-col justify-center gap-3 p-6">
    <h1 class="text-3xl font-bold text-brand">Карта</h1>
    <p class="text-neutral-600">Заглушка просмотрщика. Ссылка: {{ slug }}</p>
  </main>
</template>
