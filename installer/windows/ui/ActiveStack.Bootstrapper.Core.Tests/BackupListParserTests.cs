using ActiveStack.Bootstrapper.Core;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class BackupListParserTests
{
    [Fact]
    public void BuildFromJson_ParsesAllFields()
    {
        const string json = """
        {
          "backups": [
            {
              "id": "backup-1",
              "created_at": "2026-03-15T10:30:00Z",
              "source": "install",
              "description": "before upgrade",
              "file_count": 3,
              "pinned": true,
              "compressed": false,
              "display_label": "backup-1 (install)",
              "manifest_path": "C:\\backups\\backup-1\\manifest.json"
            }
          ]
        }
        """;

        var backups = BackupListParser.BuildFromJson(json);

        Assert.Single(backups);
        var b = backups[0];
        Assert.Equal("backup-1", b.Id);
        Assert.Equal("2026-03-15T10:30:00Z", b.CreatedAt);
        Assert.Equal("install", b.Source);
        Assert.Equal("before upgrade", b.Description);
        Assert.Equal(3, b.FileCount);
        Assert.True(b.Pinned);
        Assert.False(b.Compressed);
        Assert.Equal("backup-1 (install)", b.DisplayLabel);
        Assert.Equal("C:\\backups\\backup-1\\manifest.json", b.ManifestPath);
    }

    [Fact]
    public void BuildFromJson_EmptyStore_ReturnsEmptyListNotNull()
    {
        const string json = """{"backups":[]}""";

        var backups = BackupListParser.BuildFromJson(json);

        Assert.NotNull(backups);
        Assert.Empty(backups);
    }
}
