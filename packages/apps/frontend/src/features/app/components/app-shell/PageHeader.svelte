<script lang="ts">
	import { useAppShell } from "$lib/app-shell.svelte";
	import * as Breadcrumb from "$components/ui/breadcrumb";

	const shell = useAppShell();

	const descriptor = $derived(shell.pageDescriptor);
	const pageActions = $derived(descriptor?.actions);
</script>

{#if descriptor}
	<div class="flex items-center gap-2 text-lg">
		<!-- <Sidebar.Trigger size="icon-lg" />   -->
		<Breadcrumb.Root>
			<Breadcrumb.List>
				{#each descriptor.parents ?? [] as parent (parent.path)}
					<Breadcrumb.Item>
						<Breadcrumb.Link href={parent.path}>{parent.label}</Breadcrumb.Link>
					</Breadcrumb.Item>
					<Breadcrumb.Separator />
				{/each}
				<Breadcrumb.Item>
					<h1><Breadcrumb.Page>{descriptor.title}</Breadcrumb.Page></h1>
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
