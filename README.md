# IAM User Sync
Creates a linux group of users synced to your Google Workspace users and automatically imports their public SSH keys.

GSuite / Google Workspaces is the only provider currently available, but will soon work with AWS IAM as well.

---

## Google Workspaces Setup
Log into your Google Workspaces admin account and navigate to your users: https://admin.google.com/ac/users

Click the `More Options` Dropdown and select `Manage custom attributes`

![Alt Text](https://github.com/danetuso/iam-user-sync/blob/main/manage_custom_attributes.png)

At the top right, click `Add Custom Attribute`

Take note of what you choose for the `Category` field as we will use it later. I use `SSHKEY` in this example. The type is set to `Text` and the visibility is up to your use case. # of values is set to `Single Value`.

![Alt Text](https://github.com/danetuso/iam-user-sync/blob/main/custom_field.png)


For each user, navigate to their profile and click the User information / User Details dropdown.

Under the default Employee information section, there will now be a `SSHKEY` section. Paste the user's public ssh key in this field and save.

---

## Google Cloud Platform Setup

In order to programatically access your domain users, you must set up a Google Cloud Platform project and then assign appropriate permissions.

To do this, navigate to: https://cloud.google.com and sign in. Click Console at the top right or go straight to: https://console.cloud.google.com where it will prompt you to create your first project. Name it appropriately or use the default values.

On the left hand side, drop down the list of services and select IAM & Admin. Then select `Service Accounts` on the left, then `Create Service Account` at the top.

Name it how you like, and the following 2 fields for granting access can be skipped. Click `Create and Continue` then click `Done`.

![Alt Text](https://github.com/danetuso/iam-user-sync/blob/main/create_service_account.png)

Select your new service account from the list and select the `Keys` tab at the top. Click `Add Key` with the  key type set to JSON. Save the downloaded file as you'll need to upload this to each server along with the application.

Go back to the `Details` tab and select `Show Advanced Settings`. Note the `Client-ID` under the Domain-wide delegation section, as you'll use this in the next step.

Navigate back to https://admin.google.com and go to `Security` then dropdown `Access and Data Control` and select `API Controls`. Or navigate directly to https://admin.google.com/ac/owl

At the bottom, select `Manage Domain Wide Delegation`, then `Add new` at the top.

From the Google Cloud Platform service account, copy the `Client-ID` and paste it here. For the OAUTH scope, use `https://www.googleapis.com/auth/admin.directory.user` then click Authorize.

IMPORTANT! You must note the Google Workspace account you are currently signed into when Authorizing the Domain Wide Delegation as this user's email must be used in the config.

---

## Ubuntu Setup (Config & Credentials)

From the previous steps, you'll need the custom attribute category you assigned earlier. In our example it was `SSHKEY`. You'll also need the credentials json file that you generated as well as the administrator's email address that was used to enable the Domain-wide delegation for OAuth scopes of access.

Configuration options include: 
|  |  |
|---|---|
| `group` | The name of the linux user group to be maintained. |
| `keephomedir` |  The option to delete or keep a user's home folder when their SSH key or user is no longer detected. |
| `logfile` | The path to the applicaton's output log. |
| `credentials` | The path to the credentials json file. |
| `customattributekey` | The custom attribute category name. |
| `gsuiteadmin` | The email address of the admin that enabled domain-wide delegation for OAuth. |
| `oauthdomain` | The Google Workspace domain to check for users. Can be commented out if the domain is the same as the gsuiteadmin. |

### Example config.yml:

```
# ========
# GENERAL CONFIG
# ========
# Linux Group name to be maintained
group: "iamusersync"

# If set to false, when a user is removed from the system, their home folder will also be deleted
keephomedir: false

# Full path to log file
logfile: "/var/log/iamusersync.log"


# ========
# GSUITE
# ========
provider: "GSUITE"

# Full path to credentials file
credentials: "./credentials.json"

# Custom attribute name to query per user
customattributekey: "SSHKEY"

# Admin email used to delegate domain wide OAuth scope auhtority
gsuiteadmin: "administrator@tuso.tech"

# if domain to query differs from gsuite admin's domain
#oauthdomain: "tuso.tech"
```

---

## Ubuntu Setup (Cron) & Example Usage

When running the application, you must use the `--config` argument to define the path to the config file, otherwise you must supply each of the config variables on the command line. If you define a config file and additional arguments, those supplied on the command line will overwrite what is set in the config.

&nbsp;

I recommend using a cronjob to run the application at an interval appropriate to your needs.

```
*/15 * * * * /usr/local/bin/iamusersync --config /usr/local/etc/iamusersync/config.yml
```
Note: You can put the application, config, and log anywhere you like.

The default log file path is set to `/var/log/iamusersync.log`

&nbsp;

[Discord Screenshot Resizer](https://github.com/danetuso/discord-screenshot-resizer).
