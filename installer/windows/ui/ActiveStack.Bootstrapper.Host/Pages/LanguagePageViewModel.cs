using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core.Localization;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// First wizard page (before the Hub): a fixed bilingual title/subtitle and
/// an extensible single-select list of language choices (English/Español).
/// The initial selection follows the resolved preselection (persisted
/// preference &gt; OS UI culture &gt; English — <see cref="LanguagePreselector"/>,
/// injected by the caller). Always advanceable: a language is selected by
/// construction (gui-language-page, L4).
/// </summary>
public sealed class LanguagePageViewModel : WizardPageViewModelBase
{
    private string _selectedLanguageId;

    public LanguagePageViewModel(string preselectedLanguageId)
        : base(UiStrings.Get("en", "page.language.title"), UiStrings.Get("en", "page.language.subtitle"), "en")
    {
        Choices =
        [
            new LanguageChoice("en", UiStrings.Get("en", "page.language.choice.english")),
            new LanguageChoice("es", UiStrings.Get("en", "page.language.choice.spanish"))
        ];

        _selectedLanguageId = Choices.Any(c => string.Equals(c.Id, preselectedLanguageId, StringComparison.OrdinalIgnoreCase))
            ? preselectedLanguageId
            : "en";
    }

    public ObservableCollection<LanguageChoice> Choices { get; }

    public string SelectedLanguageId
    {
        get => _selectedLanguageId;
        set
        {
            if (string.Equals(_selectedLanguageId, value, StringComparison.Ordinal))
            {
                return;
            }

            _selectedLanguageId = value;
            RaisePropertyChanged(nameof(SelectedLanguageId));
        }
    }

    public override bool CanAdvance => !string.IsNullOrEmpty(_selectedLanguageId);
}
