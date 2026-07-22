<script lang="ts">
	import { cn } from '$lib/utils';
	import MarkdownRenderer from './MarkdownRenderer.svelte';
	import Reasoning from './Reasoning.svelte';
	import ReasoningTrigger from './ReasoningTrigger.svelte';
	import ReasoningContent from './ReasoningContent.svelte';
	import ActionItem from './ActionItem.svelte';
	import AnswerIcon from '../icons/AnswerIcon.svelte';
	import QuestionIcon from '../icons/QuestionIcon.svelte';
	import Icon from '../icon/Icon.svelte';
	import type { ChatMessage, ActionLog } from './chat-types';

	interface Props {
		message: ChatMessage;
		isStreaming?: boolean;
		isLastAssistant?: boolean;
		responseTime?: string;
		onCopy?: (text: string) => void;
		onToggleAction?: (id: string) => void;
		actionLogs?: ActionLog[];
	}

	let { message, isStreaming = false, isLastAssistant = false, responseTime = '', onCopy, onToggleAction, actionLogs = [], ...restProps }: Props = $props();

	let copied = $state(false);
	let copyTimer: ReturnType<typeof setTimeout> | null = null;

	async function handleCopy() {
		if (!onCopy || !message.content) return;
		copied = true;
		onCopy(message.content);
		if (copyTimer) clearTimeout(copyTimer);
		copyTimer = setTimeout(() => (copied = false), 2000);
	}

	let showReasoning = $derived(
		isStreaming || !!message.thinkingContent
	);

	let reasoningDuration = $derived(message.thinkingDuration ?? 0);
	let combinedActions = $derived(
		actionLogs.length > 0 ? actionLogs : (message.actions ?? [])
	);
</script>

{#if message.role === 'user'}
	<!-- User message -->
	<div
		class="animate-in fade-in slide-in-from-bottom-2 duration-300 flex items-end justify-end gap-2.5"
		{...restProps}
	>
		<div
			class="max-w-[75%] rounded-2xl rounded-br-md px-4 py-2.5 text-sm leading-relaxed shadow-md shadow-black/20"
			style="background-color: var(--accent-primary); color: var(--accent-primary-fg)"
		>
			<p class="whitespace-pre-wrap">{message.content}</p>
		</div>
		<div class="flex shrink-0 items-center justify-center size-6" style="color: var(--accent-primary)">
			<QuestionIcon size={24} />
		</div>
	</div>
{:else}
	<!-- Assistant message -->
	<div
		class="animate-in fade-in slide-in-from-bottom-2 duration-300 group flex items-start gap-3"
		{...restProps}
	>
		<!-- Avatar -->
		<div class="flex shrink-0 items-center justify-center size-6">
			<AnswerIcon size={24} style="color: var(--text-secondary)" />
		</div>

		<!-- Content area -->
		<div class="flex-1 min-w-0">
			<div
				class="rounded-xl px-4 py-3 flex flex-col gap-1.5"
				style="background-color: color-mix(in srgb, var(--bg-secondary) 30%, transparent); border: 1px solid var(--border-subtle)"
			>
				{#if showReasoning}
					<Reasoning
						isStreaming={isStreaming}
						{reasoningDuration}
						defaultOpen={isStreaming}
					>
						<ReasoningTrigger />
						<ReasoningContent content={message.thinkingContent || ''} sections={message.thinkingSections || []}>
							{#if combinedActions.length > 0}
								<div class="flex flex-wrap items-center gap-1.5 mt-3 pt-3 border-t border-[var(--border-subtle)]">
									{#each combinedActions as log (log.id)}
										<ActionItem {log} onToggle={onToggleAction} />
									{/each}
								</div>
							{/if}
						</ReasoningContent>
					</Reasoning>
				{/if}

				{#if message.content}
					<MarkdownRenderer content={message.content} />
				{/if}
			</div>

			<!-- Footer -->
			{#if isLastAssistant && message.content && !isStreaming}
				<div
					class="flex items-center gap-2 mt-1.5 transition-opacity duration-200 {isLastAssistant ? 'opacity-0 group-hover:opacity-100' : ''}"
				>
					{#if responseTime}
						<span
							class="text-[10px] font-medium rounded-md px-1.5 py-0.5"
							style="background-color: var(--bg-tertiary); color: var(--text-faint)"
						>
							{responseTime}
						</span>
					{/if}
					<button
						type="button"
						onclick={handleCopy}
						class="flex items-center gap-1 text-[10px] cursor-pointer transition-colors hover:opacity-80"
						style="color: var(--text-faint)"
					>
						{#if copied}
							<Icon name="check" size={12} />
							<span>Copied!</span>
						{:else}
							<Icon name="copy" size={12} />
							<span>Copy</span>
						{/if}
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}