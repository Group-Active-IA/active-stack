using System.Windows.Input;

namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Minimal parameterless <see cref="ICommand"/> for page-VM actions that
/// have no natural TwoWay-bindable property (e.g. invoking the folder
/// picker). No page VM needs <see cref="CanExecuteChanged"/> to vary — every
/// use so far is always executable — so this stays deliberately small
/// instead of pulling in a full command-manager pattern.
/// </summary>
public sealed class RelayCommand : ICommand
{
    private readonly Action _execute;

    public RelayCommand(Action execute)
    {
        _execute = execute;
    }

    public event EventHandler? CanExecuteChanged
    {
        add { }
        remove { }
    }

    public bool CanExecute(object? parameter) => true;

    public void Execute(object? parameter) => _execute();
}
