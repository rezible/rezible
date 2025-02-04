<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Header } from "svelte-ux";
	import { getMeetingSessionOptions, type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	type Props = { session: MeetingSession };
	const { session }: Props = $props();

	const attr = $derived(session.attributes);

	// const queryOptions = () => getMeetingSessionOptions({path: {id}});
	// const query = createQuery(queryOptions);
</script>

<div class="grid grid-cols-2 gap-2 flex-1 min-h-0 overflow-y-hidden">
	<div class="border p-2">
		<Header title="Session Document" />
	</div>

	<div class="border p-2">
		<Header title="Discussion" />
	</div>
</div>
