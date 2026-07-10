using System.Globalization;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages;
using ActiveStack.Bootstrapper.Host.Pages.Install;

namespace ActiveStack.Bootstrapper.Host;

public partial class MainWindow : System.Windows.Window
{
    private readonly ShellViewModel _viewModel;
    private readonly BootstrapperLaunchOptions _launchOptions;

    public MainWindow(string? enginePath = null)
    {
        InitializeComponent();

        _launchOptions = BootstrapperLaunchOptions.ReadFromEnvironment();
        BootstrapperTrace.Configure(_launchOptions.TraceEnabled);

        var resolvedEnginePath = ResolveEnginePath(enginePath);
        BootstrapperTrace.Write($"MainWindow ctor resolvedEnginePath={resolvedEnginePath}");

        var homeDir = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
        // ACTIVE_STACK_UI_LANG (automation) > persisted preference > OS UI culture > English.
        var persistedOrOverride = _launchOptions.LanguageOverride ?? LanguagePreference.Load(homeDir);
        var initialLanguage = LanguagePreselector.Resolve(persistedOrOverride, static () => CultureInfo.CurrentUICulture.Name);

        _viewModel = new ShellViewModel(new ProcessInstallerEngineClient(resolvedEnginePath), new WpfFolderPicker(), initialLanguage);
        DataContext = _viewModel;
        // Finish on the Complete page (D2a, design.md): ShellViewModel stays
        // WPF-free and raises CloseRequested instead of owning the Window.
        _viewModel.CloseRequested += (_, _) => Close();

        Loaded += OnLoadedAsync;
    }

    public ShellViewModel ViewModel => _viewModel;

    private async void OnLoadedAsync(object sender, System.Windows.RoutedEventArgs e)
    {
        if (!_launchOptions.AutoInstall)
        {
            return;
        }

        BootstrapperTrace.Write("MainWindow auto-install requested; driving the wizard to Review, then confirming.");
        await RunAutoInstallAsync();

        if (_launchOptions.AutoCloseWhenFinished)
        {
            Close();
        }
    }

    /// <summary>
    /// Headless/CI automation path (ACTIVE_STACK_AUTO_INSTALL). Delegates to
    /// <see cref="AutoInstallDriver"/> (no WPF dependency, unit-testable):
    /// traverses the preselected Language page, selects "Install" on the
    /// Hub, applies the assistant/mode overrides from launch options as
    /// each relevant page is reached, advances through the wizard, and
    /// confirms on Review.
    /// </summary>
    private Task RunAutoInstallAsync() =>
        AutoInstallDriver.RunAsync(_viewModel, _launchOptions.AssistantId, _launchOptions.InstallModeId, BootstrapperTrace.Write);

    private void BackButton_Click(object sender, System.Windows.RoutedEventArgs e)
    {
        BootstrapperTrace.Write("BackButton_Click");
        _viewModel.GoBack();
    }

    private async void PrimaryButton_Click(object sender, System.Windows.RoutedEventArgs e)
    {
        BootstrapperTrace.Write($"PrimaryButton_Click label={_viewModel.PrimaryLabel} enabled={_viewModel.PrimaryEnabled}");
        await _viewModel.AdvanceAsync();
        BootstrapperTrace.Write("PrimaryButton_Click completed");
    }

    private void CancelButton_Click(object sender, System.Windows.RoutedEventArgs e)
    {
        BootstrapperTrace.Write("CancelButton_Click");
        Close();
    }

    private static string ResolveEnginePath(string? enginePath)
    {
        if (!string.IsNullOrWhiteSpace(enginePath))
        {
            return enginePath;
        }

        var environmentOverride = Environment.GetEnvironmentVariable("ACTIVE_STACK_ENGINE_PATH");
        if (!string.IsNullOrWhiteSpace(environmentOverride))
        {
            return environmentOverride;
        }

        var localPayload = System.IO.Path.Combine(AppContext.BaseDirectory, "active-stack.exe");
        if (System.IO.File.Exists(localPayload))
        {
            return localPayload;
        }

        return System.IO.Path.GetFullPath(System.IO.Path.Combine(AppContext.BaseDirectory, "..", "..", "..", "..", "payload", "active-stack.exe"));
    }
}
