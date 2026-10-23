#!/usr/bin/env node
// Кросс-платформенный скрипт: создаёт .env из .env.example.
// Работает на Windows, Linux и macOS (нужен только Node.js).
//
// Использование (из корня проекта):
//   node scripts/init-env.js                  # backend/.env из backend/.env.example
//   node scripts/init-env.js --dir=.          # ./.env из ./.env.example
//   node scripts/init-env.js --force          # перезаписать существующий .env

const fs = require('fs');
const path = require('path');

let targetDir = path.resolve(process.cwd(), 'backend');
let force = false;

for (const arg of process.argv.slice(2)) {
  if (arg === '--force') force = true;
  else if (arg.startsWith('--dir=')) {
    targetDir = path.resolve(process.cwd(), arg.slice('--dir='.length));
  } else {
    console.error(`Неизвестный аргумент: ${arg}`);
    console.error('Доступно: --force, --dir=<путь>');
    process.exit(2);
  }
}

const examplePath = path.join(targetDir, '.env.example');
const envPath = path.join(targetDir, '.env');

if (!fs.existsSync(examplePath)) {
  console.error(`❌ Не найден файл: ${examplePath}`);
  console.error('   Создайте .env.example (он коммитится в git, без секретов).');
  process.exit(1);
}

if (fs.existsSync(envPath) && !force) {
  console.log(`⚠️  ${envPath} уже существует — не перезаписываю.`);
  console.log('   Используйте --force, если хотите перезаписать.');
  process.exit(0);
}

fs.copyFileSync(examplePath, envPath);
console.log(`✅ Создан ${envPath}`);
console.log('   Откройте его и заполните реальные значения (пароли БД и т.п.).');