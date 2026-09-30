<script lang="ts">
	import { SvelteSet } from "svelte/reactivity";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import { Checkbox } from "$components/ui/checkbox";
	import { Spinner } from "$components/ui/spinner";

	import { useIntegrationProviderController } from "../integration-provider/controller.svelte";

	type Props = {
		name: string;
		title: string;
		description: string;
	};
	const { name, title, description }: Props = $props();

	const ctrl = useIntegrationProviderController();

	const options = $derived(ctrl.integrations.installTargetsFor(name));
	const pending = $derived(ctrl.installTargetsPendingName === name);
	const error = $derived(ctrl.installTargetsErrorFor(name));

	const selectedRefs = new SvelteSet<string>();

	const toggle = (ref: string, selected: boolean) => {
		if (selected) {
			selectedRefs.add(ref);
		} else {
			selectedRefs.delete(ref);
		}
	};

	const confirm = () => {
		ctrl.installFromTargets(name, Array.from(selectedRefs));
	};
</script>

{#if options.length > 0}
	<div class="flex flex-col gap-3 border border-primary/40 bg-muted/40 p-4">
		<div class="flex flex-col gap-1">
			<span class="text-sm font-medium">{title}</span>
			<span class="text-sm text-muted-foreground">{description}</span>
		</div>

		<div class="flex flex-col gap-2">
			{#each options as option (option.resourceRef.resourceRef)}
				{@const ref = option.resourceRef.resourceRef}
				<label class="flex items-center gap-3 border bg-background p-3 text-sm">
					<Checkbox
						checked={selectedRefs.has(ref)}
						onCheckedChange={(checked) => toggle(ref, !!checked)}
					/>
					<span class="font-medium">{option.displayName}</span>
				</label>
			{/each}
		</div>

		{#if error}
			<InlineAlert {error} dismissable={false} />
		{/if}

		<Button class="w-fit" disabled={selectedRefs.size === 0 || pending} onclick={confirm}>
			{#if pending}
				<Spinner />
			{/if}
			Connect selected
		</Button>
	</div>
{/if}
