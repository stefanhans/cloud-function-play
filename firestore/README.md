### Google Cloud Platform's Cloud Functions

Using the Google Cloud Functions Go 1.11 Alpha, I have implemented a simple service with Google Firestore.

Currently, you need to register for [early access](https://docs.google.com/forms/d/e/1FAIpQLSfJ08R2z7FumQyYGGuTyK4x5M-6ch7WmJ_3uWYI5SdZUb5SBw/viewform).

After preparing the environment, we start setting some environment variables:

* GCP_PROJECT=\<project id>
* GOOGLE_APPLICATION_CREDENTIALS=\<json file>
* GCP_REGION=\<region>

We use Go modules for our dependencies, therefore, 
having the source files in place, we execute the following:

```
export GO111MODULE=on
go mod init && go mod vendor
```

Now, we can deploy our functions:

```
gcloud alpha functions deploy subscribe --region $GCP_REGION --entry-point Subscribe --runtime go111 --trigger-http
gcloud alpha functions deploy list --region $GCP_REGION --entry-point List --runtime go111 --trigger-http
gcloud alpha functions deploy unsubscribe --region $GCP_REGION --entry-point Unsubscribe --runtime go111 --trigger-http
gcloud alpha functions deploy reset --region $GCP_REGION --entry-point Reset --runtime go111 --trigger-http


```

Using the URL from the deployment, we test the functions:


```
curl <url>/subscribe <name> <ip> <port>
curl <url>/list
curl <url>/unsubscribe <id>
curl <url>/reset
```
