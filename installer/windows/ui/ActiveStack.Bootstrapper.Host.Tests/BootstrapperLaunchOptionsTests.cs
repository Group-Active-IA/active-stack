using ActiveStack.Bootstrapper.Host;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class BootstrapperLaunchOptionsTests
{
    [Fact]
    public void ReadFromEnvironment_ParsesAutomationFlagsAndSelections()
    {
        var variables = new Dictionary<string, string?>(StringComparer.OrdinalIgnoreCase)
        {
            ["ACTIVE_STACK_AUTO_INSTALL"] = "true",
            ["ACTIVE_STACK_AUTO_CLOSE"] = "1",
            ["ACTIVE_STACK_INSTALL_ASSISTANT"] = "codex",
            ["ACTIVE_STACK_INSTALL_MODE"] = "lite",
            ["ACTIVE_STACK_BOOTSTRAPPER_TRACE"] = "true"
        };

        var options = BootstrapperLaunchOptions.ReadFromEnvironment(name => variables.TryGetValue(name, out var value) ? value : null);

        Assert.True(options.AutoInstall);
        Assert.True(options.AutoCloseWhenFinished);
        Assert.True(options.TraceEnabled);
        Assert.Equal("codex", options.AssistantId);
        Assert.Equal("lite", options.InstallModeId);
    }

    [Fact]
    public void ReadFromEnvironment_DefaultsToDisabledFlags()
    {
        var options = BootstrapperLaunchOptions.ReadFromEnvironment(_ => null);

        Assert.False(options.AutoInstall);
        Assert.False(options.AutoCloseWhenFinished);
        Assert.False(options.TraceEnabled);
        Assert.Null(options.AssistantId);
        Assert.Null(options.InstallModeId);
    }

    [Fact]
    public void MainWindowXaml_UsesOneWayBindingForProgressValue()
    {
        var xamlPath = Path.GetFullPath(Path.Combine(
            AppContext.BaseDirectory,
            "..", "..", "..", "..",
            "ActiveStack.Bootstrapper.Host",
            "MainWindow.xaml"));

        var xaml = File.ReadAllText(xamlPath);

        Assert.Contains("Value=\"{Binding ProgressValue, Mode=OneWay}\"", xaml);
    }
}
