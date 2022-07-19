# aws-billing-store

This project deploys an automated CUR ingestion which updates an Athena table setup to enable querying of the latest snapshots for each month. This is an alternative to the pre configured athena setup provided by the Billing Team and is more suited to customers with 10s or 100s of CUR files provided in each report.

# Overview

The goal of this project is to provide a consistent view of cost and usage reports (CUR) at all times in a Athena table. To do this we use hive symlinks, which are updated each time a new snapshot arrives providing an atomic single file update for new CURs. This is different to the status table provided by AWS Billing team in the pre configured Athena setup.

# How It Works

When you enable the option to keep all versions of the CUR, AWS will upload a new snapshot then once complete update the manifest containing a list of file paths for that billing period. We use an s3 file create event to trigger reading of that manifest and creation of a symlink in the hive directory we maintain in the same bucket. This provides Athena with a partitioned structure to query without worrying about CUR files being updated while it is reading them.

# Prerequisites

1. An AWS account.
2. [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html).
3. Exported environment variables for `AWS_DEFAULT_REGION`, `AWS_REGION` and `AWS_PROFILE`. 

# Deployment

First you will need to deploy the bucket we use to store lambda and CFN artifacts.

```
make deploy-bucket
```

Deploy the solution.

```
make deploy
```

This will deploy the following components:

1. Setup a bucket for the CUR in the region you configured via `AWS_DEFAULT_REGION`.
2. Create a CUR in the billing service.
3. Deploy the lambda which creates the hive directory containing symlinks to the latest CUR.
4. Deploy the Glue database and table used by Athena.
5. Deploy the Athena workspace with an encrypted secure S3 bucket for artifacts.
6. Deploy the template which creates partitions in Athena based on files created in the hive directory.

# What Next? 

There are some great resources with queries which provide insights from your CUR data, one of the best is [Level 300: AWS CUR Query Library](https://wellarchitectedlabs.com/cost/300_labs/300_cur_queries/) from the [The Well-Architected Labs website](https://wellarchitectedlabs.com/).

# License

This project is released under Apache 2.0 license and is copyright [Mark Wolfe](https://www.wolfe.id.au).