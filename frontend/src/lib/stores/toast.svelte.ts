import { type Snippet } from 'svelte';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
	id: string;
	type: ToastType;
	title: string;
	description?: string;
	duration?: number;
}

class ToastStore {
	toasts = $state<Toast[]>([]);

	add(toast: Omit<Toast, 'id'>) {
		const id = Math.random().toString(36).substring(2, 9);
		const newToast = { ...toast, id };
		this.toasts = [...this.toasts, newToast];

		if (toast.duration !== 0) {
			setTimeout(() => {
				this.remove(id);
			}, toast.duration || 5000);
		}
	}

	success(title: string, description?: string) {
		this.add({ type: 'success', title, description });
	}

	error(title: string, description?: string) {
		this.add({ type: 'error', title, description });
	}

	warning(title: string, description?: string) {
		this.add({ type: 'warning', title, description });
	}

	info(title: string, description?: string) {
		this.add({ type: 'info', title, description });
	}

	remove(id: string) {
		this.toasts = this.toasts.filter((t) => t.id !== id);
	}
}

export const toastStore = new ToastStore();
