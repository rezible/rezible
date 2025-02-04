<script lang="ts">
	import type { JSONContent, HTMLContent } from "@tiptap/core";
	import { Editor as SvelteEditor, EditorContent } from "svelte-tiptap";
	import { configureBaseExtensions } from "@rezible/documents/tiptap-extensions";
	import { onMount } from "svelte";
	import { Header } from "svelte-ux";

	const content: HTMLContent = ``;

	/*
	let editor = $state<SvelteEditor>();

	onMount(() => {
		editor = new SvelteEditor({
			content,
			extensions: [...configureBaseExtensions(true)],
			editable: false,
			editorProps: {
				attributes: {
					class: 'm-5 max-w-none focus:outline-none h-full'
				}
			},
		});

		return () => {if (editor) editor.destroy()}
	})
	*/
</script>

<!--
{#if editor}
<div
	class="grow overflow-y-auto min-h-0 bg-surface-100"
	tabindex="-1"
	spellcheck="false"
>
	<EditorContent {editor} />
</div>
{/if}
-->

<div
	class="w-full flex flex-col min-h-0 max-h-full justify-center rounded-lg shadow-md border p-2 bg-surface-200/80 text-surface-content"
>
	<div class="border-b-2 border-primary-900 p-2 px-8 mb-2">
		<Header title="Search Service Oncall Handover from @johnd to @texm">
			<span slot="subheading">
				<span class="text-sm">October 21-28, APAC to EMEA</span>
			</span>
		</Header>
	</div>

	<div class="px-6 py-4 overflow-y-auto">
		<div class="space-y-4">
			<p>Busy Shift, hope yours is less eventful!</p>

			<!-- System Status -->
			<div>
				<h2 class="text-lg font-semibold">Current System Status</h2>
				<ul class="mt-2 text-sm">
					<li>
						Search latency: <span class="font-medium"
							>p95 at 180ms</span
						> (within SLA)
					</li>
					<li>
						Error rate: <span class="font-medium">0.02%</span>
					</li>
					<li>
						Index freshness: <span class="font-medium"
							>3min lag</span
						> (normal)
					</li>
				</ul>
			</div>

			<!-- Active Incidents -->
			<div>
				<h2 class="text-lg font-semibold">Incidents</h2>
				<a href="/incidents">
					<div
						class="mt-2 p-3 bg-danger-900/20 rounded-md hover:bg-danger-900/40 w-fit"
					>
						<p class="text-sm font-medium">
							INC-4857: Intermittent 5xx errors from EU-West-2
							replica
						</p>
						<ul class="mt-2 text-sm">
							<li>Started: 10/25 03:00 UTC</li>
							<li>Impact: ~0.1% of EU traffic affected</li>
							<li>
								Mitigation: Enabled fallback to EU-Central-1
							</li>
							<li>
								Next steps: AWS support investigating underlying
								cause
							</li>
						</ul>
					</div>
				</a>
				<a href="/incidents">
					<div
						class="mt-2 p-3 bg-danger-900/20 rounded-md hover:bg-danger-900/40 w-fit"
					>
						<p class="text-sm font-medium">
							INC-4862: Elevated query latency for long-tail
							searches
						</p>
						<ul class="mt-2 text-sm">
							<li>Started: 10/25 07:15 UTC</li>
							<li>
								Impact: ~5% of complex queries seeing 2-3s
								latency
							</li>
							<li>
								Mitigation: Rolled back schema changes, backfill
								in progress
							</li>
							<li>
								Next steps: Monitor cache hit rates, retry
								migration w/ safeguards
							</li>
						</ul>
					</div>
				</a>
			</div>

			<!-- Recent Changes -->
			<div>
				<h2 class="text-lg font-semibold">Recent Changes</h2>
				<ul class="mt-2 text-sm list-disc pl-5">
					<li>Deployed v2.8.4 to prod (ranking algorithm update)</li>
					<li>
						Enabled circuit breaker for downstream recommendations
						service
					</li>
					<li>Increased cache TTL to 5min on static queries</li>
				</ul>
			</div>

			<!-- Watching -->
			<div>
				<h2 class="text-lg font-semibold">Watching</h2>
				<ul class="mt-2 text-sm list-disc pl-5">
					<li>Memory usage in AP-South-1 cluster climbing (~82%)</li>
					<li>
						New ranking model showing 2% lower CTR than baseline
					</li>
				</ul>
			</div>

			<!-- Handoff Items -->
			<div>
				<h2 class="text-lg font-semibold">Handoff Items</h2>
				<ol class="mt-2 text-sm list-decimal pl-5">
					<li>
						Monitor EU-West-2 recovery after AWS completes
						investigation
					</li>
					<li>Memory investigation ticket: SRCH-2245</li>
					<li>
						Ranking A/B test results due in 6hrs - rollback if CTR
						gap exceeds 3%
					</li>
				</ol>
			</div>
		</div>
	</div>
</div>
