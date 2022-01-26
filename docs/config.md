# Configuration

**Config Options**

|Option|Description|
|---|---|
| `group` | The name of the linux user group to be maintained. |
| `keephomedir` | The option to delete or keep a user's home folder when their SSH key or user is no longer detected. |
| `logfile` | The path to the applicaton's output log. |
| `provider` | The provider to configure the application for. |
| `provider-options` | Options related to the provider. ([GSuite](./gsuite.md#adding-configuration-options-for-gsuite), [AWS IAM](./aws.md))|

**Example config.yml**

```yaml
# ========
# GENERAL CONFIG
# ========
# Linux Group name to be maintained
group: "iamusersync"

# If set to false, when a user is removed from the system,
# their home folder will also be deleted
keephomedir: false

# Full path to log file
logfile: "/var/log/iamusersync.log"

# The provider to configure the application for
# and it's properties
provider: "<PROVIDER>"
provider-options:
  . . .
```