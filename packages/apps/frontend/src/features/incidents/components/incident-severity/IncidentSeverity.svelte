<script lang="ts">
	import type { IncidentSeverity } from "$lib/api";
	import { Badge } from "$components/ui/badge";
	import { incidentSeverityVariant } from "./severity";
	type Props = { severity?: IncidentSeverity };
	let { severity }: Props = $props();
</script>

<Badge variant={incidentSeverityVariant(severity)}>{severity?.attributes.name ?? "Unspecified"}</Badge>
