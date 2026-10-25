<script lang="ts">
	import OncallRosterView from "$features/oncall/views/roster";

	const { data } = $props();
</script>

<OncallRosterView slug={data.slug} />
