using System.Collections.ObjectModel;
using System.Windows.Input;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Backups;

/// <summary>
/// Standalone page (not a wizard step, D7 design.md): lists backups from
/// <see cref="IInstallerEngineClient.ListBackupsAsync"/> by
/// <c>DisplayLabel</c> and offers per-item Restore/Delete/Rename actions
/// with inline confirmations matching the TUI text. <see cref="CanAdvance"/>
/// is always false — the shell suppresses its primary "Next" here; Back
/// returns to the Hub. Re-lists from the store after every action so the
/// view can never drift from it.
/// </summary>
public sealed class BackupsPageViewModel : WizardPageViewModelBase
{
    private const string RestoreWarning = "This will OVERWRITE your current configuration.";
    private const string DeleteWarning = "This will PERMANENTLY DELETE this backup.";

    private readonly IInstallerEngineClient _engineClient;
    private BackupPageAction _action = BackupPageAction.None;
    private BackupEntry? _targetBackup;
    private string _renameInput = string.Empty;
    private string? _lastMessage;

    public BackupsPageViewModel(IInstallerEngineClient engineClient)
        : base("Manage backups", "Restore, rename, or delete a previous Active Stack backup.")
    {
        _engineClient = engineClient;

        BeginRestoreCommand = new RelayCommand<BackupEntry>(backup => { if (backup is not null) BeginRestore(backup); });
        BeginDeleteCommand = new RelayCommand<BackupEntry>(backup => { if (backup is not null) BeginDelete(backup); });
        BeginRenameCommand = new RelayCommand<BackupEntry>(backup => { if (backup is not null) BeginRename(backup); });
        ConfirmActionCommand = new RelayCommand(() => _ = ConfirmActionAsync());
        CancelActionCommand = new RelayCommand(CancelAction);
    }

    public ObservableCollection<BackupEntry> Items { get; } = [];

    /// <summary>Bound by each item's Restore button — no code-behind click handler needed.</summary>
    public ICommand BeginRestoreCommand { get; }

    /// <summary>Bound by each item's Delete button.</summary>
    public ICommand BeginDeleteCommand { get; }

    /// <summary>Bound by each item's Rename button.</summary>
    public ICommand BeginRenameCommand { get; }

    /// <summary>Bound by the confirmation overlay's Confirm button.</summary>
    public ICommand ConfirmActionCommand { get; }

    /// <summary>Bound by the confirmation overlay's Cancel button.</summary>
    public ICommand CancelActionCommand { get; }

    public bool IsEmpty => Items.Count == 0;

    public BackupPageAction Action
    {
        get => _action;
        private set
        {
            if (_action == value)
            {
                return;
            }

            _action = value;
            RaisePropertyChanged(nameof(Action));
            RaisePropertyChanged(nameof(IsConfirming));
            RaisePropertyChanged(nameof(ConfirmationText));
        }
    }

    /// <summary>True while a Restore/Delete/Rename confirmation is active, for the view's confirmation overlay.</summary>
    public bool IsConfirming => Action != BackupPageAction.None;

    public BackupEntry? TargetBackup
    {
        get => _targetBackup;
        private set => SetField(ref _targetBackup, value);
    }

    public string RenameInput
    {
        get => _renameInput;
        set => SetField(ref _renameInput, value);
    }

    public string? LastMessage
    {
        get => _lastMessage;
        private set => SetField(ref _lastMessage, value);
    }

    public string ConfirmationText => Action switch
    {
        BackupPageAction.Restore => RestoreWarning,
        BackupPageAction.Delete => DeleteWarning,
        _ => string.Empty
    };

    /// <summary>Never advanceable — Backups is a standalone page, not a wizard step (D7).</summary>
    public override bool CanAdvance => false;

    public async Task RefreshAsync(CancellationToken cancellationToken = default)
    {
        var backups = await _engineClient.ListBackupsAsync(cancellationToken);

        Items.Clear();
        foreach (var backup in backups)
        {
            Items.Add(backup);
        }

        RaisePropertyChanged(nameof(IsEmpty));
    }

    public void BeginRestore(BackupEntry backup) => BeginAction(BackupPageAction.Restore, backup);

    public void BeginDelete(BackupEntry backup) => BeginAction(BackupPageAction.Delete, backup);

    public void BeginRename(BackupEntry backup) => BeginAction(BackupPageAction.Rename, backup);

    public void CancelAction()
    {
        Action = BackupPageAction.None;
        TargetBackup = null;
        RenameInput = string.Empty;
    }

    public async Task ConfirmActionAsync(CancellationToken cancellationToken = default)
    {
        if (Action == BackupPageAction.None || TargetBackup is null)
        {
            return;
        }

        var actionId = Action switch
        {
            BackupPageAction.Restore => "restore",
            BackupPageAction.Delete => "delete",
            BackupPageAction.Rename => "rename",
            _ => throw new InvalidOperationException("No action is active.")
        };

        var description = Action == BackupPageAction.Rename ? RenameInput : null;
        var backupId = TargetBackup.Id;

        var result = await _engineClient.RunBackupActionAsync(actionId, backupId, description, cancellationToken);

        LastMessage = result.Message;
        Action = BackupPageAction.None;
        TargetBackup = null;
        RenameInput = string.Empty;

        await RefreshAsync(cancellationToken);
    }

    private void BeginAction(BackupPageAction action, BackupEntry backup)
    {
        TargetBackup = backup;
        RenameInput = string.Empty;
        Action = action;
    }
}
