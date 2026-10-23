<script setup lang="ts">
import LoginForm from "@/components/auth/LoginForm.vue";
</script>

<template>
  <main
    class="mx-auto flex min-h-screen max-w-md flex-col justify-center p-6 bg-gray-50"
  >
    <div class="bg-white p-8 rounded-2xl shadow-lg border border-gray-200">
      <div class="mb-6 text-center">
        <h1 class="text-2xl font-bold text-gray-900">Вход в систему</h1>
        <p class="text-sm text-gray-500 mt-1">
          Сервис «Карта пространства» ТТК
        </p>
      </div>
      <LoginForm />
      <p class="mt-6 text-center text-sm text-gray-500">
        Нет аккаунта?
        <router-link
          to="/register"
          class="text-red-600 font-semibold hover:underline"
          >Зарегистрироваться</router-link
        >
      </p>
    </div>
  </main>
</template>
