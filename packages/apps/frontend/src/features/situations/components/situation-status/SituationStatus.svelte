<script lang="ts">
	import type { SituationAttributes } from "$lib/api";
	import RiEyeLine from "remixicon-svelte/icons/eye-line";
	import RiSearchLine from "remixicon-svelte/icons/search-line";
	import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";

	type Props = { attributes: SituationAttributes };
	let { attributes }: Props = $props();

	const statuses = {
		observed: { label: "Observed", icon: RiEyeLine, class: "text-status-info-foreground" },
		investigating: {
			label: "Investigating",
			icon: RiSearchLine,
			class: "text-status-warning-foreground",
		},
		closed: { label: "Closed", icon: RiCheckboxCircleLine, class: "text-muted-foreground" },
	};
	const status = $derived.by(() => {
		if (!!attributes.closedAt) return "closed";
		if (attributes.investigation) return "investigating";
		return "observed";
	});
	const presentation = $derived(statuses[status]);
</script>

<span class="inline-flex items-center gap-1.5 whitespace-nowrap text-xs {presentation.class}">
	<presentation.icon class="size-4 shrink-0" aria-hidden="true" />
	{presentation.label}
</span>
