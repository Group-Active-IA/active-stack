namespace ActiveStack.Bootstrapper.Host;

public partial class App : System.Windows.Application
{
    public App()
    {
        // Both hosts (Burn BA and --standalone) construct App directly instead of
        // going through the WPF-generated Main, so App.xaml — and with it the
        // PageTemplates.xaml merged dictionary that holds every wizard DataTemplate —
        // never loads unless we do it here. Without this, ContentControl falls back
        // to ToString() and pages render as raw type names.
        InitializeComponent();

        // D5, design.md: secondary safety net only — ShellViewModel's
        // targeted try/catch (D1) is the primary mechanism that turns a
        // stream/engine failure into the visible CompleteState.Error page.
        // This handler exists so a fault D1 does NOT anticipate still
        // degrades to something visible (logged + marked handled) instead
        // of WPF's default: a silent process kill.
        DispatcherUnhandledException += (_, e) =>
        {
            BootstrapperTrace.Write($"App.DispatcherUnhandledException: {e.Exception}");
            e.Handled = true;
        };

        AppDomain.CurrentDomain.UnhandledException += (_, e) =>
            BootstrapperTrace.Write($"AppDomain.UnhandledException (terminating={e.IsTerminating}): {e.ExceptionObject}");

        TaskScheduler.UnobservedTaskException += (_, e) =>
        {
            BootstrapperTrace.Write($"TaskScheduler.UnobservedTaskException: {e.Exception}");
            e.SetObserved();
        };
    }
}
