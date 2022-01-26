# GSuite Provider Setup

## Google Workspaces Setup
1. Log into your Google Workspaces admin account and navigate to your users: https://admin.google.com/ac/users
2. Click the `More Options` Dropdown and select `Manage custom attributes`
   ![Alt Text](https://github.com/danetuso/iam-user-sync/blob/main/manage_custom_attributes.png)
3. At the top right, click `Add Custom Attribute`
   ![Alt Text](https://github.com/danetuso/iam-user-sync/tree/master/docs/resources/custom_field.png)
   - Take note of what you choose for the `Category` field as we will use it later. 
   - I use `SSHKEY` in this example. 
   - The `type` is set to `Text`.
   - The `visibility` is up to your use case.
   - The number of values is set to `Single Value`.
4. For each user, navigate to their profile and click the `User information / User Details` dropdown.
5. Under the default Employee information section, there will now be a `SSHKEY` section. Paste the user's public ssh key in this field and save.

## Google Cloud Platform Setup

In order to programmatically access your domain users, you must set up a Google Cloud Platform project and then assign appropriate permissions.

1. Navigate to: https://cloud.google.com and sign in.
2. Click Console at the top right or go straight to: https://console.cloud.google.com
3. Follow the prompts to create your first project.
4. Name it appropriately or use the default values.
5. On the left hand side, drop down the list of services and select IAM & Admin.
6. Select `Service Accounts` on the left, then `Create Service Account` at the top.
7. Name it how you like, *(you can skip the following 2 fields for granting access)*.
8. Click `Create and Continue` then click `Done`.
   ![Alt Text](https://github.com/danetuso/iam-user-sync/tree/master/docs/resources/create_service_account.png)
9. Select your new service account from the list and select the `Keys` tab at the top.
10. Click `Add Key` with the key type set to JSON.
11. Save the downloaded file as you'll need to upload this to each server along with the application.
12. Go back to the `Details` tab and select `Show Advanced Settings`.
13. Note the `Client-ID` under the Domain-wide delegation section, as you'll use this in the next step.
14. Navigate back to https://admin.google.com and go to `Security`.
15. Dropdown `Access and Data Control` and select `API Controls`. Or navigate directly to https://admin.google.com/ac/owl
16. At the bottom, select `Manage Domain Wide Delegation`, then `Add new` at the top.
17. From the Google Cloud Platform service account, copy the `Client-ID` and paste it here.
18. For the OAUTH scope, use `https://www.googleapis.com/auth/admin.directory.user` then click Authorize.

**IMPORTANT!**: You must note the Google Workspace account you are currently signed into when Authorizing the Domain Wide Delegation as this user's email must be used in the config.

## Adding configuration options for GSuite

See [Configuration](./config.md) for more information about config files

**Note:** From the previous steps, you'll need the custom attribute category you assigned earlier. In our example it was `SSHKEY`. You'll also need the credentials json file that you generated as well as the administrator's email address that was used to enable the Domain-wide delegation for OAuth scopes of access.

### GSuite Specific Provider Options

|Option|Description|
|---|---|
| `credentials` | The path to the credentials json file. |
| `customattributekey` | The custom attribute category name. |
| `gsuiteadmin` | The email address of the admin that enabled domain-wide delegation for OAuth. |
| `oauthdomain` | The Google Workspace domain to check for users. Can be commented out if the domain is the same as the gsuiteadmin. |

```yaml
provider: "GSUITE"
provider-options:
  # Full path to credentials file
  credentials: "./credentials.json"

  # Custom attribute name to query per user
  customattributekey: "SSHKEY"

  # Admin email used to delegate domain wide OAuth
  # scope authority
  gsuiteadmin: "administrator@tuso.tech"

  # If the domain to query differs from gsuite admin's domain
  #oauthdomain: "tuso.tech"
```