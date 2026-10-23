<script setup lang="ts">
import DevelopmentStub from "@/components/common/DevelopmentStub.vue";
defineProps<{ id?: string }>();
</script>

<template>
  <DevelopmentStub
    title="Редактор карты"
    :description="`Редактирование пространства (ID: ${id})`"
  />
</template>
