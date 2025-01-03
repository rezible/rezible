import type { ParamMatcher } from '@sveltejs/kit';

export const match = ((param?: string): param is (undefined | 'retrospective') => {
	return !param || param === 'retrospective';
}) satisfies ParamMatcher;