namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Mutable model shared across every Starter-flow page view-model: the
/// chosen starter id, the target project path, and the agents to wire it
/// for (defaulting to claude + opencode per the plan). Created once when the
/// Starter flow starts (from the Hub) and threaded through
/// <see cref="ShellViewModel"/> and the page view-models, mirroring
/// <see cref="InstallSelection"/> (D4, design.md). <see cref="StarterFlow"/>
/// only ever reads it — it never mutates it.
/// </summary>
public sealed class StarterSelection
{
    public string StarterId { get; set; } = string.Empty;

    public string ProjectPath { get; set; } = string.Empty;

    public List<string> Agents { get; set; } = ["claude", "opencode"];
}
