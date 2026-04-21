<script lang="ts">
	import { useAppShell } from "$lib/app-shell.svelte";
	import * as Sidebar from "$components/ui/sidebar";
	import * as Breadcrumb from "$components/ui/breadcrumb";

	const shell = useAppShell();

	const pageBreadcrumbs = $derived(shell.breadcrumbs);
	const pageActions = $derived(shell.pageActions);
	const propsFn = $derived(pageActions?.propsFn ?? (() => ({})));
	const pageActionsProps = $derived.by(() => (propsFn()));
</script>

<div class="flex items-center gap-2">
    <Sidebar.Trigger size="icon-lg" />  
    <Breadcrumb.Root>
        <Breadcrumb.List>
            {#each pageBreadcrumbs as crumb, i}
                {#if i > 0}
                    <Breadcrumb.Separator />
                {/if}
                <Breadcrumb.Item>
                    {#if i < (pageBreadcrumbs.length - 1)}
                        <Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
                    {:else}
                        <Breadcrumb.Page>{crumb.label}</Breadcrumb.Page>
                    {/if}
                </Breadcrumb.Item>
            {/each}
        </Breadcrumb.List>
    </Breadcrumb.Root>
</div>

{#if pageActions}
    <div class="flex items-center">
        <pageActions.component {...pageActionsProps} />
    </div>
{/if}