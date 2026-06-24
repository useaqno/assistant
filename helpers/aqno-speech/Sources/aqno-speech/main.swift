// aqno-speech — Apple SpeechAnalyzer STT helper for the Aqno daemon.
//
// Usage:
//   aqno-speech --probe              → prints "ok" (macOS 26+) or "unsupported"
//   aqno-speech [--lang pt-BR]       → reads a WAV from stdin, prints {"text":…}
//
// The Go daemon (internal/voice, darwin) spawns this when the user selects the
// "apple" STT engine. Gated to macOS 26 via @available; on older systems
// --probe reports "unsupported" and the daemon falls back to whisper.cpp.
@preconcurrency import AVFoundation
import Foundation
import Speech

func argValue(_ name: String) -> String? {
    let args = CommandLine.arguments
    guard let i = args.firstIndex(of: name), i + 1 < args.count else { return nil }
    return args[i + 1]
}

func fail(_ message: String, code: Int32 = 1) -> Never {
    FileHandle.standardError.write(Data((message + "\n").utf8))
    exit(code)
}

func makeError(_ code: Int, _ message: String) -> NSError {
    NSError(domain: "aqno-speech", code: code, userInfo: [NSLocalizedDescriptionKey: message])
}

@available(macOS 26, *)
func ensureModel(_ transcriber: SpeechTranscriber, locale: Locale) async throws {
    if let request = try await AssetInventory.assetInstallationRequest(supporting: [transcriber]) {
        try await request.downloadAndInstall()
    }
    let supported = await SpeechTranscriber.supportedLocales
    let isSupported = supported.contains { $0.identifier(.bcp47) == locale.identifier(.bcp47) }
    guard isSupported else { return }
    let reserved = await AssetInventory.reservedLocales
    if !reserved.contains(where: { $0.identifier(.bcp47) == locale.identifier(.bcp47) }) {
        try await AssetInventory.reserve(locale: locale)
    }
}

@available(macOS 26, *)
func transcribe(wav url: URL, locale: Locale) async throws -> String {
    let transcriber = SpeechTranscriber(
        locale: locale, transcriptionOptions: [], reportingOptions: [], attributeOptions: [])
    try await ensureModel(transcriber, locale: locale)

    let analyzer = SpeechAnalyzer(modules: [transcriber])
    guard let format = await SpeechAnalyzer.bestAvailableAudioFormat(compatibleWith: [transcriber])
    else {
        throw makeError(10, "no compatible audio format")
    }

    // Read the whole WAV into a buffer, then convert to the analyzer's format.
    let file = try AVAudioFile(forReading: url)
    let inFormat = file.processingFormat
    guard
        let inBuffer = AVAudioPCMBuffer(
            pcmFormat: inFormat, frameCapacity: AVAudioFrameCount(file.length))
    else {
        throw makeError(11, "alloc input buffer")
    }
    try file.read(into: inBuffer)

    guard let converter = AVAudioConverter(from: inFormat, to: format) else {
        throw makeError(12, "no converter")
    }
    let ratio = format.sampleRate / inFormat.sampleRate
    let capacity = AVAudioFrameCount(Double(inBuffer.frameLength) * ratio) + 4096
    guard let outBuffer = AVAudioPCMBuffer(pcmFormat: format, frameCapacity: capacity) else {
        throw makeError(13, "alloc output buffer")
    }
    let source = SingleShotInput(buffer: inBuffer)
    var convError: NSError?
    converter.convert(to: outBuffer, error: &convError) { _, status in
        source.next(status)
    }
    if let convError { throw convError }

    // Collect final results while streaming the single converted buffer.
    let collector = Task { () -> String in
        var finalText = AttributedString("")
        for try await result in transcriber.results where result.isFinal {
            finalText += result.text
        }
        return String(finalText.characters)
    }

    let (stream, continuation) = AsyncStream<AnalyzerInput>.makeStream()
    try await analyzer.start(inputSequence: stream)
    continuation.yield(AnalyzerInput(buffer: outBuffer))
    continuation.finish()
    try await analyzer.finalizeAndFinishThroughEndOfInput()
    return try await collector.value
}

/// SingleShotInput feeds one buffer to AVAudioConverter then signals end.
final class SingleShotInput {
    private let buffer: AVAudioPCMBuffer
    private var consumed = false
    init(buffer: AVAudioPCMBuffer) { self.buffer = buffer }
    func next(_ status: UnsafeMutablePointer<AVAudioConverterInputStatus>) -> AVAudioBuffer? {
        if consumed {
            status.pointee = .noDataNow
            return nil
        }
        consumed = true
        status.pointee = .haveData
        return buffer
    }
}

@available(macOS 26, *)
func run() async {
    let locale = Locale(identifier: argValue("--lang") ?? "pt-BR")
    let data = FileHandle.standardInput.readDataToEndOfFile()
    guard !data.isEmpty else { fail("empty audio on stdin") }
    let tmp = URL(fileURLWithPath: NSTemporaryDirectory())
        .appendingPathComponent("aqno-\(UUID().uuidString).wav")
    do {
        try data.write(to: tmp)
        defer { try? FileManager.default.removeItem(at: tmp) }
        let text = try await transcribe(wav: tmp, locale: locale)
        let json = try JSONSerialization.data(withJSONObject: ["text": text])
        FileHandle.standardOutput.write(json)
    } catch {
        fail("transcription failed: \(error.localizedDescription)")
    }
}

// --probe: report availability without touching macOS-26 symbols.
if CommandLine.arguments.contains("--probe") {
    if #available(macOS 26, *) {
        print("ok")
    } else {
        print("unsupported")
    }
    exit(0)
}

if #available(macOS 26, *) {
    await run()
} else {
    fail("unsupported: SpeechAnalyzer requires macOS 26+", code: 2)
}
