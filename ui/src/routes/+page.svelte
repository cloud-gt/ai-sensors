<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { base } from '$app/paths';
	import * as api from '$lib/api';
	import type { Command } from '$lib/types';

	let commands = $state<Command[]>([]);
	let statuses = $state<Record<string, string>>({});
	let error = $state('');
	let showForm = $state(false);
	let loading = $state(true);
	let pollInterval: ReturnType<typeof setInterval>;

	// Create form
	let newName = $state('');
	let newCommand = $state('');
	let newWorkDir = $state('');
	let creating = $state(false);

	async function loadCommands() {
		try {
			commands = await api.listCommands();
			const newStatuses: Record<string, string> = {};
			for (const cmd of commands) {
				try {
					newStatuses[cmd.id] = await api.getStatus(cmd.id);
				} catch {
					newStatuses[cmd.id] = 'not_started';
				}
			}
			statuses = newStatuses;
			error = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load commands';
		} finally {
			loading = false;
		}
	}

	async function handleCreate() {
		if (!newName.trim() || !newCommand.trim() || !newWorkDir.trim()) return;
		creating = true;
		error = '';
		try {
			await api.createCommand(newName.trim(), newCommand.trim(), newWorkDir.trim());
			newName = '';
			newCommand = '';
			newWorkDir = '';
			showForm = false;
			await loadCommands();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create command';
		} finally {
			creating = false;
		}
	}

	async function handleStart(id: string) {
		try {
			await api.startCommand(id);
			await loadCommands();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to start command';
		}
	}

	async function handleStop(id: string) {
		try {
			await api.stopCommand(id);
			await loadCommands();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to stop command';
		}
	}

	async function handleDelete(id: string) {
		try {
			await api.deleteCommand(id);
			await loadCommands();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete command';
		}
	}

	function statusColor(status: string) {
		switch (status) {
			case 'running': return 'bg-signal-run-bg text-signal-run border-signal-run/20';
			case 'stopped': return 'bg-signal-stop-bg text-signal-stop border-signal-stop/20';
			default: return 'bg-signal-idle-bg text-signal-idle border-signal-idle/20';
		}
	}

	function statusLabel(status: string) {
		switch (status) {
			case 'running': return 'RUN';
			case 'stopped': return 'STOP';
			default: return 'IDLE';
		}
	}

	onMount(() => {
		loadCommands();
		pollInterval = setInterval(loadCommands, 3000);
	});

	onDestroy(() => {
		if (pollInterval) clearInterval(pollInterval);
	});
</script>

