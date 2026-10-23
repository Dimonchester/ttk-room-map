<script setup lang="ts">
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import MapsHeader from "@/components/maps/MapsHeader.vue";
import MapsTable from "@/components/maps/MapsTable.vue";

const auth = useAuthStore();
const router = useRouter();

async function logout() {
  await auth.logout();
  await router.push({ name: "login" });
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <MapsHeader :email="auth.user?.email || 'admin@ttk.ru'" @logout="logout" />
    <main class="flex-1 max-w-5xl w-full mx-auto p-6">
      <MapsTable />
    </main>
  </div>
</template>
