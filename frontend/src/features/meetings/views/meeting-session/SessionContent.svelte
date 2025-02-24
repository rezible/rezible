<script lang="ts">
	import { Header } from "svelte-ux";
	import type { MeetingSession } from "$lib/api";

	type Props = { session: MeetingSession };
	const { session }: Props = $props();

	const attr = $derived(session.attributes);
</script>

<div class="grid grid-cols-2 gap-2 flex-1 min-h-0 overflow-y-hidden">
	<div class="border p-2">
		<Header title="Session Document" />
	</div>

	<div class="border p-2">
		<Header title="Discussion" />
	</div>
</div>
