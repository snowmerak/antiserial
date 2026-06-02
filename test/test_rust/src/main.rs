mod schema_v2;

use schema_v2::Payload;

fn main() -> Result<(), &'static str> {
    let mut p = Payload::default();
    p.id = 1234567890;
    p.uuid = "abc";
    p.active = true;
    p.tags = vec!["go", "rust"];

    let mut serialized = Vec::new();
    p.serialize(&mut serialized);

    let expected_bytes = vec![
        0x0F,                                           // Bitmap
        0xD2, 0x02, 0x96, 0x49, 0x00, 0x00, 0x00, 0x00, // Id
        0x03, 0x00, 0x61, 0x62, 0x63,                   // Uuid
        0x01,                                           // Active
        0x02, 0x00,                                     // Tags Count
        0x02, 0x00, 0x67, 0x6F,                         // Tag 0 ("go")
        0x04, 0x00, 0x72, 0x75, 0x73, 0x74,             // Tag 1 ("rust")
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
