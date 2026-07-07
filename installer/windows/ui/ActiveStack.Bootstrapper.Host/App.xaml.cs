namespace ActiveStack.Bootstrapper.Host;

public partial class App : System.Windows.Application
{
    public App()
    {
        // Both hosts (Burn BA and --standalone) construct App directly instead of
        // going through the WPF-generated Main, so App.xaml — and with it the
        // PageTemplates.xaml merged dictionary that holds every wizard DataTemplate —
        // never loads unless we do it here. Without this, ContentControl falls back
        // to ToString() and pages render as raw type names.
        InitializeComponent();
    }
}
