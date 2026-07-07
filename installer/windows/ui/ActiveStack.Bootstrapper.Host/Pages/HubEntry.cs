namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>One entry on the Hub page (D8, design.md).</summary>
public sealed record HubEntry(string Id, string Label, bool IsEnabled, string? Tooltip);
