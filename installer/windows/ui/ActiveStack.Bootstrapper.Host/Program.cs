using System.Windows;
using WixToolset.BootstrapperApplicationApi;

namespace ActiveStack.Bootstrapper.Host;

internal static class Program
{
    private static int Main(string[] args)
    {
        if (args.Any(arg => string.Equals(arg, "--standalone", StringComparison.OrdinalIgnoreCase)))
        {
            return RunStandalone();
        }

        var application = new ActiveStackBootstrapperApplication();
        ManagedBootstrapperApplication.Run(application);
        return 0;
    }

    private static int RunStandalone()
    {
        var thread = new Thread(() =>
        {
            var app = new App
            {
                ShutdownMode = ShutdownMode.OnMainWindowClose
            };

            var window = new MainWindow();
            app.Run(window);
        });

        thread.SetApartmentState(ApartmentState.STA);
        thread.Start();
        thread.Join();
        return 0;
    }
}
