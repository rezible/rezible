<script lang="ts">
	import { afterNavigate } from "$app/navigation";
	import type { PageDescriptor } from "$lib/app-shell.svelte";
	import * as Sidebar from "$components/ui/sidebar";
	import * as Breadcrumb from "$components/ui/breadcrumb";
	import { Badge } from "$components/ui/badge";

	type Props = {
		pageDescriptor: PageDescriptor;
		featureRailActive: boolean;
	};
	const { pageDescriptor, featureRailActive }: Props = $props();

	let heading = $state<HTMLElement>();

	// Move focus to the page heading on meaningful route changes, leaving
	// in-page navigation and the initial page load undisturbed.
	afterNavigate(({ from, to }) => {
		if (!from || from.url.pathname === to?.url.pathname) return;
		heading?.focus();
	});
</script>

<div class="flex min-w-0 flex-1 items-center justify-between gap-4">
	<div class="flex min-w-0 flex-1 items-center gap-2 overflow-hidden">
		<Sidebar.Trigger class={featureRailActive ? undefined : "md:hidden"} />
		<Breadcrumb.Root class="min-w-0">
			<Breadcrumb.List>
				{#each pageDescriptor.parents ?? [] as parent (parent.path)}
					<Breadcrumb.Item>
						<Breadcrumb.Link href={parent.path}>{parent.label}</Breadcrumb.Link>
					</Breadcrumb.Item>
					<Breadcrumb.Separator />
				{/each}
				<Breadcrumb.Item>
					<h1 bind:this={heading} tabindex="-1" class="min-w-0 truncate outline-none">
						<Breadcrumb.Page class="truncate">{pageDescriptor.title}</Breadcrumb.Page>
					</h1>
				</Breadcrumb.Item>
			</Breadcrumb.List>
		</Breadcrumb.Root>
		{#if pageDescriptor.status}<Badge variant="secondary">{pageDescriptor.status}</Badge>{/if}
	</div>
	<div class="flex shrink-0 items-center">
		{@render pageDescriptor.pageActions?.()}
	</div>
</div>
