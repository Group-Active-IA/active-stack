using ActiveStack.Bootstrapper.Host;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

public sealed class InstallProgressSnapshotParserTests
{
    // D2, design.md (Bug B): the engine may interleave a non-JSON diagnostic
    // line on stdout; a single malformed line must not abort the whole
    // stream with an uncaught JsonException. Parse() tolerates it as an
    // ignorable/log-type snapshot carrying the raw line instead.

    [Fact]
    public void Parse_NonJsonLine_DoesNotThrowAndReturnsLogSnapshotCarryingTheRawLine()
    {
        const string line = "warning: something";

        var snapshot = InstallProgressSnapshotParser.Parse(line);

        Assert.Equal("log", snapshot.Type);
        Assert.Equal(line, snapshot.Message);
        Assert.False(snapshot.Success);
    }

    [Fact]
    public void Parse_ABareTokenLine_AlsoDoesNotThrowAndReturnsALogSnapshot()
    {
        // A different non-JSON shape than the first case (a bare token, not
        // a "word: word" sentence) — rules out matching one literal string.
        const string line = "npm-warn-EBADENGINE";

        var snapshot = InstallProgressSnapshotParser.Parse(line);

        Assert.Equal("log", snapshot.Type);
        Assert.Equal(line, snapshot.Message);
    }

    [Fact]
    public void Parse_WhitespaceLine_StillDoesNotThrow()
    {
        // The streaming loop (RunStreamingCommandAsync) already skips
        // blank/whitespace lines before ever calling Parse() — that
        // existing contract is untouched by the parser hardening. This
        // pins Parse() itself as equally tolerant if ever called directly
        // with one, rather than resurrecting the uncaught JsonException.
        var snapshot = InstallProgressSnapshotParser.Parse("   ");

        Assert.Equal("log", snapshot.Type);
    }

    [Fact]
    public void Parse_ValidJsonLine_StillParsesToTheCorrectTypedSnapshot()
    {
        const string line = """{"type":"step_started","phase":"apply","step_id":"openspec","message":"Installing OpenSpec","success":false}""";

        var snapshot = InstallProgressSnapshotParser.Parse(line);

        Assert.Equal("step_started", snapshot.Type);
        Assert.Equal("apply", snapshot.Phase);
        Assert.Equal("openspec", snapshot.StepId);
        Assert.Equal("Installing OpenSpec", snapshot.Message);
    }
}
