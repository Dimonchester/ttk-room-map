<script setup lang="ts">
import DevelopmentStub from "@/components/common/DevelopmentStub.vue";
defineProps<{ slug?: string }>();
</script>

<template>
  <DevelopmentStub
    title="Просмотр карты"
    :description="`Публичный просмотр для посетителя (Слуг: ${slug})`"
  />
</template>
