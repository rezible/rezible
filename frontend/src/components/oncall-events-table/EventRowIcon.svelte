<script lang="ts">
	import type { OncallEventAttributes } from "$lib/api";
	import { mdiPhoneAlert, mdiFire, mdiChatQuestion } from "@mdi/js";
	import { Icon } from "svelte-ux";

	type Props = {
		kind: OncallEventAttributes["kind"];
	}
	const { kind }: Props = $props();

	const icon = $derived.by(() => {
		switch (kind) {
		case "incident": return mdiFire;
		case "alert": return mdiPhoneAlert;
		}
		return mdiChatQuestion;
	});
</script>

<div class="grid place-items-center">
	<Icon
		data={icon}
		classes={{ root: `${kind === 'incident' ? 'bg-danger-900/50' : 'bg-warning-700/50'} rounded-full p-2 w-auto h-10 text-white` }}
	/>
</div>