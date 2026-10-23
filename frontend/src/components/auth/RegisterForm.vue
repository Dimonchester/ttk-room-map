<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { Mail, Lock, UserPlus } from "lucide-vue-next";

const auth = useAuthStore();
const router = useRouter();

const email = ref("");
const password = ref("");
const loading = ref(false);

async function handleRegister() {
  console.log(
    "Нажата кнопка: Зарегистрироваться (отправка POST /api/auth/register)",
  );
  loading.value = true;
  try {
    await auth.register(email.value, password.value);
    console.log("Регистрация прошла успешно");
    await router.push("/");
  } catch (e) {
    console.warn("Ошибка регистрации (возможно, email уже занят - статус 409)");
    alert(
      "Ошибка регистрации. Возможно, пользователь с таким email уже существует.",
    );
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <form @submit.prevent="handleRegister" class="space-y-4">
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
          placeholder="org@ttk.ru"
          class="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-xl focus:ring-2 focus:ring-red-600 focus:outline-none text-sm"
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
          placeholder="Минимум 8 символов"
          class="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-xl focus:ring-2 focus:ring-red-600 focus:outline-none text-sm"
        />
      </div>
    </div>
    <button
      type="submit"
      :disabled="loading"
      class="w-full flex items-center justify-center gap-2 bg-gray-900 hover:bg-black text-white font-medium py-2.5 rounded-xl transition-colors text-sm disabled:opacity-50"
    >
      <UserPlus class="w-4 h-4" />
      <span>{{ loading ? "Создание..." : "Зарегистрироваться" }}</span>
    </button>
  </form>
</template>
