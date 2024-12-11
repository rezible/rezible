import type { HandleClientError } from '@sveltejs/kit';

export const handleError: HandleClientError = async ({ error, event, status, message }) => {	
    const errorId = crypto.randomUUID();
	const appError: App.Error = {status, message, errorId};
	console.error("handle error", error, event);
	return appError;
};