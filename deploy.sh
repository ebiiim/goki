#!/bin/sh

set -x

# load .env
set -o allexport; source .env; set +o allexport

# set GCP project
gcloud config set project $GCP_ID

# build and push container image
docker build -t gcr.io/$GCP_ID/$APP_NAME:$VERSION .
docker push gcr.io/$GCP_ID/$APP_NAME:$VERSION

# deploy
# --max-instances=1 avoids potential GCS read/write race condition
gcloud run deploy $APP_NAME \
  --image gcr.io/$GCP_ID/$APP_NAME:$VERSION \
  --platform managed \
  --memory=128Mi --cpu=1000m \
  --max-instances=1 \
  --set-env-vars=GCP_ID=$GCP_ID,GCS_DB_BUCKET=$GCS_DB_BUCKET,TWITTER_CONSUMER_KEY=$TWITTER_CONSUMER_KEY,TWITTER_CONSUMER_SECRET=$TWITTER_CONSUMER_SECRET,TWITTER_CALLBACK_SERVER_NAME=$TWITTER_CALLBACK_SERVER_NAME \
  --region=asia-northeast1 \
  --service-account=$GCP_RUN_SERVICE_ACCOUNT \
  --allow-unauthenticated
