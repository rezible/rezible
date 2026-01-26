<script lang="ts">
	import type { ConfigComponentProps } from '../types';
	import { onMount } from "svelte";

    const { integration, configured, onConfigChange }: ConfigComponentProps = $props();

    const configValid = $derived(configured?.attributes.configValid);

    onMount(() => {

    });
</script>

<span>slack config</span>