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

    private void DispatchStatus(string message, string? stepId)
    {
        if (_window is null)
        {
            return;
        }

        _window.Dispatcher.Invoke(() => _window.ViewModel.ReportExternalProgress(message, stepId));
    }

    private void DispatchCompletion(bool success, string message, string? stepId)
    {
        if (_window is null)
        {
            return;
        }

        _window.Dispatcher.Invoke(() => _window.ViewModel.CompleteExternalInstall(success, message, stepId));
    }

    private void DispatchFailure(string title, string details, string? stepId)
    {
        if (_window is null)
        {
            return;
        }

        _window.Dispatcher.Invoke(() => _window.ViewModel.FailExternalInstall(title, details, stepId));
    }
}
