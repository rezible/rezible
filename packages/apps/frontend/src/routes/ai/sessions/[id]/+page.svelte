<script lang="ts">
	import type { PageProps } from "./$types";
	import { SessionDetailView } from "$features/ai/views/session-detail";

	const { params }: PageProps = $props();
</script>

<SessionDetailView sessionId={params.id} />
