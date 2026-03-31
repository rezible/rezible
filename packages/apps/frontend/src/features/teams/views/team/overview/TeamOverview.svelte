<script lang="ts">
	import { mdiSlack } from "@mdi/js";
	import TeamUsersTable from "./TeamUsersTable.svelte";
	import TeamRostersTable from "./TeamRostersTable.svelte";

	const slackChannelOptions = [{ label: "team-channel", value: "foo" }];
	const slackChannel = slackChannelOptions[0].value;
</script>

<div class="flex gap-2 p-2">
	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Details</span>

		<span>slack channel</span>
		<!-- <SelectField
			label="Slack Channel"
			labelPlacement="top"
			icon={mdiSlack}
			readonly
			options={slackChannelOptions}
			value={slackChannel}
			clearable={false}
		/> -->
	</div>

	<TeamUsersTable />

	<TeamRostersTable />
</div>
