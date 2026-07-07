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
        _viewModel = new ShellViewModel(new ProcessInstallerEngineClient(resolvedEnginePath), new WpfFolderPicker());
        DataContext = _viewModel;

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
    /// Headless/CI automation path (ACTIVE_STACK_AUTO_INSTALL): selects
    /// "Install" on the Hub, applies the assistant/mode overrides from
    /// launch options as each relevant page is reached, advances through
    /// the wizard, and confirms on Review. There is no longer a single
    /// "start install" call to preserve verbatim — the wizard replaces the
    /// old one-screen dashboard — so this walks the same
    /// <see cref="ShellViewModel"/> the interactive user drives.
    /// </summary>
    private async Task RunAutoInstallAsync()
    {
        if (_viewModel.CurrentPage is HubPageViewModel hub)
        {
            hub.SelectedEntryId = "install";
        }

        while (_viewModel.CurrentPage is not ReviewPageViewModel)
        {
            ApplyAutomationOverrides();

            if (!_viewModel.PrimaryEnabled)
            {
                BootstrapperTrace.Write("MainWindow auto-install stalled: current page cannot advance.");
                return;
            }

            await _viewModel.AdvanceAsync();
        }

        await _viewModel.AdvanceAsync();
    }

    private void ApplyAutomationOverrides()
    {
        if (_viewModel.CurrentPage is AssistantsPageViewModel assistants &&
            !string.IsNullOrWhiteSpace(_launchOptions.AssistantId))
        {
            foreach (var choice in assistants.Choices)
            {
                choice.IsSelected = string.Equals(choice.Id, _launchOptions.AssistantId, System.StringComparison.OrdinalIgnoreCase);
            }
        }

        if (_viewModel.CurrentPage is InstallTypePageViewModel installType &&
            !string.IsNullOrWhiteSpace(_launchOptions.InstallModeId))
        {
            installType.SelectedId = _launchOptions.InstallModeId!;
        }
    }

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
