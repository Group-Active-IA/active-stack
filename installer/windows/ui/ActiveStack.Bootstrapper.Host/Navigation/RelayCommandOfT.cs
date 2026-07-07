using System.Windows.Input;

namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Minimal parameterized <see cref="ICommand"/> companion to
/// <see cref="RelayCommand"/>, used where the view needs to pass the
/// clicked item (e.g. which <c>BackupEntry</c> a Restore/Delete/Rename
/// button targets) as the command parameter.
/// </summary>
public sealed class RelayCommand<T> : ICommand
    where T : class
{
    private readonly Action<T?> _execute;

    public RelayCommand(Action<T?> execute)
    {
        _execute = execute;
    }

    public event EventHandler? CanExecuteChanged
    {
        add { }
        remove { }
    }

    public bool CanExecute(object? parameter) => true;

    public void Execute(object? parameter) => _execute(parameter as T);
}
