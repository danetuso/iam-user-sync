# Getting Started

## Configuring for your IAM provider

- [Configure for GSuite](./gsuite.md)
- [Configure for AWS IAM (Coming Soon)](./aws.md)
- [Configuration Documentation](./config.md)

### Example Usage

When running the application, you must use the `--config` argument to define the path to the config file, otherwise you must supply each of the config variables on the command line. If you define a config file and additional arguments, those supplied on the command line will overwrite what is set in the config.

```shell
/usr/local/bin/iamusersync --config /usr/local/etc/iamusersync/config.yml
```

### Using a Cron Job

I recommend using a cronjob to run the application at an interval appropriate to your needs.

```
*/15 * * * * /usr/local/bin/iamusersync --config /usr/local/etc/iamusersync/config.yml
```

**Note:** You can put the application, config, and log anywhere you like. The default log file path is set to `/var/log/iamusersync.log`

---

## Build the application

```
git clone https://github.com/danetuso/iam-user-sync.git
cd iam-user-sync
go mod init iam-user-sync
go mod tidy
go build
```

Additonal optional commands to move the binary and config:
```
mv ./iam-user-sync /usr/local/bin/iamusersync
mkdir /usr/local/etc/iamusersync
cp ./config.yml /usr/local/etc/iamusersync/config.yml
```