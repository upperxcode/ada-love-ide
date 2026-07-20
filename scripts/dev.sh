#!/bin/bash
set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🚀 Dev Mode — Go + Wails + Svelte 5"

# ── 1. Instalar dependências do frontend (se necessário) ──
if [ ! -d "frontend/node_modules" ]; then
    echo "📦 Instalando dependências do frontend..."
    cd frontend && npm install && cd ..
fi

# ── 2. Wails dev com hot-reload para Go + SvelteKit ──
echo "⚡ Iniciando Wails Dev (hot-reload Go + Vite + Svelte)..."
wails dev -tags webkit2_41

echo "✅ Dev mode encerrado"
