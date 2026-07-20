#!/bin/bash
set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🚀 Build — Go + Wails + Svelte 5"

# ── 1. Frontend: instalar dependências e build ──
echo "📦 Instalando dependências do frontend..."
cd frontend
npm install
echo "🏗️  Build do frontend SvelteKit..."
npm run build
cd ..

# ── 2. Copiar output para diretório que o Wails embeda (se necessário) ──
# Wails v2 + SvelteKit: o Wails serve o frontend via Vite dev server em dev,
# e precisa do build output para produção. Ajuste o embedFS conforme sua config.
echo "📋 Preparando assets para Wails..."

# ── 3. Build Wails ──
echo "🏗️  Compilando Wails..."
wails build -tags webkit2_41

echo "✅ Build concluído! Binário em ./bin/ada-love-ide"
