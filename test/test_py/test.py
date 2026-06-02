import sys
from schema_v2 import Payload

p = Payload()
p.id = 1234567890
p.uuid = "abc"
p.active = True
p.tags = ["go", "rust"]

buf = bytearray()
p.serialize(buf)
serialized = bytes(buf)

expected_bytes = bytes([
    0x0F,                                           # Bitmap
    0xD2, 0x02, 0x96, 0x49, 0x00, 0x00, 0x00, 0x00, # Id
    0x03, 0x00, 0x61, 0x62, 0x63,                   # Uuid
    0x01,                                           # Active
    0x02, 0x00,                                     # Tags Count
    0x02, 0x00, 0x67, 0x6F,                         # Tag 0 ("go")
    0x04, 0x00, 0x72, 0x75, 0x73, 0x74              # Tag 1 ("rust")
])

if serialized != expected_bytes:
    print("Python wire format mismatch!")
    print(f"Got:  {serialized.hex()}")
    print(f"Want: {expected_bytes.hex()}")
    sys.exit(1)

decoded = Payload()
offset = decoded.deserialize(serialized, 0)

if offset != len(serialized):
    print("Python deserialize did not consume full buffer!")
    sys.exit(1)

if decoded.id != p.id or decoded.uuid != p.uuid or decoded.active != p.active:
    print("Python fields mismatch!")
    sys.exit(1)

if len(decoded.tags) != 2 or decoded.tags[0] != "go" or decoded.tags[1] != "rust":
    print("Python tags mismatch!")
    sys.exit(1)

print("Python E2E Verification: PASSED")
