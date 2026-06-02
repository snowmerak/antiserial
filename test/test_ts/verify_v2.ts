import { Payload, BufferReader } from "./schema_v2.ts";

const path = Deno.args[0];
if (!path) {
    console.error("usage: verify_v2.ts <payload_v2.bin>");
    Deno.exit(1);
}

const data = await Deno.readFile(path);
const reader = new BufferReader(data);
const p = new Payload();
p.deserialize(reader);

if (p.id !== 1234567890n || p.uuid !== "abc" || !p.active) {
    console.error("fields mismatch", p);
    Deno.exit(1);
}
if (p.tags.length !== 2 || p.tags[0] !== "go" || p.tags[1] !== "rust") {
    console.error("tags mismatch", p.tags);
    Deno.exit(1);
}
console.log("TypeScript golden verify: PASSED");