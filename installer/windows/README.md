# Windows bundle skeleton

This folder contains the first WiX Burn base for the native Windows installer.

The release target is a single `.exe` bundle. The graphical shell will package an internal bootstrap executable and delegate install logic to the existing Go engine through the `active-stack windows` commands:

- `active-stack windows detect`
- `active-stack windows options`
- `active-stack windows install`
- `active-stack windows uninstall`

Current scope of this block:

- define the WiX bundle project
- declare a bootstrapper application base
- add a custom theme placeholder
- document how the bundle will call the headless engine

This is not the final production installer yet. It is the bundle skeleton that the next blocks will complete with:

- packaged payload generation
- richer progress and download binding
- final copy and branded visuals

## Local build flow

The bundle payload is the real `active-stack.exe` binary built from `./cmd/active-stack`.

Use:

`powershell -ExecutionPolicy Bypass -File .\installer\windows\build.ps1`

Requirement:

- .NET 8 SDK installed locally for `dotnet build`

What the script does:

- builds `payload\active-stack.exe`
- builds the WiX Burn bundle with `dotnet build`
- copies the resulting installer to `dist\ActiveStack-Setup-<version>.exe`

For staging-only validation without invoking WiX:

`powershell -ExecutionPolicy Bypass -File .\installer\windows\build.ps1 -SkipBundleBuild`
