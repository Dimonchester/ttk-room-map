<script setup lang="ts">
import { Plus, MapPin, ExternalLink, Settings, Trash2 } from "lucide-vue-next";

const mockMaps = [
  {
    id: "d3b07384-d113-4ec2-a4f9-2603f29b465d",
    title: "Главный офис ТТК (Москва)",
    floors: 1,
    updatedAt: "Сегодня, 14:20",
    published: true,
    slug: "ttk-office-msk",
  },
  {
    id: "e4c88495-e224-5fd3-b500-3714f30c576e",
    title: "Выставочный стенд Expocentre",
    floors: 1,
    updatedAt: "Вчера, 10:15",
    published: false,
    slug: null,
  },
];
</script>

<template>
  <div class="space-y-6">
    <div
      class="flex justify-between items-center bg-white p-5 rounded-2xl shadow-xs border border-gray-200"
    >
      <div>
        <h2 class="text-xl font-bold text-gray-900">Ваши пространства</h2>
        <p class="text-sm text-gray-500">
          Управление планами зданий, точками интереса и панорамами
        </p>
      </div>
      <button
        class="flex items-center gap-2 bg-red-600 hover:bg-red-700 text-white px-4 py-2.5 rounded-xl font-medium text-sm transition-colors shadow-sm"
      >
        <Plus class="w-4 h-4" />
        <span>Создать карту</span>
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div
        v-for="map in mockMaps"
        :key="map.id"
        class="bg-white p-5 rounded-2xl border border-gray-200 shadow-xs hover:border-red-300 transition-all flex flex-col justify-between"
      >
        <div>
          <div class="flex justify-between items-start mb-3">
            <span
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold"
              :class="
                map.published
                  ? 'bg-green-100 text-green-800'
                  : 'bg-gray-100 text-gray-800'
              "
            >
              {{ map.published ? "Опубликована" : "Черновик" }}
            </span>
            <span class="text-xs text-gray-400"
              >Изменено: {{ map.updatedAt }}</span
            >
          </div>
          <h3
            class="text-lg font-semibold text-gray-900 mb-1 flex items-center gap-2"
          >
            <MapPin class="w-5 h-5 text-red-600 shrink-0" />
            <span class="truncate">{{ map.title }}</span>
          </h3>
          <p class="text-sm text-gray-500">
            Этажей: {{ map.floors }} | Объектов: 8
          </p>
        </div>

        <div
          class="mt-6 pt-4 border-t border-gray-100 flex items-center justify-between"
        >
          <router-link
            :to="{ name: 'editor', params: { id: map.id } }"
            class="flex items-center gap-1.5 text-sm font-medium text-gray-700 hover:text-red-600"
          >
            <Settings class="w-4 h-4" />
            <span>Редактор</span>
          </router-link>

          <div class="flex items-center gap-1">
            <a
              v-if="map.published && map.slug"
              :href="`/m/${map.slug}`"
              target="_blank"
              class="p-2 text-gray-500 hover:text-red-600 hover:bg-gray-50 rounded-xl"
            >
              <ExternalLink class="w-4 h-4" />
            </a>
            <button
              class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-xl"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
