using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Linq;
using System.Runtime.CompilerServices;
using ActiveStack.Bootstrapper.Core;

namespace ActiveStack.Bootstrapper.Host;

public sealed class MainWindowViewModel : INotifyPropertyChanged
{
    private readonly IInstallerEngineClient _engineClient;
    private Func<string, string, Task>? _externalInstallRunner;
    private bool _isLoading = true;
    private string? _selectedAssistantId;
    private string? _selectedInstallTypeId;
    private string? _errorTitle;
    private string? _errorDetails;
    private bool _isInstalling;
    private bool _installSucceeded;
    private string? _progressMessage;
    private string? _currentStepId;
    private string? _currentStepLabel;
    private double _progressValue;

    public MainWindowViewModel(IInstallerEngineClient engineClient)
    {
        _engineClient = engineClient;
        Title = "Install Active Stack";
        Subtitle = "Set up your AI coding workspace with a guided Windows installer.";
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public string Title { get; }

    public string Subtitle { get; }

    public bool IsLoading
    {
        get => _isLoading;
        private set
        {
            SetField(ref _isLoading, value);
            NotifyInstallActionStateChanged();
        }
    }

    public string? SelectedAssistantId
    {
        get => _selectedAssistantId;
        set
        {
            SetField(ref _selectedAssistantId, value);
            NotifyInstallActionStateChanged();
        }
    }

    public string? SelectedInstallTypeId
    {
        get => _selectedInstallTypeId;
        set
        {
            SetField(ref _selectedInstallTypeId, value);
            NotifyInstallActionStateChanged();
        }
    }

    public string? ErrorTitle
    {
        get => _errorTitle;
        private set
        {
            SetField(ref _errorTitle, value);
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(HasError)));
        }
    }

    public string? ErrorDetails
    {
        get => _errorDetails;
        private set
        {
            SetField(ref _errorDetails, value);
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(HasError)));
        }
    }

    public bool HasError => !string.IsNullOrWhiteSpace(ErrorTitle) || !string.IsNullOrWhiteSpace(ErrorDetails);

    public bool IsInstalling
    {
        get => _isInstalling;
        private set
        {
            SetField(ref _isInstalling, value);
            NotifyInstallActionStateChanged();
        }
    }

    public bool InstallSucceeded
    {
        get => _installSucceeded;
        private set => SetField(ref _installSucceeded, value);
    }

    public string? ProgressMessage
    {
        get => _progressMessage;
        private set => SetField(ref _progressMessage, value);
    }

    public string? CurrentStepId
    {
        get => _currentStepId;
        private set => SetField(ref _currentStepId, value);
    }

    public string? CurrentStepLabel
    {
        get => _currentStepLabel;
        private set => SetField(ref _currentStepLabel, value);
    }

    public double ProgressValue
    {
        get => _progressValue;
        private set => SetField(ref _progressValue, value);
    }

    public ObservableCollection<AssistantChoice> AssistantChoices { get; } = [];

    public ObservableCollection<InstallTypeChoice> InstallTypeChoices { get; } = [];

    public ObservableCollection<ComponentChoice> ForcedComponents { get; } = [];

    public ObservableCollection<ComponentChoice> CustomComponents { get; } = [];

    public ObservableCollection<string> RecentActivity { get; } = [];

    public bool IsInstallActionEnabled =>
        !IsLoading &&
        !IsInstalling &&
        !string.IsNullOrWhiteSpace(SelectedAssistantId) &&
        !string.IsNullOrWhiteSpace(SelectedInstallTypeId);

    public string InstallButtonText => IsInstalling ? "Installing..." : "Install now";

    public async Task LoadAsync(CancellationToken cancellationToken = default)
    {
        try
        {
            IsLoading = true;
            ErrorTitle = null;
            ErrorDetails = null;

            var state = await _engineClient.LoadSessionAsync(cancellationToken);

            ResetCollection(AssistantChoices, state.AssistantChoices);
            ResetCollection(InstallTypeChoices, state.InstallTypeChoices);
            ResetCollection(ForcedComponents, state.ForcedComponents);
            ResetCollection(CustomComponents, state.CustomComponents);

            SelectedAssistantId = state.DefaultAssistantId;
            SelectedInstallTypeId = state.RecommendedModeId;
        }
        catch (Exception ex)
        {
            ErrorTitle = "We couldn't prepare your setup yet.";
            ErrorDetails = ex.Message;
        }
        finally
        {
            IsLoading = false;
        }
    }

    public async Task StartInstallAsync(CancellationToken cancellationToken = default)
    {
        BootstrapperTrace.Write($"ViewModel.StartInstallAsync entered isInstalling={IsInstalling} selectedAssistant={SelectedAssistantId ?? "<null>"} selectedMode={SelectedInstallTypeId ?? "<null>"} externalRunner={_externalInstallRunner is not null}");
        if (IsInstalling)
        {
            BootstrapperTrace.Write("ViewModel.StartInstallAsync ignored because IsInstalling=true");
            return;
        }

        if (string.IsNullOrWhiteSpace(SelectedAssistantId) || string.IsNullOrWhiteSpace(SelectedInstallTypeId))
        {
            ErrorTitle = "We couldn't finish the installation.";
            ErrorDetails = "Select an assistant and install type before continuing.";
            InstallSucceeded = false;
            BootstrapperTrace.Write("ViewModel.StartInstallAsync validation failed due to missing selection");
            return;
        }

        try
        {
            IsInstalling = true;
            InstallSucceeded = false;
            ErrorTitle = null;
            ErrorDetails = null;
            ProgressMessage = "Starting installation.";
            CurrentStepId = "launch";
            CurrentStepLabel = "Installer";
            ProgressValue = 4;
            RecentActivity.Clear();
            AddActivity("Starting installation.");
            AddActivity("Launching installer engine.");

            if (_externalInstallRunner is not null)
            {
                BootstrapperTrace.Write("ViewModel.StartInstallAsync invoking external runner");
                await _externalInstallRunner(SelectedAssistantId, SelectedInstallTypeId);
                BootstrapperTrace.Write("ViewModel.StartInstallAsync external runner returned");
                return;
            }

            await foreach (var snapshot in _engineClient.RunInstallAsync(SelectedAssistantId, SelectedInstallTypeId, cancellationToken))
            {
                ApplySnapshot(snapshot);
            }
        }
        catch (Exception ex)
        {
            BootstrapperTrace.Write($"ViewModel.StartInstallAsync exception={ex}");
            ErrorTitle = "We couldn't finish the installation.";
            ErrorDetails = ex.Message;
            InstallSucceeded = false;
        }
        finally
        {
            if (_externalInstallRunner is null)
            {
                IsInstalling = false;
            }
            BootstrapperTrace.Write($"ViewModel.StartInstallAsync finally isInstalling={IsInstalling} success={InstallSucceeded} currentStep={CurrentStepId ?? "<null>"}");
        }
    }

    public void ConfigureExternalInstallRunner(Func<string, string, Task> externalInstallRunner)
    {
        _externalInstallRunner = externalInstallRunner;
    }

    public void ReportExternalProgress(string message, string? stepId = null)
    {
        BootstrapperTrace.Write($"ViewModel.ReportExternalProgress stepId={stepId ?? "<null>"} message={message}");
        if (!string.IsNullOrWhiteSpace(message))
        {
            ProgressMessage = message;
            AddActivity(message);
        }

        if (!string.IsNullOrWhiteSpace(stepId))
        {
            CurrentStepId = stepId;
            CurrentStepLabel = ToDisplayLabel(stepId);
        }
    }

    public void CompleteExternalInstall(bool success, string message, string? stepId = null)
    {
        BootstrapperTrace.Write($"ViewModel.CompleteExternalInstall success={success} stepId={stepId ?? "<null>"} message={message}");
        InstallSucceeded = success;
        IsInstalling = false;

        if (!string.IsNullOrWhiteSpace(message))
        {
            ProgressMessage = message;
            AddActivity(message);
        }

        if (!string.IsNullOrWhiteSpace(stepId))
        {
            CurrentStepId = stepId;
            CurrentStepLabel = ToDisplayLabel(stepId);
        }
    }

    private void AddActivity(string message)
    {
        if (string.IsNullOrWhiteSpace(message))
        {
            return;
        }

        var normalized = message.Trim();
        if (RecentActivity.Count > 0 && string.Equals(RecentActivity[^1], normalized, StringComparison.Ordinal))
        {
            return;
        }

        RecentActivity.Add(normalized);
        while (RecentActivity.Count > 8)
        {
            RecentActivity.RemoveAt(0);
        }
    }

    public void FailExternalInstall(string title, string details, string? stepId = null)
    {
        BootstrapperTrace.Write($"ViewModel.FailExternalInstall title={title} stepId={stepId ?? "<null>"} details={details}");
        ErrorTitle = title;
        ErrorDetails = details;
        InstallSucceeded = false;
        IsInstalling = false;

        if (!string.IsNullOrWhiteSpace(stepId))
        {
            CurrentStepId = stepId;
            CurrentStepLabel = ToDisplayLabel(stepId);
        }
    }

    private void ApplySnapshot(InstallProgressSnapshot snapshot)
    {
        if (!string.IsNullOrWhiteSpace(snapshot.StepId))
        {
            CurrentStepId = snapshot.StepId;
            CurrentStepLabel = ToDisplayLabel(snapshot.StepId);
        }

        var friendlyMessage = ToFriendlyMessage(snapshot);
        if (!string.IsNullOrWhiteSpace(friendlyMessage))
        {
            ProgressMessage = friendlyMessage;
            AddActivity(friendlyMessage);
        }

        ProgressValue = CalculateProgress(snapshot, ProgressValue);

        if (string.Equals(snapshot.Type, "install_finished", StringComparison.OrdinalIgnoreCase))
        {
            InstallSucceeded = snapshot.Success;
        }
    }

    private static double CalculateProgress(InstallProgressSnapshot snapshot, double current)
    {
        if (string.Equals(snapshot.Type, "install_finished", StringComparison.OrdinalIgnoreCase))
        {
            return 100;
        }

        if (string.Equals(snapshot.Type, "phase_started", StringComparison.OrdinalIgnoreCase))
        {
            return snapshot.Phase?.ToLowerInvariant() switch
            {
                "prepare" => Math.Max(current, 18),
                "apply" => Math.Max(current, 42),
                "rollback" => Math.Max(current, 70),
                _ => Math.Max(current, 8)
            };
        }

        if (string.Equals(snapshot.Type, "download_started", StringComparison.OrdinalIgnoreCase))
        {
            return Math.Min(88, Math.Max(current + 8, 52));
        }

        if (string.Equals(snapshot.Type, "step_started", StringComparison.OrdinalIgnoreCase))
        {
            return Math.Min(90, current + 7);
        }

        if (string.Equals(snapshot.Type, "step_succeeded", StringComparison.OrdinalIgnoreCase) ||
            string.Equals(snapshot.Type, "download_finished", StringComparison.OrdinalIgnoreCase))
        {
            return Math.Min(96, current + 5);
        }

        return current;
    }

    private static string ToFriendlyMessage(InstallProgressSnapshot snapshot)
    {
        var label = ToDisplayLabel(snapshot.StepId);

        return snapshot.Type.ToLowerInvariant() switch
        {
            "phase_started" => snapshot.Message ?? "Running installation.",
            "step_started" when !string.IsNullOrWhiteSpace(label) => $"Installing {label}.",
            "step_succeeded" when !string.IsNullOrWhiteSpace(label) => $"Installed {label}.",
            "step_failed" when !string.IsNullOrWhiteSpace(label) => $"Failed to install {label}.",
            "download_started" when !string.IsNullOrWhiteSpace(label) => $"Downloading {label}.",
            "download_finished" when !string.IsNullOrWhiteSpace(label) => $"Downloaded {label}.",
            "install_finished" => snapshot.Message ?? (snapshot.Success ? "Installation finished successfully." : "Installation failed."),
            _ => snapshot.Message ?? snapshot.Details ?? "Running installation."
        };
    }

    private static string? ToDisplayLabel(string? stepId)
    {
        if (string.IsNullOrWhiteSpace(stepId))
        {
            return null;
        }

        var raw = stepId.Trim();
        var segment = raw.Contains(':', StringComparison.Ordinal) ? raw[(raw.LastIndexOf(':') + 1)..] : raw;
        if (string.Equals(segment, "openspec", StringComparison.OrdinalIgnoreCase))
        {
            return "OpenSpec";
        }
        if (string.Equals(segment, "engram", StringComparison.OrdinalIgnoreCase))
        {
            return "Engram";
        }
        segment = segment.Replace("-", " ", StringComparison.Ordinal).Replace("_", " ", StringComparison.Ordinal);
        if (string.IsNullOrWhiteSpace(segment))
        {
            return raw;
        }

        return string.Join(" ", segment
            .Split(' ', StringSplitOptions.RemoveEmptyEntries)
            .Select(static part =>
            {
                if (part.Length <= 3)
                {
                    return part.ToUpperInvariant();
                }

                return char.ToUpperInvariant(part[0]) + part[1..];
            }));
    }

    private void NotifyInstallActionStateChanged()
    {
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(IsInstallActionEnabled)));
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(InstallButtonText)));
    }

    private void ResetCollection<T>(ObservableCollection<T> target, IReadOnlyList<T> values)
    {
        target.Clear();
        foreach (var value in values)
        {
            target.Add(value);
        }
    }

    private void SetField<T>(ref T field, T value, [CallerMemberName] string? propertyName = null)
    {
        if (EqualityComparer<T>.Default.Equals(field, value))
        {
            return;
        }

        field = value;
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
    }
}
