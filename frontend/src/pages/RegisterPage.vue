<script setup lang="ts">
import RegisterForm from "@/components/auth/RegisterForm.vue";
</script>

<template>
  <main
    class="mx-auto flex min-h-screen max-w-md flex-col justify-center p-6 bg-gray-50"
  >
    <div class="bg-white p-8 rounded-2xl shadow-lg border border-gray-200">
      <div class="mb-6 text-center">
        <h1 class="text-2xl font-bold text-gray-900">Регистрация</h1>
        <p class="text-sm text-gray-500 mt-1">
          Создайте учетную запись организатора
        </p>
      </div>
      <RegisterForm />
      <p class="mt-6 text-center text-sm text-gray-500">
        Уже есть аккаунт?
        <router-link
          to="/login"
          class="text-red-600 font-semibold hover:underline"
          >Войти</router-link
        >
      </p>
    </div>
  </main>
</template>
