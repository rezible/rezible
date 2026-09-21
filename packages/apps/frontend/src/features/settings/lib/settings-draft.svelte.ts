import { SvelteMap } from "svelte/reactivity";

// Settings API values are JSON records. Keep drafts separate from query-cache data.
const copy = <T>(value: T): T => JSON.parse(JSON.stringify(value));

export class SettingsDraft<T> {
	value: T = $state() as T;
	private saved: T = $state.raw() as T;
	dirty = $derived(JSON.stringify(this.value) !== JSON.stringify(this.saved));

	constructor(initial: T) {
		this.accept(initial);
	}

	receive(value: T) {
		if (!this.dirty) this.accept(value);
	}

	accept(value: T) {
		this.saved = copy(value);
		this.value = copy(value);
	}

	cancel = () => {
		this.value = copy(this.saved);
	};
}

export class SettingsDraftList<T extends { id: string }> {
	private forms = $state.raw(new SvelteMap<string, SettingsDraft<T>>());
	items = $derived(Array.from(this.forms.values(), (form) => form.value));
	dirty = $derived(Array.from(this.forms.values()).some((form) => form.dirty));

	receive(items: T[]) {
		const next = new SvelteMap<string, SettingsDraft<T>>();
		for (const item of items) {
			const form = this.forms.get(item.id);
			if (form) form.receive(item);
			next.set(item.id, form ?? new SettingsDraft(item));
		}
		for (const [id, form] of this.forms) {
			if (!next.has(id) && form.dirty) next.set(id, form);
		}
		this.forms = next;
	}

	get(id: string) {
		return this.forms.get(id)!;
	}

	accept(item: T) {
		const form = this.forms.get(item.id);
		if (form) form.accept(item);
		else this.forms.set(item.id, new SettingsDraft(item));
	}
}
