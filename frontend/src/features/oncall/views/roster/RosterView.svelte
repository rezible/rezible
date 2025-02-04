<script lang="ts">
	import { type OncallRoster } from "$lib/api";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	const attr = $derived(roster.attributes);
</script>

<span>{attr.name}</span>
