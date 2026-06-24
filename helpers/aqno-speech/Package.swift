// swift-tools-version:6.0
import PackageDescription

// aqno-speech — a tiny CLI that transcribes audio with Apple's SpeechAnalyzer
// (macOS 26+). The Go daemon spawns it as the optional "apple" STT engine.
let package = Package(
    name: "aqno-speech",
    platforms: [.macOS(.v15)],
    targets: [
        .executableTarget(name: "aqno-speech", path: "Sources/aqno-speech")
    ]
)
