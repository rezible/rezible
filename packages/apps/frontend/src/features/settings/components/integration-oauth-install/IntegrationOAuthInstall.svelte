<script lang="ts">
	import type { Snippet } from "svelte";

	import * as Alert from "$components/ui/alert";
	import { Button, type ButtonSize, type ButtonVariant } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";

	import { useIntegrationProviderController } from "../integration-provider/controller.svelte";

	type Props = {
		name: string;
		label: string;
		// Distinguishes several buttons that install the same integration.
		origin?: string;
		variant?: ButtonVariant;
		size?: ButtonSize;
		icon?: Snippet;
	};
	const { name, label, origin = name, variant = "default", size = "default", icon }: Props = $props();

	const ctrl = useIntegrationProviderController();

	const pending = $derived(ctrl.oauth.isPending(origin));
	const error = $derived(ctrl.oauth.errorFor(origin));

	const start = () => {
		ctrl.oauth.startFlowFor(name, origin);
	};
</script>

<div class="flex flex-col gap-3">
	<Button {variant} {size} class="w-fit" disabled={ctrl.oauth.inFlow} onclick={start}>
		{#if pending}
			<Spinner />
			Waiting for sign-in…
		{:else}
			{@render icon?.()}
			{label}
		{/if}
	</Button>

	{#if error}
		<Alert.Root variant="destructive">
			<Alert.Title>{error.title ?? "Sign-in failed"}</Alert.Title>
			<Alert.Description>{error.detail ?? "Sign-in did not complete."}</Alert.Description>
			<Alert.Action>
				<Button size="sm" variant="outline" onclick={start}>Try again</Button>
			</Alert.Action>
		</Alert.Root>
	{/if}
</div>
