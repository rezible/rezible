<script lang="ts">
	import type { PageProps } from "./$types";
	import OncallShiftView from "$features/oncall/views/shift";

	const { params }: PageProps = $props();
</script>

<OncallShiftView id={params.id} />
