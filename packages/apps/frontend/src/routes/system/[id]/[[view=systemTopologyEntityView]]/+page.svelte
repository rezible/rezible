<script lang="ts">
	import { page } from "$app/state";
	import SystemTopologyEntityView from "$features/system-topology/views/system-topology-entity";
</script>

<SystemTopologyEntityView id={page.params.id || ""} />
