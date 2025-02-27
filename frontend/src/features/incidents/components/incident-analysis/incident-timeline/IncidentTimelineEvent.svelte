<script lang="ts" module>
	export type TimelineEventComponentProps = {
		label: string;
	};
</script>

<script lang="ts">
	import { onMount } from "svelte";

	let { label }: TimelineEventComponentProps = $props();

	onMount(() => {
		// console.log("mounted");
		// return () => {console.log("unmounting")}
	});
</script>

<span>{label}</span>
