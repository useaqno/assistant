// Aqno desktop shell (Tauri 2).
//
// Responsibilities of the Rust layer:
//   1. Spawn the Go daemon (`aqnod`) as a managed sidecar on startup.
//   2. Wait for its `AQNOD_LISTENING <port>` handshake on stdout, then emit a
//      `daemon-ready` event so the SvelteKit front knows the base URL.
//   3. Expose `daemon_url` to the webview via IPC.
//   4. Tear the daemon down when the app exits.

use std::sync::Mutex;

use tauri::{Emitter, Manager, RunEvent};
use tauri_plugin_shell::process::{CommandChild, CommandEvent};
use tauri_plugin_shell::ShellExt;

const DAEMON_PORT: u16 = 8787;

#[derive(Default)]
struct DaemonState {
    port: Mutex<u16>,
    child: Mutex<Option<CommandChild>>,
}

#[derive(Clone, serde::Serialize)]
struct DaemonReady {
    url: String,
    port: u16,
}

/// Base URL the webview should use to reach the daemon.
#[tauri::command]
fn daemon_url(state: tauri::State<DaemonState>) -> String {
    let port = *state.port.lock().unwrap();
    format!("http://127.0.0.1:{}", port)
}

/// Whether this is a packaged (release) build. In `tauri dev` (debug, no .app
/// bundle) macOS TCC refuses microphone/speech access and aborts the process,
/// so the webview avoids capturing audio unless bundled.
#[tauri::command]
fn is_bundled() -> bool {
    !cfg!(debug_assertions)
}

pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(DaemonState::default())
        .invoke_handler(tauri::generate_handler![daemon_url, is_bundled])
        .setup(|app| {
            *app.state::<DaemonState>().port.lock().unwrap() = DAEMON_PORT;

            // Global push-to-talk hotkey (⌥Space): tell the webview to listen.
            #[cfg(desktop)]
            {
                use tauri_plugin_global_shortcut::{
                    Builder as ShortcutBuilder, GlobalShortcutExt, ShortcutState,
                };
                app.handle().plugin(
                    ShortcutBuilder::new()
                        .with_handler(|app, _shortcut, event| {
                            if event.state == ShortcutState::Pressed {
                                let _ = app.emit("voice-hotkey", ());
                            }
                        })
                        .build(),
                )?;
                let _ = app.global_shortcut().register("Alt+Space");
            }

            // Launch the bundled Go sidecar. Tauri resolves the platform binary
            // `binaries/aqnod-<target-triple>` declared in tauri.conf.json.
            let sidecar = app
                .shell()
                .sidecar("aqnod")
                .expect("failed to create `aqnod` sidecar command")
                .args(["--port", &DAEMON_PORT.to_string()]);

            let (mut rx, child) = sidecar.spawn().expect("failed to spawn aqnod");
            *app.state::<DaemonState>().child.lock().unwrap() = Some(child);

            let handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                while let Some(event) = rx.recv().await {
                    match event {
                        CommandEvent::Stdout(line) => {
                            let text = String::from_utf8_lossy(&line);
                            if let Some(rest) = text.trim().strip_prefix("AQNOD_LISTENING ") {
                                if let Ok(port) = rest.trim().parse::<u16>() {
                                    let _ = handle.emit(
                                        "daemon-ready",
                                        DaemonReady {
                                            url: format!("http://127.0.0.1:{}", port),
                                            port,
                                        },
                                    );
                                }
                            }
                        }
                        CommandEvent::Stderr(line) => {
                            eprintln!("[aqnod] {}", String::from_utf8_lossy(&line));
                        }
                        CommandEvent::Terminated(payload) => {
                            eprintln!("[aqnod] terminated: {:?}", payload);
                        }
                        _ => {}
                    }
                }
            });

            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building Aqno")
        .run(|app_handle, event| {
            // Make sure the daemon dies with the app.
            if let RunEvent::Exit = event {
                if let Some(child) = app_handle
                    .state::<DaemonState>()
                    .child
                    .lock()
                    .unwrap()
                    .take()
                {
                    let _ = child.kill();
                }
            }
        });
}
