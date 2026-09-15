import getPort, { portNumbers } from "get-port";

const ranges = {
    backend: 20000,
    frontend: 21000,
    documents: 22000,
};

async function printPort(service: string) {
  const start = ranges[service];
  if (!start) throw new Error("Expected service: backend, frontend, or documents");

  const end = start + 999;
  const port = await getPort({ port: portNumbers(start, end) });
  if (port < start || port > end) throw new Error(`No available ports for ${service}`);
  console.log(port);
}

await printPort(process.argv[2]);
