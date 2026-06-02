import { Payload, BufferWriter, BufferReader } from "./schema_v2.ts";

const p = new Payload();
p.id = 1234567890n;
p.uuid = "abc";
p.active = true;
p.tags = ["go", "rust"];

const writer = new BufferWriter();
p.serialize(writer);
const serialized = writer.getFinishedBytes();

const expectedBytes = new Uint8Array([
    0x0F,                                           // Bitmap
    0xD2, 0x02, 0x96, 0x49, 0x00, 0x00, 0x00, 0x00, // Id
    0x03, 0x00, 0x61, 0x62, 0x63,                   // Uuid
    0x01,                                           // Active
    0x02, 0x00,                                     // Tags Count
    0x02, 0x00, 0x67, 0x6F,                         // Tag 0 ("go")
    0x04, 0x00, 0x72, 0x75, 0x73, 0x74              // Tag 1 ("rust")
]);

const equals = (a: Uint8Array, b: Uint8Array): boolean => {
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i++) {
        if (a[i] !== b[i]) return false;
    }
    return true;
};

if (!equals(serialized, expectedBytes)) {
    console.error("TypeScript wire format mismatch!");
    console.error("Got: ", serialized);
    console.error("Want: ", expectedBytes);
    Deno.exit(1);
}

const reader = new BufferReader(serialized);
const decoded = new Payload();
decoded.deserialize(reader);

if (decoded.id !== p.id || decoded.uuid !== p.uuid || decoded.active !== p.active) {
    console.error("TypeScript fields mismatch!");
    Deno.exit(1);
}

if (decoded.tags.length !== 2 || decoded.tags[0] !== "go" || decoded.tags[1] !== "rust") {
    console.error("TypeScript tags mismatch!");
    Deno.exit(1);
}

console.log("TypeScript E2E Verification: PASSED");
