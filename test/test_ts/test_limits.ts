import { Payload, BufferWriter } from "./schema_v2.ts";

const p = new Payload();
p.uuid = "a".repeat(65536);

const writer = new BufferWriter();
try {
    p.serialize(writer);
    console.error("expected Error for string length > 65535");
    Deno.exit(1);
} catch (e) {
    if (!(e instanceof Error) || !e.message.includes("uint16")) {
        console.error("unexpected error:", e);
        Deno.exit(1);
    }
}

console.log("TypeScript limits test: PASSED");