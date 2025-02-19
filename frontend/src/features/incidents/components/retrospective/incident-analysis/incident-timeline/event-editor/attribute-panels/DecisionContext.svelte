<script lang="ts">
	import { mdiGithub, mdiPlus, mdiSlack, mdiWeb } from "@mdi/js";
	import { Button, Field, Icon, ListItem, SelectField, TextField } from "svelte-ux";
	import { eventAttributes } from "./eventAttributes.svelte";

	const decisionCtx = $derived(eventAttributes.decisionContext);

	let newOption = $state<string>();
	const confirmAddingOption = () => {
		if (!newOption) return;
		decisionCtx.optionsConsidered.push($state.snapshot(newOption));
		newOption = undefined;
	};

	let newConstraint = $state<string>();
	const confirmAddingConstraint = () => {
		if (!newConstraint) return;
		decisionCtx.constraints.push($state.snapshot(newConstraint));
		newConstraint = undefined;
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	<div class="flex flex-col gap-2 border p-2">
		<span class="text-surface-content">Options Considered</span>

		{#each decisionCtx.optionsConsidered as opt, i}
			<ListItem
				title={opt}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Button>edit</Button>
				</div>
			</ListItem>
		{/each}

		<TextField dense bind:value={newOption} label="Add Option">
			<span slot="append">
				<Button
					icon={mdiPlus}
					class="text-surface-content/50 p-2"
					on:click={confirmAddingOption}
					disabled={!newOption}
				/>
			</span>
		</TextField>
	</div>

	<div class="flex flex-col gap-2 border p-2">
		<span class="text-surface-content">Constraints</span>

		{#each decisionCtx.constraints as cst, i}
			<ListItem
				title={cst}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Button>edit</Button>
				</div>
			</ListItem>
		{/each}

		<TextField dense bind:value={newConstraint} label="Add Constraint">
			<span slot="append">
				<Button
					icon={mdiPlus}
					class="text-surface-content/50 p-2"
					on:click={confirmAddingConstraint}
					disabled={!newConstraint}
				/>
			</span>
		</TextField>
	</div>

	<div class="flex flex-col gap-2 border p-2">
		<span class="text-surface-content">Decision Rationale</span>

		<TextField multiline bind:value={decisionCtx.decisionRationale} label=""></TextField>
	</div>
</div>
