using System.Runtime.CompilerServices;
using System.IO;

namespace ActiveStack.Bootstrapper.Host;

internal sealed class InstallProgressFileTailer
{
    public async IAsyncEnumerable<InstallProgressSnapshot> TailAsync(
        string path,
        Func<bool> shouldContinue,
        [EnumeratorCancellation] CancellationToken cancellationToken = default)
    {
        var seenLines = 0;

        while (shouldContinue() && !cancellationToken.IsCancellationRequested)
        {
            cancellationToken.ThrowIfCancellationRequested();

            if (!File.Exists(path))
            {
                if (!shouldContinue())
                {
                    yield break;
                }

                await Task.Delay(100, cancellationToken);
                continue;
            }

            string[] lines;
            try
            {
                lines = await File.ReadAllLinesAsync(path, cancellationToken);
            }
            catch (IOException)
            {
                await Task.Delay(50, cancellationToken);
                continue;
            }

            while (seenLines < lines.Length)
            {
                var line = lines[seenLines].Trim();
                seenLines++;

                if (line.Length == 0)
                {
                    continue;
                }

                yield return InstallProgressSnapshotParser.Parse(line);
            }

            if (!shouldContinue())
            {
                yield break;
            }

            await Task.Delay(100, cancellationToken);
        }
    }
}
