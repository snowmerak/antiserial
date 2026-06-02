import sys
from schema_v2 import Payload

long_uuid = "a" * 65536
p = Payload()
p.uuid = long_uuid

buf = bytearray()
try:
    p.serialize(buf)
except ValueError as e:
    if "uint16" not in str(e):
        print(f"unexpected error: {e}")
        sys.exit(1)
    print("Python limits test: PASSED")
    sys.exit(0)

print("expected ValueError for string length > 65535")
sys.exit(1)