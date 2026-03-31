<script lang="ts">
	import { mdiChevronRight, mdiPencil } from "@mdi/js";
	import Header from "$components/header/Header.svelte";
	import { Button } from "$components/ui/button";

</script>

<div class="grid grid-cols-4 gap-2 h-full">
	<div class="flex flex-col h-full border">
		<Header title="Incident Retrospectives" classes={{title: "text-lg font-medium", root: "border-b h-10 px-1"}} />


		<div class="flex-1 flex flex-col overflow-y-auto">
			<div class="">
			</div>
		</div>
	</div>

	<div class="flex flex-col h-full border">
		<Header title="Links" classes={{title: "text-lg font-medium", root: "border-b h-10 px-1"}}>
			{#snippet actions()}
				<Button>Add Link</Button>
			{/snippet}
		</Header>


		<div class="flex-1 flex flex-col overflow-y-auto">
			<div class="">
				<pre>Onboarding documents
Team-specific tools/dashboards
Recent incident retrospectives
Recent meeting notes (ops reviews)</pre>
			</div>
		</div>
	</div>

	<div class="flex flex-col h-full border">
		<Header title="Playbooks" classes={{title: "text-lg font-medium", root: "border-b h-10 px-1"}}>
			{#snippet actions()}
				<Button>Add Playbook</Button>
			{/snippet}
		</Header>

		<div class="flex-1 flex flex-col overflow-y-auto p-1">
			<!-- playbook list item -->
			<div class="flex flex-col w-full border p-1">
				<Header title="Title">
					{#snippet actions()}
						<Button>edit</Button>
						<Button>delete</Button>
					{/snippet}
				</Header>
				<div class="border">
				<pre>Last updated date
Health score (based on usage/feedback)
Owner
Quick actions (edit, archive)
Tags for categorization (service, type of issue)</pre>
				</div>
			</div>
		</div>
	</div>
</div>