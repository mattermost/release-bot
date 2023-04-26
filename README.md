# release-bot

Application which subscribes to public repository webhooks and will trigger private pipelines depending on the needs of public repositories. 

## Github App Configuration 

It is **mandatory** to create a Github application Organization wide and control the access through thh Github App settings of the [Organization](https://docs.github.com/en/apps/creating-github-apps/creating-github-apps/creating-a-github-app). This is due to the fact that this application is security oriented so and enables triggering pipelines on private repos from public webhooks. 

> The application is using [this](https://github.com/bradleyfalzon/ghinstallation) library to authenticate over Github.

Once you create the Github App you will need 3 things to configure the application. [Sample Config](https://github.com/mattermost/release-bot/blob/main/config/testdata/config_sample.yaml#L10)

- **Integration ID**: *You can find it on the Advanced page of the Github App by checking the header of the webhook with the name **X-GitHub-Hook-Installation-Target-ID***
- **Private Key:** *This is unique per Github Application and it is generated upon creation*
- **Webhook Secret:** *This is mandatory and needed to secure the webhooks and ensure they are coming from github. More info [here](https://docs.github.com/en/webhooks-and-events/webhooks/securing-your-webhooks)*

After that you can subscribe to webhooks referring to the repos your application has access to.

- Repository access is set under the Github Apps page on the Organization settings page
- Webhooks are configured on the Github App page from the Developer settings

> For every repo you allow access webhook events will be delivered. For every change in the webhook permissions you need to accept them from the Github App page on the Organization settings page.

## Pipelines Configuration

In order for a pipeline of a private repo to be triggered **it needs to be declared on the configuration file under the [pipelines spec](https://github.com/mattermost/release-bot/blob/main/config/config.go#L32)**. This is by design for security reasons.
 
> **All configuration properties are required**

The workflow file to be triggered **must have at least the [workflow dispatch event configured](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch)** and an input called `payload`. You can then use our custom action to export some triggere metadata if needed.  
Simplest example below:

```yaml
name: test

on:
  workflow_dispatch:
    inputs:
      payload:
        description: "The triggerer payload"
        required: false
        type: string

jobs:
  export-the-variables:
    runs-on: ubuntu-22.04
    steps:
      - uses: mattermost/actions/release/expose-webhook-vars@b34961014df898b876673ca578c06ad6efc1577b
        with:
          payload: ${{ inputs.payload }}

      - run: env | grep TRIGGERER
```

| Configuration Property | Type          | Available Values                                               | Description                                                                 |
| ---------------------- | ------------- | -------------------------------------------------------------- | --------------------------------------------------------------------------- |
| organisation           | `string`      | valid Github organization name                                 | The Github organization of the repo that has the pipeline to be triggered   |
| repository             | `string`      | valid Github repository name                                   | The Github repository of the repo that has the pipeline to be triggered     |
| workflow               | `string`      | valid workflow file name                                       | The Github workflow file name to trigger                                    |
| targetBranch           | `string`      | the target branch of the worklflow                             | The Github repository branch of the worklfow to trigger                     |
| context                | `string`      | valid status check name                                        | The Github status check name to apply to the originating commit             |
| timeout                | `duration`    | valid [golang duration](https://pkg.go.dev/time#ParseDuration) | The duration to wait for the workflow to finish                             |
| sleepSeconds           | `int64`       | valid seconds integer                                          | The duration to sleep after triggering the workflow finish                  |
| conditions             | `[]condition` | valid list of conditions                                       | The conditions that need to apply in order for the workflow to be triggered |

Below a detailed condition table

| Configuration Property | Type       | Available Values                                                                                                                  | Description                                                 |
| ---------------------- | ---------- | --------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| repository             | `regex`    | should match the ***\<org\>/\<repo\>*** format                                                                                    | The Github repo where the event comes from                  |
| webhook                | `[]string` | can include `pull_request`,`push`,`workflow_run`                                                                                  | The actual webhook event that will trigger the workflow     |
| workflow               | `string`   | the workflow name if `workflow_run` in webhook list                                                                               | The workflow name of the repo to check for condition        |
| type                   | `string`   | can be one of `pull_request`,`push`,`tag`                                                                                         | The type of event that triggered the webhook                |
| status                 | `string`   | can be one of `requested`,`in_progress`,`completed`,`queued`,`pending`,`waiting`                                                  | The status of the workflow if `workflow_run` in webhook     |
| conclusion             | `string`   | can be one of `success`,`failure`,`neutral`,`cancelled`,`timed_out`,`action_required`,`stale`,`null`,`skipped`,` startup_failure` | The conslusion of the workflow if `workflow_run` in webhook |

> All conditions **must** match in order for the workflow to be triggered