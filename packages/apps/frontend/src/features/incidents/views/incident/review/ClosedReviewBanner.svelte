<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import RiLockLine from "remixicon-svelte/icons/lock-line";
	import { useIncidentView } from "../controller.svelte";
	type Props = { content: "report" | "analysis" };
	const { content }: Props = $props();
	const view = useIncidentView();
</script>

{#if view.reviewClosed}
	<Alert.Root>
		<RiLockLine />
		<Alert.Title>This review is closed. The {content} is read-only.</Alert.Title>
		<Alert.Description>
			{view.closedReviewLine}
		</Alert.Description>
	</Alert.Root>
{/if}
