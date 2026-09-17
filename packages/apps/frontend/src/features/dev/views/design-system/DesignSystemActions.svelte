<script lang="ts">
	import { mode, resetMode, setMode, userPrefersMode } from "mode-watcher";
	import * as Tabs from "$components/ui/tabs";

	const themeMode = $derived(userPrefersMode.current);
	const effectiveMode = $derived(mode.current);

	const chooseTheme = (value: string) => {
		if (value === "system") resetMode();
		else if (value === "light" || value === "dark") setMode(value);
	};
</script>

<div class="flex items-center gap-3">
	<span class="hidden text-xs text-muted-foreground sm:inline">Effective: {effectiveMode}</span>
	<Tabs.Root value={themeMode} onValueChange={chooseTheme}>
		<Tabs.List aria-label="Colour theme">
			<Tabs.Trigger value="light">Light</Tabs.Trigger>
			<Tabs.Trigger value="dark">Dark</Tabs.Trigger>
			<Tabs.Trigger value="system">System</Tabs.Trigger>
		</Tabs.List>
	</Tabs.Root>
</div>
