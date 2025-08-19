fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_prost_build::configure()
        .build_server(true)
        .build_client(false)
        .file_descriptor_set_path("src/market_descriptor.bin")  // Add this
        .compile_protos(
            &["../proto/market.proto"],
            &["../proto/"],
        )?;

    println!("cargo:rerun-if-changed=../proto/market.proto");
    Ok(())
}