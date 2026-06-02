import sys
from schema_nested import WithGeo

w = WithGeo()
w.id = 42

buf = bytearray()
w.serialize(buf)
serialized = bytes(buf)

want_len = 1 + 8 + 1
if len(serialized) != want_len:
    print(f"wire length: got {len(serialized)} want {want_len}")
    sys.exit(1)
if serialized[0] != 0x03:
    print(f"outer bitmap: got 0x{serialized[0]:02x} want 0x03")
    sys.exit(1)
if serialized[-1] != 0x00:
    print(f"nested geo bitmap: got 0x{serialized[-1]:02x} want 0x00")
    sys.exit(1)

decoded = WithGeo()
offset = decoded.deserialize(serialized, 0)
if offset != len(serialized) or decoded.id != 42:
    print("nested round-trip failed")
    sys.exit(1)

print("Python nested struct test: PASSED")