<script lang="ts">
	import type { AuthSessionConfig } from "$lib/api";
	import { AuthFlowController } from "./authFlowController.svelte";

	import { mdiKey } from "@mdi/js";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import { Button } from "$components/ui/button";

    type Props = {
        config: AuthSessionConfig;
    };
    const { config }: Props = $props();

	const controller = $derived(new AuthFlowController(config));
</script>

{#if controller.loading}
    <LoadingIndicator />
{:else if !!controller.error}
    <InlineAlert 
        error={controller.error}
        onDismiss={() => {controller.clearError()}}
    />
{:else}
    <Button onclick={() => {controller.doSignIn()}} color="primary">
        <span class="flex items-center gap-2">
            Sign In
            <Icon data={mdiKey} />
        </span>
    </Button>
{/if}