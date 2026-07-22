# Regras de Desenvolvimento

## SEM FALLBACK
Fallback é gambiarra. Fallback mascara erros.

Nunca use `|| 'default'`, `?? 'fallback'`, ou qualquer valor padrão que esconda a ausência de um dado real. Se um parâmetro obrigatório não foi passado, o erro deve explodir na hora — não ser engolido por um valor fictício.

**Certo:**
```typescript
async function initSession(workspacePath: string) {
    const sess = await CreateSessionWithConfig(workspacePath, ...);
}
```

**Errado:**
```typescript
const wsPath = workspacePath || 'default-workspace'; // NUNCA
```
