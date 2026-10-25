<script lang="ts">
	import type { PageProps } from "./$types";
	import TaskDetailView from "$features/tasks/views/task-detail";

	const { params }: PageProps = $props();
</script>

<TaskDetailView taskId={params.id} />
