<script lang="ts">
	import { createMutation, createQuery } from "@tanstack/svelte-query";
	import { mdiMagnify } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { addWatchedOncallRosterMutation, listOncallRostersOptions } from "$lib/api";
	import { debounce } from "$lib/utils.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import FormDialog from "$components/form-dialog/FormDialog.svelte";

	type Props = {
		open: boolean;
		current: string[];
		onUpdated: () => void;
	};
	let { open = $bindable(), current, onUpdated }: Props = $props();

	const currentMap = $derived(new Set<string>(current));

	let search = $state<string>();
	const updateSearch = debounce((v: string) => (search = (!!v ? v : undefined)), 500);
	const rostersQuery = createQuery(() => ({
		...listOncallRostersOptions({query: {search}}),
		enabled: open,
	}));
	const rosters = $derived(rostersQuery.data?.data ?? []);

	const optionsLoading = $derived(rostersQuery.isFetching);
	const options = $derived(rosters
		.filter(r => (!currentMap.has(r.id)))
		.map((r) => ({ value: r.id, label: r.attributes.name }))
	);

	let value = $state<string>();
	const valueRoster = $derived(options.find(o => (o.value === value)));
	const saveEnabled = $derived(!!value);
	const confirmText = $derived(!!valueRoster ? `Watch ${valueRoster.label}` : "Watch");

	const onClose = () => {
		open = false;
		search = undefined;
		value = undefined;
	};

	const watchRosterMutation = createMutation(() => ({
		...addWatchedOncallRosterMutation(),
		onSuccess: () => {
			onClose();
			onUpdated();
		}
	}))
	const onConfirm = () => {
		if (!value) return;
		watchRosterMutation.mutate({path: {id: value}});
	};
</script>

<FormDialog title="Add Watched Roster" {open} {onClose} {onConfirm} {saveEnabled} {confirmText}>
	<div class="w-full gap-2 p-2">
		<!--SelectField
			label="Name"
			placeholder="Search Rosters"
			bind:value
			{options}
			on:inputChange={e => updateSearch(e.detail)}
			search={async (v, o) => (o)}
			loading={optionsLoading}
		>
			<div slot="prepend" class="mr-2">
				{#if value}
					<Avatar kind="roster" id={value} size={28} />
				{:else}
					<Icon data={mdiMagnify} />
				{/if}
			</div>
		</SelectField-->
	</div>
</FormDialog>
