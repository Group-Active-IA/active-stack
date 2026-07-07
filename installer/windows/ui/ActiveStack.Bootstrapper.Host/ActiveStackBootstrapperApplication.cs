using System.IO;
using System.Windows;
using WixToolset.BootstrapperApplicationApi;

namespace ActiveStack.Bootstrapper.Host;

internal sealed class ActiveStackBootstrapperApplication : BootstrapperApplication
{
    private MainWindow? _window;

    protected override void Run()
    {
        BootstrapperTrace.Write("BA.Run entered");
        engine.CloseSplashScreen();

        var app = new App
        {
            ShutdownMode = ShutdownMode.OnMainWindowClose
        };

        _window = new MainWindow(Path.Combine(AppContext.BaseDirectory, "active-stack.exe"));
        _window.Loaded += (_, _) => engine.Detect();
        _window.Closed += (_, _) =>
        {
            var exitCode = _window.ViewModel.InstallSucceeded ? 0 : 1;
            engine.Quit(exitCode);
        };

        this.DetectComplete += (_, _) =>
        {
            BootstrapperTrace.Write("BA.DetectComplete");
            DispatchStatus("Setup is ready to install.", "detect");
        };
        this.Error += OnError;

        app.Run(_window);
    }

    private void OnError(object? sender, EventArgs e)
    {
        if (e is not WixToolset.BootstrapperApplicationApi.ErrorEventArgs args)
        {
            return;
        }

        BootstrapperTrace.Write($"BA.OnError packageId={args.PackageId} message={args.ErrorMessage}");

        DispatchFailure("We couldn't finish the installation.", args.ErrorMessage, args.PackageId);
    }

    /// <summary>
    /// Traces Burn detection status. The old dashboard (<c>MainWindowViewModel</c>)
    /// surfaced this as free-form status text on its single screen; the
    /// wizard has no such single text slot (each page owns its own state),
    /// so this is now trace-only — Burn's detect/apply lifecycle never
    /// drives the actual install (that stays on the Process/ArgumentList
    /// path via <see cref="ProcessInstallerEngineClient"/>, unchanged).
    /// </summary>
    private static void DispatchStatus(string message, string? stepId) =>
        BootstrapperTrace.Write($"BA status message={message} stepId={stepId ?? "<null>"}");

    private static void DispatchFailure(string title, string details, string? stepId) =>
        BootstrapperTrace.Write($"BA failure title={title} stepId={stepId ?? "<null>"} details={details}");
}
