<script lang="ts">
	import { page } from "$app/state";
	import { useAppShell } from "$lib/app-shell.svelte";
	import * as Breadcrumb from "$components/ui/breadcrumb";
	import type { ResolvedPathname } from "$app/types";

	const shell = useAppShell();

	const descriptor = $derived(shell.pageDescriptor);
	const pageActions = $derived(descriptor?.actions);

	let heading = $state<HTMLElement>();
	let focusedPath: string | undefined = undefined;
	let initialLoad = true;

	// Move focus to the page heading on meaningful route changes, leaving
	// in-page navigation and the initial page load undisturbed.
	const maybeMoveNavigationFocus = (pathname: ResolvedPathname) => {
		if (initialLoad) {
			initialLoad = false;
			focusedPath = pathname;
			return;
		}
		// Wait for the new view's heading to render before focusing it.
		if (!!heading && focusedPath !== pathname) {
			focusedPath = pathname;
			heading.focus();
		}
	}
	$effect(() => maybeMoveNavigationFocus(page.url.pathname));
</script>

{#if descriptor}
	<div class="flex items-center gap-2 text-lg">
		<Breadcrumb.Root>
			<Breadcrumb.List>
				{#each descriptor.parents ?? [] as parent (parent.path)}
					<Breadcrumb.Item>
						<Breadcrumb.Link href={parent.path}>{parent.label}</Breadcrumb.Link>
					</Breadcrumb.Item>
					<Breadcrumb.Separator />
				{/each}
				<Breadcrumb.Item>
					<h1 bind:this={heading} tabindex="-1" class="outline-none">
						<Breadcrumb.Page>{descriptor.title}</Breadcrumb.Page>
					</h1>
				</Breadcrumb.Item>
			</Breadcrumb.List>
		</Breadcrumb.Root>
	</div>
{/if}

{#if pageActions}
	<div class="flex items-center">
		<pageActions.component {...pageActions.props} />
	</div>
{/if}
