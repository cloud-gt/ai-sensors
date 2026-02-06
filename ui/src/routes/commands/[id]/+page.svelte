<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { page } from '$app/state';
	import { base } from '$app/paths';
	import * as api from '$lib/api';
	import type { Command } from '$lib/types';

	let command = $state<Command | null>(null);
	let status = $state('not_started');
	let output = $state<string[]>([]);
	let error = $state('');
	let loading = $state(true);
	let autoScroll = $state(true);
	let pollInterval: ReturnType<typeof setInterval>;
	let terminalEl: HTMLElement;

	const id = $derived(page.params.id);

	async function load() {
		try {
			command = await api.getCommand(id);
			try {
				status = await api.getStatus(id);
				output = await api.getOutput(id, 500);
			} catch {
				status = 'not_started';
				output = [];
			}
			error = '';
			if (autoScroll) {
				await tick();
				scrollToBottom();
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load command';
		} finally {
			loading = false;
		}
	}

	function scrollToBottom() {
		if (terminalEl) {
			terminalEl.scrollTop = terminalEl.scrollHeight;
		}
	}

	function handleScroll() {
		if (!terminalEl) return;
		const { scrollTop, scrollHeight, clientHeight } = terminalEl;
		autoScroll = scrollHeight - scrollTop - clientHeight < 40;
	}

	async function handleStart() {
		try {
			await api.startCommand(id);
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to start';
		}
	}

	async function handleStop() {
		try {
			await api.stopCommand(id);
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to stop';
		}
	}

	function statusColor(s: string) {
		switch (s) {
			case 'running': return 'bg-signal-run-bg text-signal-run border-signal-run/20';
			case 'stopped': return 'bg-signal-stop-bg text-signal-stop border-signal-stop/20';
			default: return 'bg-signal-idle-bg text-signal-idle border-signal-idle/20';
		}
	}

	function statusLabel(s: string) {
		switch (s) {
			case 'running': return 'RUNNING';
			case 'stopped': return 'STOPPED';
			default: return 'IDLE';
		}
	}

	onMount(() => {
		load();
		pollInterval = setInterval(load, 2000);
	});

	onDestroy(() => {
		if (pollInterval) clearInterval(pollInterval);
	});
</script>

<div class="space-y-5 animate-fade-in">
	<!-- Back link -->
	<a href="{base}/" class="inline-flex items-center gap-1.5 font-mono text-xs text-text-muted hover:text-amber-glow transition-colors no-underline">
		<svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
			<path d="M10 4L6 8l4 4" />
		</svg>
		commands
	</a>

	{#if error}
		<div class="bg-signal-stop-bg border border-signal-stop/20 rounded-lg px-4 py-3 font-mono text-sm text-signal-stop flex items-center gap-2">
			<svg class="w-4 h-4 shrink-0" viewBox="0 0 16 16" fill="currentColor">
				<path d="M8 1a7 7 0 100 14A7 7 0 008 1zm-.75 4.75a.75.75 0 011.5 0v3a.75.75 0 01-1.5 0v-3zM8 11a1 1 0 100 2 1 1 0 000-2z"/>
			</svg>
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="text-center py-16">
			<div class="inline-block w-5 h-5 border-2 border-surface-4 border-t-amber-glow rounded-full" style="animation: spin 0.8s linear infinite;"></div>
		</div>
	{:else if command}
		<!-- Command header -->
		<div class="bg-surface-1 border border-border rounded-lg p-5">
			<div class="flex items-start justify-between gap-4">
				<div class="min-w-0">
					<div class="flex items-center gap-3 mb-2">
						<h1 class="text-lg font-semibold text-text-primary truncate">{command.name}</h1>
						<span class="font-mono text-[10px] uppercase tracking-widest px-2 py-0.5 rounded border shrink-0 {statusColor(status)}">
							{statusLabel(status)}
						</span>
					</div>

					<div class="space-y-1.5">
						<div class="flex items-center gap-2">
							<span class="font-mono text-[10px] uppercase text-text-muted tracking-wider w-12 shrink-0">cmd</span>
							<code class="font-mono text-sm text-text-secondary">{command.command}</code>
						</div>
						<div class="flex items-center gap-2">
							<span class="font-mono text-[10px] uppercase text-text-muted tracking-wider w-12 shrink-0">dir</span>
							<code class="font-mono text-sm text-text-secondary">{command.work_dir}</code>
						</div>
						<div class="flex items-center gap-2">
							<span class="font-mono text-[10px] uppercase text-text-muted tracking-wider w-12 shrink-0">id</span>
							<code class="font-mono text-xs text-text-muted">{command.id}</code>
						</div>
					</div>
				</div>

				<!-- Controls -->
				<div class="flex items-center gap-2 shrink-0">
					{#if status === 'running'}
						<button
							onclick={handleStop}
							class="h-9 px-4 rounded-md bg-signal-stop-bg border border-signal-stop/20 font-mono text-sm text-signal-stop hover:bg-signal-stop/20 transition-colors"
						>
							Stop
						</button>
					{:else}
						<button
							onclick={handleStart}
							class="h-9 px-4 rounded-md bg-signal-run-bg border border-signal-run/20 font-mono text-sm text-signal-run hover:bg-signal-run/20 transition-colors"
						>
							Start
						</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Terminal output -->
		<div class="bg-surface-1 border border-border rounded-lg overflow-hidden">
			<!-- Terminal header -->
			<div class="flex items-center justify-between px-4 py-2.5 border-b border-border-subtle bg-surface-2/50">
				<div class="flex items-center gap-2">
					<div class="flex gap-1.5">
						<div class="w-2.5 h-2.5 rounded-full bg-signal-stop/60"></div>
						<div class="w-2.5 h-2.5 rounded-full bg-amber-dim/60"></div>
						<div class="w-2.5 h-2.5 rounded-full bg-signal-run/60"></div>
					</div>
					<span class="font-mono text-[10px] text-text-muted uppercase tracking-wider ml-2">output</span>
				</div>
				<div class="flex items-center gap-3">
					{#if status === 'running'}
						<span class="font-mono text-[10px] text-signal-run flex items-center gap-1.5">
							<div class="w-1.5 h-1.5 rounded-full bg-signal-run" style="animation: pulse-dot 1.5s ease-in-out infinite;"></div>
							live
						</span>
					{/if}
					<span class="font-mono text-[10px] text-text-muted">{output.length} lines</span>
				</div>
			</div>

			<!-- Terminal body -->
			<div
				bind:this={terminalEl}
				onscroll={handleScroll}
				class="p-4 h-[500px] overflow-y-auto relative"
				style="background: #08080a;"
			>
				<!-- Subtle scanline overlay -->
				<div class="pointer-events-none absolute inset-0 opacity-[0.03]" style="background: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(255,255,255,0.03) 2px, rgba(255,255,255,0.03) 4px);"></div>

				{#if output.length > 0}
					<div class="relative z-10">
						{#each output as line, i}
							<div class="flex gap-3 hover:bg-white/[0.02] -mx-1 px-1 rounded">
								<span class="font-mono text-[10px] text-text-muted/40 select-none w-8 text-right shrink-0 leading-5">{i + 1}</span>
								<pre class="font-mono text-[13px] text-amber-bright/90 whitespace-pre-wrap break-all leading-5 min-h-5">{line || ' '}</pre>
							</div>
						{/each}
					</div>
				{:else}
					<div class="flex items-center justify-center h-full">
						<p class="font-mono text-sm text-text-muted/40">
							{status === 'not_started' ? 'Start the command to see output' : 'Waiting for output...'}
						</p>
					</div>
				{/if}

				<!-- Auto-scroll indicator -->
				{#if !autoScroll && output.length > 0}
					<button
						onclick={() => { autoScroll = true; scrollToBottom(); }}
						class="fixed bottom-12 right-12 h-8 px-3 rounded-full bg-surface-3 border border-border font-mono text-xs text-text-secondary hover:text-text-primary hover:border-amber-dim/30 transition-colors shadow-lg z-20"
					>
						Scroll to bottom
					</button>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
