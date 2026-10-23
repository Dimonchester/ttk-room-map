<script setup lang="ts">
import { LogOut, Map } from "lucide-vue-next";

defineProps<{ email?: string }>();
defineEmits<{ (e: "logout"): void }>();
</script>

<template>
  <header
    class="bg-white border-b border-gray-200 px-6 py-4 flex justify-between items-center shadow-xs"
  >
    <div class="flex items-center gap-3">
      <div class="bg-red-600 p-2 rounded-xl text-white">
        <Map class="w-5 h-5" />
      </div>
      <div>
        <h1 class="font-bold text-gray-900 leading-tight">
          Карта пространства
        </h1>
        <p class="text-xs text-gray-500">ТТК. ТрансТелеком</p>
      </div>
    </div>
    <div class="flex items-center gap-4">
      <span class="text-sm text-gray-600 hidden sm:inline font-medium">{{
        email
      }}</span>
      <button
        @click="$emit('logout')"
        class="flex items-center gap-1.5 text-sm text-gray-700 hover:text-red-600 bg-gray-100 hover:bg-red-50 px-3 py-2 rounded-xl transition-colors font-medium"
      >
        <LogOut class="w-4 h-4" />
        <span>Выйти</span>
      </button>
    </div>
  </header>
</template>
