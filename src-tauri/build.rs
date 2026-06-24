fn main() {
    // Embed Info.plist (mic/speech usage descriptions) directly into the macOS
    // executable so TCC accepts microphone/speech access even for the bare
    // `tauri dev` binary (which is not a .app bundle). Without this, invoking
    // SpeechRecognition/getUserMedia in the WKWebView aborts the process.
    #[cfg(target_os = "macos")]
    {
        if let Ok(path) = std::path::Path::new("Info.plist").canonicalize() {
            println!("cargo:rerun-if-changed=Info.plist");
            println!(
                "cargo:rustc-link-arg=-Wl,-sectcreate,__TEXT,__info_plist,{}",
                path.display()
            );
        }
    }

    tauri_build::build()
}
