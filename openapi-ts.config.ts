import {createConfig as frontendConfig } from "./frontend/openapi-ts.config";
import {createConfig as documentsServerConfig } from "./documents-server/openapi-ts.config";

const input = {path: "/tmp/rezible-spec.yaml"};
const output = (dir: string) => ({path: `${dir}/src/lib/api/oapi.gen`});

export default await Promise.all([
	frontendConfig(input, output("frontend")),
	documentsServerConfig(input, output("documents-server"))
]);
