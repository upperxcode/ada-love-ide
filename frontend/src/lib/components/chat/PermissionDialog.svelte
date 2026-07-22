<script lang="ts">
	import { Icon } from '$lib/components/icon';

	interface Props {
		open: boolean;
		toolName: string;
		args: string;
		reason: string;
		targetPath: string;
		mode: string;
		onDecision: (decision: 'allow_once' | 'allow_session' | 'deny') => void;
	}

	let {
		open = false,
		toolName = '',
		args = '',
		reason = '',
		targetPath = '',
		mode = '',
		onDecision = (_d: 'allow_once' | 'allow_session' | 'deny') => {}
	}: Props = $props();
</script>

{#if open}
	<div class="fixed inset-0 z-[200] flex items-center justify-center">
		<div class="fixed inset-0 bg-black/50" onclick={() => onDecision('deny')}></div>
		<div class="relative z-10 w-[360px] rounded-xl shadow-2xl overflow-hidden" style="background: var(--bg-secondary); border: 1px solid var(--border-primary)">
			<div class="px-5 py-4 flex items-center gap-3 border-b border-[var(--border-primary)]">
				<div class="flex items-center justify-center w-8 h-8 rounded-full" style="background: var(--accent-primary)/15">
					<Icon name="shield" size={16} style="color: var(--accent-primary)" />
				</div>
				<div>
					<p class="text-[13px] font-semibold" style="color: var(--text-primary)">Permissão necessária</p>
					<p class="text-[10px] uppercase tracking-wider opacity-50 font-sans">Modo {mode}</p>
				</div>
			</div>
			<div class="px-5 py-4 flex flex-col gap-3">
				<p class="text-[12px] leading-relaxed" style="color: var(--text-secondary)">
					O modelo quer <strong style="color: var(--text-primary)">{toolName}</strong>
					{#if targetPath}
						em <code class="text-[11px] px-1.5 py-0.5 rounded" style="background: var(--bg-tertiary); color: var(--text-primary)">{targetPath}</code>
					{/if}
				</p>
				{#if args}
					<pre class="m-0 text-[10px] p-2 rounded-lg overflow-x-auto whitespace-pre-wrap" style="background: var(--bg-tertiary); color: var(--text-faint); font-family: inherit; max-height: 100px">{args}</pre>
				{/if}
				<p class="text-[11px] leading-relaxed opacity-60" style="color: var(--text-faint)">{reason}</p>

				<div class="flex flex-col gap-2 mt-1">
					<button type="button" onclick={() => onDecision('allow_once')}
						class="w-full px-4 py-2.5 rounded-lg text-[12px] font-semibold transition-all cursor-pointer hover:brightness-110 active:scale-[0.98]"
						style="background: var(--accent-primary); color: var(--accent-primary-fg); border: none">
						Permitir uma vez
					</button>
					{#if mode === 'EDIT'}
						<button type="button" onclick={() => onDecision('allow_session')}
							class="w-full px-4 py-2.5 rounded-lg text-[12px] font-semibold transition-all cursor-pointer hover:opacity-80 active:scale-[0.98]"
							style="background: transparent; color: var(--text-primary); border: 1px solid var(--border-primary)">
							Sempre nesta sessão
						</button>
					{/if}
					<button type="button" onclick={() => onDecision('deny')}
						class="w-full px-4 py-2.5 rounded-lg text-[12px] font-semibold transition-all cursor-pointer hover:opacity-80 active:scale-[0.98]"
						style="background: transparent; color: var(--status-error); border: 1px solid var(--status-error)/30">
						Negar
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
