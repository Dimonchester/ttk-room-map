<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { setMockSession } from "@/api/session";
import { Mail, Lock, ArrowRight, FlaskConical } from "lucide-vue-next";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const email = ref("admin@ttk.ru");
const password = ref("");
const loading = ref(false);

async function handleLogin() {
  loading.value = true;
  try {
    await auth.login(email.value, password.value);
    const redirectPath = (route.query.redirect as string) || "/maps";
    await router.push(redirectPath);
  } catch {
    alert("Бэкенд еще не запущен.");
  } finally {
    loading.value = false;
  }
}

// Тестовый вход: проставляем данные в session.ts и переходим в ЛК
async function handleTestLogin() {
  setMockSession();
  const redirectPath = (route.query.redirect as string) || "/maps";
  await router.push(redirectPath);
}
</script>

<template>
  <form @submit.prevent="handleLogin" class="space-y-4">
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
      <div class="relative">
        <span
          class="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400"
        >
          <Mail class="w-5 h-5" />
        </span>
        <input
          v-model="email"
          type="email"
          required
          placeholder="admin@ttk.ru"
          class="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-xl focus:ring-2 focus:ring-red-600 focus:outline-hidden text-sm"
        />
      </div>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
      <div class="relative">
        <span
          class="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400"
        >
          <Lock class="w-5 h-5" />
        </span>
        <input
          v-model="password"
          type="password"
          required
          placeholder="••••••••"
          class="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-xl focus:ring-2 focus:ring-red-600 focus:outline-hidden text-sm"
        />
      </div>
    </div>

    <button
      type="submit"
      :disabled="loading"
      class="w-full flex items-center justify-center gap-2 bg-red-600 hover:bg-red-700 text-white font-medium py-2.5 rounded-xl transition-colors shadow-xs text-sm disabled:opacity-50 cursor-pointer"
    >
      <span>{{ loading ? "Вход..." : "Войти в систему" }}</span>
      <ArrowRight class="w-4 h-4" />
    </button>

    <div class="pt-2 border-t border-gray-100 mt-4">
      <button
        type="button"
        @click="handleTestLogin"
        class="w-full flex items-center justify-center gap-2 bg-amber-50 hover:bg-amber-100 text-amber-800 font-medium py-2.5 rounded-xl transition-colors text-sm border border-amber-200 cursor-pointer"
      >
        <FlaskConical class="w-4 h-4 text-amber-600" />
        <span>Тестовый вход</span>
      </button>
    </div>
  </form>
</template>
