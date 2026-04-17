<script lang="ts">
	import { Toaster as Sonner, type ToasterProps as SonnerProps } from "svelte-sonner";
	import { mode } from "mode-watcher";
	import RiLoaderLine from 'remixicon-svelte/icons/loader-line';
	import RiCheckboxCircleLine from 'remixicon-svelte/icons/checkbox-circle-line';
	import RiErrorWarningLine from 'remixicon-svelte/icons/error-warning-line';
	import RiInformationLine from 'remixicon-svelte/icons/information-line';
	import RiCloseCircleLine from 'remixicon-svelte/icons/close-circle-line';

	let { ...restProps }: SonnerProps = $props();
</script>

<Sonner
	theme={mode.current}
	class="toaster group"
	style="--normal-bg: var(--color-popover); --normal-text: var(--color-popover-foreground); --normal-border: var(--color-border);"
	{...restProps}
>
	{#snippet loadingIcon()}
		<RiLoaderLine class="size-4 animate-spin" />
	{/snippet}
	{#snippet successIcon()}
		<RiCheckboxCircleLine class="size-4" />
	{/snippet}
	{#snippet errorIcon()}
		<RiErrorWarningLine class="size-4" />
	{/snippet}
	{#snippet infoIcon()}
		<RiInformationLine class="size-4" />
	{/snippet}
	{#snippet warningIcon()}
		<RiCloseCircleLine class="size-4" />
	{/snippet}
</Sonner>
