using ActiveStack.Bootstrapper.Core;
using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

public sealed class BackupActionResultParserTests
{
    [Fact]
    public void BuildFromJson_SuccessResponse_ParsesFields()
    {
        const string json = """{"success":true,"message":"restored","id":"backup-1"}""";

        var result = BackupActionResultParser.BuildFromJson(json);

        Assert.True(result.Success);
        Assert.Equal("backup-1", result.Id);
        Assert.Equal("restored", result.Message);
    }

    [Fact]
    public void BuildFromJson_FailureResponse_PreservesMessageAndId()
    {
        const string json = """{"success":false,"message":"backup not found","id":"does-not-exist"}""";

        var result = BackupActionResultParser.BuildFromJson(json);

        Assert.False(result.Success);
        Assert.Equal("does-not-exist", result.Id);
        Assert.Equal("backup not found", result.Message);
    }
}
