using ActiveStack.Bootstrapper.Core.Localization;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class LanguagePreferenceTests
{
    [Fact]
    public void SaveThenLoad_RoundTripsTheLanguage()
    {
        var homeDir = CreateTempHomeDir();
        try
        {
            LanguagePreference.Save(homeDir, "es");

            Assert.Equal("es", LanguagePreference.Load(homeDir));
        }
        finally
        {
            Directory.Delete(homeDir, recursive: true);
        }
    }

    [Fact]
    public void Save_PreservesAPreExistingUnrelatedKey()
    {
        var homeDir = CreateTempHomeDir();
        try
        {
            var configPath = ConfigPath(homeDir);
            Directory.CreateDirectory(Path.GetDirectoryName(configPath)!);
            File.WriteAllText(configPath, """{"other_key":"keep-me"}""");

            LanguagePreference.Save(homeDir, "es");

            var json = File.ReadAllText(configPath);
            Assert.Contains("\"language\":\"es\"", json);
            Assert.Contains("\"other_key\":\"keep-me\"", json);
            Assert.Equal("es", LanguagePreference.Load(homeDir));
        }
        finally
        {
            Directory.Delete(homeDir, recursive: true);
        }
    }

    [Fact]
    public void Load_NoFile_ReturnsNullWithoutThrowing()
    {
        var homeDir = CreateTempHomeDir();
        try
        {
            Assert.Null(LanguagePreference.Load(homeDir));
        }
        finally
        {
            Directory.Delete(homeDir, recursive: true);
        }
    }

    [Fact]
    public void Load_FileWithoutLanguageKey_ReturnsNullWithoutThrowing()
    {
        var homeDir = CreateTempHomeDir();
        try
        {
            var configPath = ConfigPath(homeDir);
            Directory.CreateDirectory(Path.GetDirectoryName(configPath)!);
            File.WriteAllText(configPath, """{"other_key":"value"}""");

            Assert.Null(LanguagePreference.Load(homeDir));
        }
        finally
        {
            Directory.Delete(homeDir, recursive: true);
        }
    }

    [Fact]
    public void Load_MalformedOrEmptyJson_ReturnsNullWithoutThrowing()
    {
        var homeDir = CreateTempHomeDir();
        try
        {
            var configPath = ConfigPath(homeDir);
            Directory.CreateDirectory(Path.GetDirectoryName(configPath)!);

            File.WriteAllText(configPath, "");
            Assert.Null(LanguagePreference.Load(homeDir));

            File.WriteAllText(configPath, "{ not valid json ");
            Assert.Null(LanguagePreference.Load(homeDir));
        }
        finally
        {
            Directory.Delete(homeDir, recursive: true);
        }
    }

    [Fact]
    public void Save_MalformedExistingFile_OverwritesWithAFreshValidDocument()
    {
        var homeDir = CreateTempHomeDir();
        try
        {
            var configPath = ConfigPath(homeDir);
            Directory.CreateDirectory(Path.GetDirectoryName(configPath)!);
            File.WriteAllText(configPath, "{ not valid json ");

            LanguagePreference.Save(homeDir, "en");

            Assert.Equal("en", LanguagePreference.Load(homeDir));
        }
        finally
        {
            Directory.Delete(homeDir, recursive: true);
        }
    }

    private static string ConfigPath(string homeDir) => Path.Combine(homeDir, ".active-stack", "config.json");

    private static string CreateTempHomeDir()
    {
        var dir = Path.Combine(Path.GetTempPath(), "active-stack-tests-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(dir);
        return dir;
    }
}
