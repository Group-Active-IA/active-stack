using ActiveStack.Bootstrapper.Host.Pages;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

/// <summary>
/// Covers the shared <c>DetailHeader</c> property threaded through
/// <see cref="WizardPageViewModelBase"/>'s <c>lang</c> constructor parameter
/// (gui-detail-panel, L5, design.md D4). Exercised via <see cref="HubPageViewModel"/>
/// since it's the simplest concrete page-VM (lang-only constructor).
/// </summary>
public sealed class WizardPageViewModelBaseTests
{
    [Fact]
    public void DetailHeader_DefaultsToEnglish()
    {
        var page = new HubPageViewModel();

        Assert.Equal("Details", page.DetailHeader);
    }

    [Fact]
    public void DetailHeader_ReflectsSpanishWhenLangIsEs()
    {
        var page = new HubPageViewModel("es");

        Assert.Equal("Detalle", page.DetailHeader);
    }
}
