using ActiveStack.Bootstrapper.Host;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class InstallProgressFileTailerTests
{
    [Fact]
    public void Parse_MapsSnakeCaseJsonFields()
    {
        var snapshot = InstallProgressSnapshotParser.Parse("""
            {"type":"step_started","phase":"apply","step_id":"openspec","message":"Installing OpenSpec.","success":false}
            """);

        Assert.Equal("step_started", snapshot.Type);
        Assert.Equal("apply", snapshot.Phase);
        Assert.Equal("openspec", snapshot.StepId);
        Assert.Equal("Installing OpenSpec.", snapshot.Message);
        Assert.False(snapshot.Success);
    }

    [Fact]
    public async Task TailAsync_YieldsAppendedJsonLines()
    {
        var path = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.jsonl");
        var tailer = new InstallProgressFileTailer();
        var keepRunning = true;
        var results = new List<InstallProgressSnapshot>();

        try
        {
            var tailTask = Task.Run(async () =>
            {
                await foreach (var snapshot in tailer.TailAsync(path, () => keepRunning))
                {
                    results.Add(snapshot);
                    if (results.Count >= 2)
                    {
                        keepRunning = false;
                    }
                }
            });

            await File.WriteAllTextAsync(path, "{\"type\":\"phase_started\",\"phase\":\"install\",\"message\":\"Starting installation.\",\"success\":false}\n");
            await Task.Delay(150);
            await File.AppendAllTextAsync(path, "{\"type\":\"install_finished\",\"phase\":\"install\",\"message\":\"Installation finished successfully.\",\"success\":true}\n");

            await tailTask;

            Assert.Equal(2, results.Count);
            Assert.Equal("phase_started", results[0].Type);
            Assert.Equal("install_finished", results[1].Type);
            Assert.True(results[1].Success);
        }
        finally
        {
            keepRunning = false;
            if (File.Exists(path))
            {
                File.Delete(path);
            }
        }
    }
}
