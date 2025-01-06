

const createDataState = () => {
	
	const setIncidentId = (id: string) => {
		console.log("incident", id);
	}
	
	return {
		setIncidentId,
	}
}
export const data = createDataState();