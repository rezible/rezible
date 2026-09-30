<script lang="ts">
	import { afterNavigate } from "$app/navigation";
	import type { PageDescriptor } from "$lib/app-shell.svelte";
	import * as Sidebar from "$components/ui/sidebar";
	import * as Breadcrumb from "$components/ui/breadcrumb";
	import { Separator } from "$components/ui/separator";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import RelatedObjectLinks from "./RelatedObjectLinks.svelte";

	type Props = {
		descriptor: PageDescriptor;
	};
	const { descriptor }: Props = $props();

	let currentCrumb = $state<HTMLElement>();

	const parents = $derived(descriptor.parents ?? []);
	const related = $derived(descriptor.related ?? []);
	const showActionSeparator = $derived(related.length > 0 && !!descriptor.pageActions);
	const currentCrumbClass =
		"min-w-0 max-w-[56ch] truncate text-sm font-medium text-foreground outline-none";

	// Move focus to the current page crumb on meaningful route changes, leaving
	// in-page navigation and the initial page load undisturbed.
	afterNavigate(({ from, to }) => {
		if (!from || from.url.pathname === to?.url.pathname) {
			return;
		}
		currentCrumb?.focus();
	});
</script>

<header class="bg-card flex h-14 shrink-0 items-center gap-3 border-b px-4 sm:px-6">
	<div class="flex min-w-0 flex-1 items-center gap-2">
		<Sidebar.Trigger class="-ms-1 md:hidden" />
		<Breadcrumb.Root class="min-w-0">
			<Breadcrumb.List class="flex-nowrap">
				{#each parents as parent, index (parent.path)}
					{@const isLastParent = index === parents.length - 1}
					<Breadcrumb.Item class={isLastParent ? undefined : "max-sm:hidden"}>
						<Breadcrumb.Link
							href={parent.path}
							class="text-xs whitespace-nowrap text-muted-foreground hover:text-foreground"
						>
							{parent.label}
						</Breadcrumb.Link>
					</Breadcrumb.Item>
					<Breadcrumb.Separator
						class={isLastParent
							? "text-muted-foreground/60"
							: "text-muted-foreground/60 max-sm:hidden"}
					/>
				{/each}
				<Breadcrumb.Item class="min-w-0">
					{#if descriptor.contentHeading}
						<span
							bind:this={currentCrumb}
							tabindex="-1"
							aria-current="page"
							title={descriptor.title}
							class={currentCrumbClass}
						>
							{descriptor.title}
						</span>
					{:else}
						<h1
							bind:this={currentCrumb}
							tabindex="-1"
							title={descriptor.title}
							class={currentCrumbClass}
						>
							{descriptor.title}
						</h1>
					{/if}
				</Breadcrumb.Item>
			</Breadcrumb.List>
		</Breadcrumb.Root>
		{#if descriptor.status}
			<StatusBadge status={descriptor.status} class="ms-1 shrink-0" />
		{/if}
	</div>
	<div class="flex shrink-0 items-center gap-2">
		{#if related.length}
			<RelatedObjectLinks links={related} />
		{/if}
		{#if showActionSeparator}
			<Separator orientation="vertical" class="mx-1 h-5! max-sm:hidden" />
		{/if}
		{@render descriptor.pageActions?.()}
	</div>
</header>
