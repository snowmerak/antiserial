mod schema_v2;

use schema_v2::Payload;
use std::env;
use std::fs;

fn verify_golden(path: &str) -> Result<(), &'static str> {
    let data = fs::read(path).map_err(|_| "failed to read golden file")?;
    let mut offset = 0;
    let decoded = Payload::deserialize(&data, &mut offset)?;
    if offset != data.len() {
        return Err("did not consume full buffer");
    }
    if decoded.id != 1234567890 || decoded.uuid != "abc" || !decoded.active {
        return Err("field mismatch");
    }
    if decoded.tags.len() != 2 || decoded.tags[0] != "go" || decoded.tags[1] != "rust" {
        return Err("tags mismatch");
    }
    println!("Rust golden verify: PASSED");
    Ok(())
}

fn e2e_roundtrip() -> Result<(), &'static str> {
    let mut p = Payload::default();
    p.id = 1234567890;
    p.uuid = "abc";
    p.active = true;
    p.tags = vec!["go", "rust"];

    let mut serialized = Vec::new();
    p.serialize(&mut serialized)?;

    let expected_bytes = vec![
        0x0F,
        0xD2, 0x02, 0x96, 0x49, 0x00, 0x00, 0x00, 0x00,
        0x03, 0x00, 0x61, 0x62, 0x63,
        0x01,
        0x02, 0x00,
        0x02, 0x00, 0x67, 0x6F,
        0x04, 0x00, 0x72, 0x75, 0x73, 0x74,
    ];

    if serialized != expected_bytes {
        eprintln!("Rust wire format mismatch!");
        eprintln!("Got:  {:?}", serialized);
        eprintln!("Want: {:?}", expected_bytes);
        std::process::exit(1);
    }

    let mut offset = 0;
    let decoded = Payload::deserialize(&serialized, &mut offset)?;

    if offset != serialized.len() {
        eprintln!("Rust deserialize did not consume full buffer!");
        std::process::exit(1);
    }

    if decoded.id != p.id || decoded.uuid != p.uuid || decoded.active != p.active {
        eprintln!("Rust fields mismatch!");
        std::process::exit(1);
    }

    if decoded.tags.len() != 2 || decoded.tags[0] != "go" || decoded.tags[1] != "rust" {
        eprintln!("Rust tags mismatch!");
        std::process::exit(1);
    }

    println!("Rust E2E Verification: PASSED");
    Ok(())
}

fn main() -> Result<(), &'static str> {
    let args: Vec<String> = env::args().collect();
    if args.len() == 2 {
        return verify_golden(&args[1]);
    }
    e2e_roundtrip()
}