import { generateKeyPairSync } from "node:crypto";

const { privateKey, publicKey } = generateKeyPairSync("ed25519");
const { d } = privateKey.export({ format: "jwk" });
const { x } = publicKey.export({ format: "jwk" });
if (!d || !x) throw new Error("Unable to export Ed25519 document session keys");

console.log(JSON.stringify({
  seedHex: Buffer.from(d, "base64url").toString("hex"),
  publicKeyHex: Buffer.from(x, "base64url").toString("hex"),
}));
