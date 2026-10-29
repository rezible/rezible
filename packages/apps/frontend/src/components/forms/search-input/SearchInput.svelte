<script lang="ts">
	import { Input } from "$components/ui/input";

	type Props = {
		value: string;
		/** Called with the raw text on every keystroke; callers debounce and trim. */
		oninput: (value: string) => void;
		label?: string;
	};

	const { value, oninput, label = "Search" }: Props = $props();

	const id = $props.id();
</script>

<div class="flex min-w-0 flex-col gap-1.5">
	<label for={`${id}-search`} class="region-label">{label}</label>
	<Input
		id={`${id}-search`}
		type="search"
		{value}
		oninput={(event) => oninput(event.currentTarget.value)}
	/>
</div>
