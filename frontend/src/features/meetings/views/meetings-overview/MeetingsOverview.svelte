<script lang="ts">
	import { Button, Header, Icon } from "svelte-ux";
	import { mdiPlus } from "@mdi/js";

	import CreateMeetingDialog from "./CreateMeetingDialog.svelte";
	import ScheduledMeetingsList from "./ScheduledMeetingsList.svelte";
	import UpcomingSessionsList from "./UpcomingSessionsList.svelte";
	import SplitPage from "$components/split-page/SplitPage.svelte";

	let createOpen = $state(false);

	const onMeetingCreated = () => {
		createOpen = false;
	};
</script>

<SplitPage nav={scheduledListNav}>
	<Header title="Upcoming" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<svelte:fragment slot="actions">
			<Button
				classes={{ root: "w-fit h-fit" }}
				variant="fill"
				color="primary"
				on:click={() => {
					createOpen = true;
				}}
				>
				Create New
				<Icon data={mdiPlus} />
			</Button>
		</svelte:fragment>
	</Header>

	<UpcomingSessionsList />
</SplitPage>

{#snippet scheduledListNav()}
	<Header title="Scheduled" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<svelte:fragment slot="actions">
			<Button
				classes={{ root: "w-fit" }}
				variant="fill"
				color="primary"
				on:click={() => {
					createOpen = true;
				}}
				>
				Schedule New
				<Icon data={mdiPlus} />
			</Button>
		</svelte:fragment>
	</Header>

	<ScheduledMeetingsList />
{/snippet}

<CreateMeetingDialog bind:open={createOpen} onCreated={onMeetingCreated} />
