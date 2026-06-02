"""Decode a golden payload_v2.bin and verify field values."""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "test_py"))
from schema_v2 import Payload  # noqa: E402


def main() -> None:
    if len(sys.argv) != 2:
        print("usage: verify_v2.py <payload_v2.bin>")
        sys.exit(1)
    data = Path(sys.argv[1]).read_bytes()
    p = Payload()
    offset = p.deserialize(data, 0)
    if offset != len(data):
        print(f"consumed {offset} of {len(data)}")
        sys.exit(1)
    if p.id != 1234567890 or p.uuid != "abc" or not p.active:
        print(f"fields mismatch: id={p.id} uuid={p.uuid} active={p.active}")
        sys.exit(1)
    if len(p.tags) != 2 or p.tags[0] != "go" or p.tags[1] != "rust":
        print(f"tags mismatch: {p.tags}")
        sys.exit(1)
    print("Python golden verify: PASSED")


if __name__ == "__main__":
    main()