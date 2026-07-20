import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]): string {
	return twMerge(clsx(inputs));
}

export type WithElementRef<T extends Record<string, any>> = T & {
	ref?: Element | null;
};

export type WithoutChildrenOrChild<T extends Record<string, any>> = Omit<
	T,
	'children' | 'child'
>;