<div class="space-y-6">
	<!-- Header row -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold text-text-primary">Commands</h1>
			<p class="text-sm text-text-muted mt-0.5 font-mono">
				{commands.length} registered{#if Object.values(statuses).filter(s => s === 'running').length > 0}
					&middot; <span class="text-signal-run">{Object.values(statuses).filter(s => s === 'running').length} running</span>
				{/if}
			</p>
		</div>
		<button
			onclick={() => showForm = !showForm}
			class="h-9 px-4 rounded-md font-mono text-sm font-medium transition-all
				{showForm
					? 'bg-surface-3 text-text-secondary hover:bg-surface-4'
					: 'bg-amber-glow text-surface-0 hover:bg-amber-bright'}"
		>
			{showForm ? 'Cancel' : '+ New command'}
		</button>
	</div>

	<!-- Error banner -->
	{#if error}
		<div class="bg-signal-stop-bg border border-signal-stop/20 rounded-lg px-4 py-3 font-mono text-sm text-signal-stop flex items-center gap-2 animate-fade-in">
			<svg class="w-4 h-4 shrink-0" viewBox="0 0 16 16" fill="currentColor">
				<path d="M8 1a7 7 0 100 14A7 7 0 008 1zm-.75 4.75a.75.75 0 011.5 0v3a.75.75 0 01-1.5 0v-3zM8 11a1 1 0 100 2 1 1 0 000-2z"/>
			</svg>
			{error}
		</div>
	{/if}

	<!-- Create form -->
	{#if showForm}
		<div class="bg-surface-1 border border-border rounded-lg p-5 animate-fade-in">
			<div class="flex items-center gap-2 mb-4">
				<div class="w-1.5 h-4 bg-amber-glow rounded-full"></div>
				<h2 class="font-mono text-sm font-semibold text-text-primary">New command</h2>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); handleCreate(); }} class="space-y-4">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-3">
					<div>
						<label for="name" class="block font-mono text-xs text-text-muted mb-1.5 uppercase tracking-wider">Name</label>
						<input
							id="name"
							bind:value={newName}
							placeholder="my-watcher"
							required
							class="w-full bg-surface-2 border border-border rounded-md px-3 py-2 font-mono text-sm text-text-primary placeholder:text-text-muted/50 focus:outline-none focus:border-amber-dim focus:ring-1 focus:ring-amber-dim/30 transition-colors"
						/>
					</div>
					<div>
						<label for="command" class="block font-mono text-xs text-text-muted mb-1.5 uppercase tracking-wider">Command</label>
						<input
							id="command"
							bind:value={newCommand}
							placeholder="go test ./..."
							required
							class="w-full bg-surface-2 border border-border rounded-md px-3 py-2 font-mono text-sm text-text-primary placeholder:text-text-muted/50 focus:outline-none focus:border-amber-dim focus:ring-1 focus:ring-amber-dim/30 transition-colors"
						/>
					</div>
					<div>
						<label for="workdir" class="block font-mono text-xs text-text-muted mb-1.5 uppercase tracking-wider">Work dir</label>
						<input
							id="workdir"
							bind:value={newWorkDir}
							placeholder="/path/to/project"
							required
							class="w-full bg-surface-2 border border-border rounded-md px-3 py-2 font-mono text-sm text-text-primary placeholder:text-text-muted/50 focus:outline-none focus:border-amber-dim focus:ring-1 focus:ring-amber-dim/30 transition-colors"
						/>
					</div>
				</div>
				<div class="flex justify-end">
					<button
						type="submit"
						disabled={creating}
						class="h-9 px-5 rounded-md bg-amber-glow text-surface-0 font-mono text-sm font-medium hover:bg-amber-bright disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
					>
						{creating ? 'Creating...' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Command list -->
	{#if loading}
		<div class="text-center py-16">
			<div class="inline-block w-5 h-5 border-2 border-surface-4 border-t-amber-glow rounded-full" style="animation: spin 0.8s linear infinite;"></div>
			<p class="font-mono text-sm text-text-muted mt-3">Loading commands...</p>
		</div>
	{:else if commands.length === 0 && !showForm}
		<div class="border border-border-subtle border-dashed rounded-lg py-16 text-center">
			<div class="w-12 h-12 rounded-full bg-surface-2 border border-border flex items-center justify-center mx-auto mb-4">
				<svg class="w-5 h-5 text-text-muted" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
					<rect x="2" y="3" width="12" height="10" rx="1.5" />
					<path d="M5 7l2 2 2-2" />
				</svg>
			</div>
			<p class="font-mono text-sm text-text-muted">No commands registered</p>
			<p class="text-xs text-text-muted/60 mt-1">Create one to get started</p>
		</div>
	{:else}
		<div class="space-y-2">
			{#each commands as cmd, i (cmd.id)}
				{@const status = statuses[cmd.id] ?? 'not_started'}
				<div
					class="group bg-surface-1 border border-border hover:border-border/80 rounded-lg transition-all animate-fade-in"
					style="animation-delay: {i * 50}ms"
				>
					<div class="px-5 py-4 flex items-center gap-4">
						<!-- Status dot -->
						<div class="shrink-0">
							{#if status === 'running'}
								<div class="w-2.5 h-2.5 rounded-full bg-signal-run" style="animation: pulse-dot 2s ease-in-out infinite;"></div>
							{:else if status === 'stopped'}
								<div class="w-2.5 h-2.5 rounded-full bg-signal-stop"></div>
							{:else}
								<div class="w-2.5 h-2.5 rounded-full bg-surface-4"></div>
							{/if}
						</div>

						<!-- Command info -->
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2.5">
								<a
									href="{base}/commands/{cmd.id}"
									class="font-mono text-sm font-medium text-text-primary hover:text-amber-glow transition-colors no-underline"
								>
									{cmd.name}
								</a>
								<span class="font-mono text-[10px] uppercase tracking-widest px-1.5 py-0.5 rounded border {statusColor(status)}">
									{statusLabel(status)}
								</span>
							</div>
							<div class="flex items-center gap-3 mt-1">
								<code class="font-mono text-xs text-text-secondary truncate">{cmd.command}</code>
								<span class="font-mono text-[10px] text-text-muted truncate hidden sm:inline">{cmd.work_dir}</span>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
							{#if status === 'running'}
								<button
									onclick={() => handleStop(cmd.id)}
									class="h-7 px-3 rounded bg-signal-stop-bg border border-signal-stop/20 font-mono text-xs text-signal-stop hover:bg-signal-stop/20 transition-colors"
								>
									Stop
								</button>
							{:else}
								<button
									onclick={() => handleStart(cmd.id)}
									class="h-7 px-3 rounded bg-signal-run-bg border border-signal-run/20 font-mono text-xs text-signal-run hover:bg-signal-run/20 transition-colors"
								>
									Start
								</button>
							{/if}
							{#if status !== 'running'}
								<button
									onclick={() => handleDelete(cmd.id)}
									class="h-7 w-7 rounded bg-surface-2 border border-border flex items-center justify-center text-text-muted hover:text-signal-stop hover:border-signal-stop/30 transition-colors"
									title="Delete"
								>
									<svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
										<path d="M4 4l8 8M12 4l-8 8" />
									</svg>
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
