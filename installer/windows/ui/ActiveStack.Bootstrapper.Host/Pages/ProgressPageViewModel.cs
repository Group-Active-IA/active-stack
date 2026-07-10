using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core.Localization;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Generic Installing page: renders any <see cref="IAsyncEnumerable{T}"/> of
/// <see cref="InstallProgressSnapshot"/> as phase-based progress, friendly
/// step messages, and an activity log (D4, design.md — this is what lets C5
/// reuse it for starter/uninstall streams). <c>ApplySnapshot</c>,
/// <c>CalculateProgress</c>, <c>ToFriendlyMessage</c>, and
/// <c>ToDisplayLabel</c> are moved verbatim from the former
/// <c>MainWindowViewModel</c> (D6, design.md); their behavior is pinned by
/// the migrated <c>ProgressPageViewModelTests</c>. Has no footer nav — the
/// shell hides Back/Next/Cancel while this page is current.
/// </summary>
public sealed class ProgressPageViewModel : WizardPageViewModelBase
{
    private readonly string _lang;
    private string? _progressMessage;
    private string? _currentStepId;
    private string? _currentStepLabel;
    private double _progressValue;
    private bool _installSucceeded;
    private bool _isFinished;

    public ProgressPageViewModel(string lang = "en")
        : base(UiStrings.Get(lang, "page.installing.title"), UiStrings.Get(lang, "page.installing.subtitle"), lang)
    {
        _lang = lang;
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

    public bool InstallSucceeded
    {
        get => _installSucceeded;
        private set => SetField(ref _installSucceeded, value);
    }

    /// <summary>True once a terminal (<c>install_finished</c>) snapshot has been applied.</summary>
    public bool IsFinished
    {
        get => _isFinished;
        private set => SetField(ref _isFinished, value);
    }

    /// <summary>True if any <c>step_degraded</c> (best-effort) event was seen (C-32).</summary>
    public bool HadDegradedSteps { get; private set; }

    /// <summary>True if the pipeline entered its <c>rollback</c> phase.</summary>
    public bool HadRollback { get; private set; }

    /// <summary>The final (<c>install_finished</c>) snapshot, for the Complete page's state mapping.</summary>
    public InstallProgressSnapshot? TerminalSnapshot { get; private set; }

    public ObservableCollection<string> RecentActivity { get; } = [];

    /// <summary>The Progress page has no footer nav — advancing is terminal-driven, not user-driven.</summary>
    public override bool CanAdvance => false;

    /// <summary>
    /// Consumes the given install-progress stream end to end, applying each
    /// snapshot as it arrives.
    /// </summary>
    public async Task ConsumeAsync(IAsyncEnumerable<InstallProgressSnapshot> stream, CancellationToken cancellationToken = default)
    {
        await foreach (var snapshot in stream.WithCancellation(cancellationToken))
        {
            ApplySnapshot(snapshot);
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

        if (string.Equals(snapshot.Type, "step_degraded", StringComparison.OrdinalIgnoreCase))
        {
            HadDegradedSteps = true;
        }

        if (string.Equals(snapshot.Type, "phase_started", StringComparison.OrdinalIgnoreCase) &&
            string.Equals(snapshot.Phase, "rollback", StringComparison.OrdinalIgnoreCase))
        {
            HadRollback = true;
        }

        if (string.Equals(snapshot.Type, "install_finished", StringComparison.OrdinalIgnoreCase))
        {
            InstallSucceeded = snapshot.Success;
            IsFinished = true;
            TerminalSnapshot = snapshot;
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

    /// <summary>
    /// Maps a snapshot to the user-facing status line. The messages this
    /// page OWNS (step/download progress, defaults) come from
    /// <see cref="UiStrings"/> in the active language; engine-supplied
    /// <c>phase_started</c>/<c>install_finished</c> messages arrive already
    /// localized (L1) and are shown verbatim (gui-language-page, L4).
    /// </summary>
    private string ToFriendlyMessage(InstallProgressSnapshot snapshot)
    {
        var label = ToDisplayLabel(snapshot.StepId);

        return snapshot.Type.ToLowerInvariant() switch
        {
            "phase_started" => snapshot.Message ?? UiStrings.Get(_lang, "progress.running_default"),
            "step_started" when !string.IsNullOrWhiteSpace(label) => string.Format(UiStrings.Get(_lang, "progress.installing_fmt"), label),
            "step_succeeded" when !string.IsNullOrWhiteSpace(label) => string.Format(UiStrings.Get(_lang, "progress.installed_fmt"), label),
            "step_failed" when !string.IsNullOrWhiteSpace(label) => string.Format(UiStrings.Get(_lang, "progress.failed_fmt"), label),
            "download_started" when !string.IsNullOrWhiteSpace(label) => string.Format(UiStrings.Get(_lang, "progress.downloading_fmt"), label),
            "download_finished" when !string.IsNullOrWhiteSpace(label) => string.Format(UiStrings.Get(_lang, "progress.downloaded_fmt"), label),
            "install_finished" => snapshot.Message ?? UiStrings.Get(_lang, snapshot.Success ? "progress.finished_success_default" : "progress.finished_failed_default"),
            _ => snapshot.Message ?? snapshot.Details ?? UiStrings.Get(_lang, "progress.running_default")
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
}
