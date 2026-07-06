namespace ActiveStack.Bootstrapper.Host;

public partial class MainWindow : System.Windows.Window
{
    private readonly MainWindowViewModel _viewModel;
    private readonly BootstrapperLaunchOptions _launchOptions;

    public MainWindow(string? enginePath = null)
    {
        InitializeComponent();

        _launchOptions = BootstrapperLaunchOptions.ReadFromEnvironment();
        BootstrapperTrace.Configure(_launchOptions.TraceEnabled);

        var resolvedEnginePath = ResolveEnginePath(enginePath);
        BootstrapperTrace.Write($"MainWindow ctor resolvedEnginePath={resolvedEnginePath}");
        _viewModel = new MainWindowViewModel(new ProcessInstallerEngineClient(resolvedEnginePath));
        DataContext = _viewModel;

        Loaded += OnLoadedAsync;
    }

    public MainWindowViewModel ViewModel => _viewModel;

    private async void OnLoadedAsync(object sender, System.Windows.RoutedEventArgs e)
    {
        await _viewModel.LoadAsync();

        if (!_launchOptions.AutoInstall)
        {
            return;
        }

        if (!string.IsNullOrWhiteSpace(_launchOptions.AssistantId))
        {
            _viewModel.SelectedAssistantId = _launchOptions.AssistantId;
        }

        if (!string.IsNullOrWhiteSpace(_launchOptions.InstallModeId))
        {
            _viewModel.SelectedInstallTypeId = _launchOptions.InstallModeId;
        }

        await _viewModel.StartInstallAsync();

        if (_launchOptions.AutoCloseWhenFinished)
        {
            Close();
        }
    }

    private async void InstallButton_Click(object sender, System.Windows.RoutedEventArgs e)
    {
        BootstrapperTrace.Write($"InstallButton_Click selectedAssistant={_viewModel.SelectedAssistantId ?? "<null>"} selectedMode={_viewModel.SelectedInstallTypeId ?? "<null>"} isEnabled={_viewModel.IsInstallActionEnabled}");
        await _viewModel.StartInstallAsync();
        BootstrapperTrace.Write($"InstallButton_Click completed isInstalling={_viewModel.IsInstalling} success={_viewModel.InstallSucceeded} currentStep={_viewModel.CurrentStepId ?? "<null>"}");
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
