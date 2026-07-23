<script lang="ts">
	import { Icon } from '$lib/components/icon';
	import { cn } from '$lib/utils';

	interface Props {
		open: boolean;
		toolName: string;
		args: string;
		reason: string;
		targetPath: string;
		mode: string;
		action?: string;
		riskLevel?: string;
		onDecision: (decision: 'allow_once' | 'allow_session' | 'deny') => void;
	}

	let {
		open = false,
		toolName = '',
		args = '',
		reason = '',
		targetPath = '',
		mode = '',
		action = '',
		riskLevel = '',
		onDecision = (_d: 'allow_once' | 'allow_session' | 'deny') => {}
	}: Props = $props();

	const riskColor: Record<string, string> = {
		'none': 'var(--text-faint)',
		'low': 'var(--status-success)',
		'medium': 'var(--status-warning)',
		'high': 'var(--status-error)',
		'critical': '#dc2626',
	};

	const riskLabel: Record<string, string> = {
		'none': 'Sem risco',
		'low': 'Baixo',
		'medium': 'Médio',
		'high': 'Alto',
		'critical': 'Crítico',
	};

	const riskIcon: Record<string, string> = {
		'none': 'info',
		'low': 'check',
		'medium': 'alert',
		'high': 'warning',
		'critical': 'alert-octagon',
	};

	function modeColor(m: string): string {
		switch (m) {
			case 'ASK': return 'var(--text-faint)';
			case 'EDIT': return 'var(--accent-primary)';
			case 'PLAN': return '#8b5cf6';
			case 'EXECUTE': return '#f97316';
			case 'FULL': return '#f97316';
			case 'ADMIN': return '#ef4444';
			default: return 'var(--text-faint)';
		}
	}

	const isDestructive = $derived(
		riskLevel === 'critical' || riskLevel === 'high' ||
		targetPath.includes('.env') || targetPath.includes('/etc/') ||
		toolName === 'exec' && (args.includes('rm -rf') || args.includes('git push --force'))
	);
</script>

{#if open}
	<div class="fixed inset-0 z-[200] flex items-center justify-center">
		<div class="fixed inset-0 bg-black/50" onclick={() => onDecision('deny')}></div>
		<div class="relative z-10 w-[380px] rounded-xl shadow-2xl overflow-hidden" style="background: var(--bg-secondary); border: 1px solid var(--border-primary)">
			<!-- Header -->
			<div class="px-5 py-4 flex items-center gap-3 border-b border-[var(--border-primary)]">
				<div class="flex items-center justify-center w-9 h-9 rounded-full" style="background: {riskColor[riskLevel] || 'var(--accent-primary)'}18">
					<Icon name={riskIcon[riskLevel] || 'shield'} size={16} style="color: {riskColor[riskLevel] || 'var(--accent-primary)'}" />
				</div>
				<div class="flex-1 min-w-0">
					<p class="text-[13px] font-semibold" style="color: var(--text-primary)">Permissão necessária</p>
					<div class="flex items-center gap-2 mt-0.5">
						<span class="text-[10px] uppercase tracking-wider font-semibold px-1.5 py-0.5 rounded" 
							style="background: {modeColor(mode)}15; color: {modeColor(mode)}">
							{mode}
						</span>
						{#if riskLevel}
							<span class="text-[10px] uppercase tracking-wider font-semibold px-1.5 py-0.5 rounded"
								style="background: {riskColor[riskLevel]}15; color: {riskColor[riskLevel]}">
								{riskLabel[riskLevel] || riskLevel}
							</span>
						{/if}
					</div>
				</div>
			</div>

			<!-- Body -->
			<div class="px-5 py-4 flex flex-col gap-3">
				<p class="text-[12px] leading-relaxed" style="color: var(--text-secondary)">
					O modelo quer usar <strong style="color: var(--text-primary)">{toolName}</strong>
					{#if targetPath}
						em <code class="text-[11px] px-1.5 py-0.5 rounded" style="background: var(--bg-tertiary); color: var(--text-primary)">{targetPath}</code>
					{/if}
				</p>

				{#if action}
					<div class="flex items-center gap-1.5">
						<span class="text-[10px] font-medium uppercase tracking-wider" style="color: var(--text-faint)">Ação:</span>
						<span class="text-[11px]" style="color: var(--text-secondary)">{action}</span>
					</div>
				{/if}

				{#if args}
					<pre class="m-0 text-[10px] p-2.5 rounded-lg overflow-x-auto whitespace-pre-wrap" 
						style="background: var(--bg-tertiary); color: var(--text-faint); font-family: inherit; max-height: 120px">{
						args.length > 200 ? args.slice(0, 200) + '...' : args
					}</pre>
				{/if}

				<p class="text-[11px] leading-relaxed opacity-60" style="color: var(--text-faint)">{reason}</p>

				{#if isDestructive}
					<div class="flex items-center gap-2 p-2 rounded-lg" 
						style="background: rgba(220, 38, 38, 0.08); border: 1px solid rgba(220, 38, 38, 0.2);">
						<Icon name="alert-octagon" size={14} style="color: #dc2626; flex-shrink: 0" />
						<p class="text-[11px] leading-relaxed" style="color: #dc2626">
							Esta ação pode causar danos irreversíveis ao sistema.
						</p>
					</div>
				{/if}

				<!-- Buttons -->
				<div class="flex flex-col gap-2 mt-1">
					<button type="button" onclick={() => onDecision('allow_once')}
						class="w-full px-4 py-2.5 rounded-lg text-[12px] font-semibold transition-all cursor-pointer active:scale-[0.98]"
						style={isDestructive
							? 'background: #dc2626; color: white; border: none;'
							: 'background: var(--accent-primary); color: var(--accent-primary-fg); border: none'}>
						{isDestructive ? 'Permitir mesmo assim' : 'Permitir uma vez'}
					</button>
					{#if mode.toUpperCase() === 'EDIT' || mode.toUpperCase() === 'EXECUTE'}
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