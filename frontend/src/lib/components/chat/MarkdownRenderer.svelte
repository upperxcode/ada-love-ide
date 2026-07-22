<script lang="ts">
	import { marked, Renderer } from 'marked';
	import hljs from 'highlight.js';

	const renderer = new Renderer();
	renderer.code = ({ text, lang }) => {
		let highlighted = text;
		if (lang && hljs.getLanguage(lang)) {
			try { highlighted = hljs.highlight(text, { language: lang }).value; }
			catch { highlighted = text; }
		} else if (text) {
			try { highlighted = hljs.highlightAuto(text).value; }
			catch { highlighted = text; }
		}
		const langClass = lang ? ` class="hljs language-${lang}"` : ' class="hljs"';
		return `<pre><code${langClass}>${highlighted}</code></pre>`;
	};

	marked.setOptions({ renderer, breaks: true });

	interface Props { content: string; }
	let { content }: Props = $props();

	let rendered = $derived.by(() => {
		if (!content) return '';
		try { return marked.parse(content) as string; }
		catch { return content; }
	});
</script>

<div class="prose prose-sm max-w-none"
	style="
		--tw-prose-body: var(--text-primary);
		--tw-prose-headings: var(--text-primary);
		--tw-prose-links: var(--accent-primary);
		--tw-prose-bold: var(--text-primary);
		--tw-prose-code: var(--text-primary);
		--tw-prose-pre-bg: #0d1117;
		--tw-prose-pre-border: 1px solid var(--border-primary);
		--tw-prose-quotes: var(--text-secondary);
		--tw-prose-quote-borders: var(--accent-primary);
		--tw-prose-hr: var(--border-primary);
		--tw-prose-th-bg: var(--bg-secondary);
		--tw-prose-td-borders: var(--border-primary);
		--tw-prose-th-borders: var(--border-primary);
		color: var(--text-primary);
		font-size: 13px;
		line-height: 1.7;
	"
>
	{@html rendered}

	<style>
		:global(.prose :where(pre)) {
			background-color: #0d1117 !important;
			border: 1px solid var(--border-primary);
			border-radius: 8px;
			padding: 1em;
			overflow-x: auto;
			margin: 0.8em 0;
			font-size: 12px;
			line-height: 1.5;
		}
		:global(.prose :where(code)) {
			font-family: 'Geist Mono Variable', ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace !important;
			font-size: 0.9em;
			padding: 0.15em 0.3em;
			border-radius: 4px;
			background-color: var(--surface-input);
			font-weight: 400;
		}
		:global(.prose :where(pre code)) {
			background: none;
			padding: 0;
			font-size: 12px;
			font-weight: 400;
		}
		:global(.prose :where(h2)) {
			border-bottom: 1px solid var(--border-primary);
			padding-bottom: 0.3em;
			margin-top: 1.5em;
		}
		:global(.prose :where(h1, h2, h3, h4)) {
			font-weight: 600;
			letter-spacing: -0.01em;
		}
		:global(.prose :where(p)) {
			margin: 0.5em 0;
		}
		:global(.prose :where(ul, ol)) {
			padding-left: 1.5em;
			margin: 0.4em 0;
		}
		:global(.prose :where(li)) {
			margin: 0.15em 0;
		}
		:global(.prose :where(blockquote)) {
			font-style: normal;
			margin: 0.6em 0;
			padding-left: 1em;
		}
		:global(.prose :where(table)) {
			border-collapse: collapse;
			width: 100%;
			margin: 0.8em 0;
			font-size: 12px;
		}
		:global(.prose :where(th, td)) {
			border: 1px solid var(--border-primary);
			padding: 0.4em 0.6em;
			text-align: left;
		}
		:global(.prose :where(th)) {
			background-color: var(--bg-secondary);
			font-weight: 600;
		}
		:global(.prose :where(a)) {
			text-decoration: underline;
			text-underline-offset: 2px;
		}
		:global(.prose :where(a:hover)) {
			opacity: 0.8;
		}
		:global(.prose :where(img)) {
			border-radius: 6px;
			border: 1px solid var(--border-primary);
			margin: 0.8em 0;
		}
		:global(.prose :where(hr)) {
			margin: 1.2em 0;
		}
	</style>
</div>
