<script>
	import '../app.css';
	import { base } from '$app/paths';

	let { children } = $props();
</script>

<svelte:head>
	<title>AI Sensors</title>
</svelte:head>

<div class="min-h-screen bg-surface-0">
	<!-- Top bar -->
	<header class="border-b border-border bg-surface-1/80 backdrop-blur-sm sticky top-0 z-50">
		<div class="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
			<a href="{base}/" class="flex items-center gap-3 group no-underline">
				<!-- Sensor icon -->
				<div class="w-8 h-8 rounded-md bg-amber-wash border border-amber-dim/30 flex items-center justify-center group-hover:border-amber-glow/50 transition-colors">
					<svg class="w-4 h-4 text-amber-glow" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
						<circle cx="8" cy="8" r="2" />
						<path d="M4.5 4.5a5 5 0 0 1 7 0" />
						<path d="M2.5 2.5a8 8 0 0 1 11 0" />
						<path d="M4.5 11.5a5 5 0 0 0 7 0" />
						<path d="M2.5 13.5a8 8 0 0 0 11 0" />
					</svg>
				</div>
				<div>
					<span class="font-mono text-sm font-semibold text-text-primary tracking-tight">AI Sensors</span>
					<span class="hidden sm:inline font-mono text-xs text-text-muted ml-2">control panel</span>
				</div>
			</a>

			<!-- Status indicator -->
			<div class="flex items-center gap-2 font-mono text-xs text-text-muted">
				<div class="w-1.5 h-1.5 rounded-full bg-signal-run" style="animation: pulse-dot 2s ease-in-out infinite;"></div>
				<span>online</span>
			</div>
		</div>
	</header>

	<!-- Content -->
	<main class="max-w-6xl mx-auto px-6 py-8">
		{@render children()}
	</main>
</div>
