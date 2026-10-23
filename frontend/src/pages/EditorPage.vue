<script setup lang="ts">
defineProps<{ id: string }>()
</script>

<template>
  <main class="mx-auto flex min-h-full max-w-md flex-col justify-center gap-3 p-6">
    <h1 class="text-3xl font-bold text-brand">Редактор</h1>
    <p class="text-neutral-600">Заглушка. Карта: {{ id }}</p>
  </main>
</template>
