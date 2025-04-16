<script lang="ts">
	import { Button, Header, Icon } from "svelte-ux";
	import { mdiChevronRight, mdiCircleMedium } from "@mdi/js";
	import { useIncidentCollaboration } from "../collaboration.svelte";
	import { cls } from "@layerstack/tailwind";
	import { WebSocketStatus } from "@hocuspocus/provider";

	const collaboration = useIncidentCollaboration();

	const connected = $derived(collaboration.connectionStatus === WebSocketStatus.Connected);
	const connecting = $derived(collaboration.connectionStatus === WebSocketStatus.Connecting);
</script>

<div class="w-64 border overflow-y-auto p-2">
	<Header>
		<span slot="title" class="text-xl flex gap-1 items-center">
			Context
			<Icon
				data={mdiCircleMedium}
				classes={{
					root: "opacity-70",
					path: cls(connected ? "fill-success" : (connecting ? "fill-warning" : "fill-default")),
				}}
			/>
		</span>
		<div slot="actions">
			<Button iconOnly icon={mdiChevronRight} />
		</div>
	</Header>

	
</div>
