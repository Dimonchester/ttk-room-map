<script setup lang="ts">
import DevelopmentStub from "@/components/common/DevelopmentStub.vue";
defineProps<{ slug?: string }>();
</script>

<template>
  <DevelopmentStub
    title="Встраиваемая карта (Embed)"
    :description="`Виджет для встраивания на сторонние сайты (Слуг: ${slug})`"
  />
</template>
