import { getContext, onDestroy, setContext } from "svelte";

export type Toast = {
	id: string;
	title: string;
	message: string;
	icon?: string;
};

export class ToastState {
	toasts = $state<Toast[]>([]);
	toastToTimeoutMap = new Map<string, ReturnType<typeof setTimeout>>();

	constructor() {
		onDestroy(() => {
			for (const timeout of this.toastToTimeoutMap.values()) {
				clearTimeout(timeout);
			}
			this.toastToTimeoutMap.clear();
		});
	}

	add(title: string, message: string, icon?: string, durationMs = 5000) {
		const id = crypto.randomUUID();
		this.toasts.push({ id, title, message, icon });

		const newTimeout = setTimeout(() => {
			this.remove(id);
		}, durationMs);
		this.toastToTimeoutMap.set(id, newTimeout);
		
		return id;
	}

	remove(id: string) {
		const timeout = this.toastToTimeoutMap.get(id);
		if (timeout) {
			clearTimeout(timeout);
			this.toastToTimeoutMap.delete(id);
		}
		this.toasts = this.toasts.filter((toast) => toast.id !== id);
	}
}

const TOAST_KEY = Symbol("TOAST_STATE");

export function setToastState() {
	return setContext(TOAST_KEY, new ToastState());
}

export function getToastState() {
	return getContext<ReturnType<typeof setToastState>>(TOAST_KEY);
}
