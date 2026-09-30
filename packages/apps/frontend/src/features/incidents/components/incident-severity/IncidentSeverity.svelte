<script lang="ts">
	import type { ExpandableIncidentSeverityAttributes } from "$lib/api";
	import { Badge } from "$components/ui/badge";
	import { incidentSeverityVariant } from "./severity";
	type Props = { severity?: ExpandableIncidentSeverityAttributes };
	let { severity }: Props = $props();
</script>

<Badge variant={incidentSeverityVariant(severity)}>{severity?.attributes?.name ?? "Unspecified"}</Badge>
