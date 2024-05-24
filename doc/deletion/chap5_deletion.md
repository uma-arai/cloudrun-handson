# **The Cloud Run ハンズオン <br /> 5章削除手順**

この章では、5章で作成したリソースを削除する手順を説明します。
なお、削除の方法としては次の2つがあります。

- プロジェクトごと削除
- 作成したリソースに絞って削除

もっともシンプルな方法は前者の「プロジェクトごと削除」です。
ただし、プロジェクトごと削除すると、そのプロジェクト内のすべてのリソースが削除されるため、他で作成したリソースも削除されてしまいます。
**今回のハンズオン専用でプロジェクトを作った場合のみ**、「プロジェクトごと削除」の方法を選択してください。

以降は、**作成したリソースに絞って削除する方法**を説明します。

## **リージョン変数の設定**

```bash
export REGION=asia-northeast1
```

## **Schedulerの削除**

```bash
gcloud scheduler jobs delete cnsrun-batch-job-scheduler --location="$REGION" --quiet
```

## **Cloud Runジョブの削除**

```bash
gcloud run jobs delete cnsrun-batch --region="$REGION" --quiet
```

## **SQLの削除**

削除保護があるため、削除保護を外して削除をします。
Cloud SQLの削除は時間がかかるため非同期オプション(`--async`)を付けて削除します。
1時間ほど経過後、削除されていることを確認するのがお勧めです。

```bash
gcloud sql instances patch cnsrun-app-instance --no-deletion-protection
gcloud sql instances delete cnsrun-app-instance --async --quiet
```

## **Cloud Buildの削除**

```bash
gcloud beta builds triggers list --region=$REGION --format=json | jq -r .[].name | grep cnsrun | xargs -I @ gcloud beta builds triggers delete --region=$REGION @
```

```bash
gcloud beta builds repositories list --region=$REGION --connection=cnsrun-app-handson --format=json | jq -r .[].name | xargs -I @ gcloud builds repositories delete --region=$REGION --connection=cnsrun-handson @ --quiet
```

```bash
gcloud beta builds connections delete cnsrun-app-handson --region=$REGION --quiet
```

## **Cloud Deployの削除**

```bash
gcloud beta deploy delivery-pipelines list --region=$REGION --format=json | jq -r .[].name | grep cnsrun | xargs -I @ gcloud beta deploy delivery-pipelines delete --region=$REGION @ --quiet --force
```

## **Cloud Runサービスの削除**

```bash
gcloud config set run/region asia-northeast1
gcloud run services list --format=json | jq -r .[].metadata.name | grep cnsrun | xargs -I @ gcloud run services delete @ --quiet
```

## **Artifact Registryの削除**

```bash
gcloud artifacts repositories delete cnsrun-app --location="$REGION" --quiet
```

## **Secret Managerの削除**

```bash
gcloud secrets delete cnsrun-app-db-password --quiet
```

## **ロードバランサの削除**

```bash
gcloud compute forwarding-rules delete cnsrun-lb --quiet --global
gcloud compute target-https-proxies delete cnsrun-https-proxies --global --quiet
gcloud compute url-maps  delete cnsrun-urlmaps --global --quiet
gcloud compute backend-services delete cnsrun-backend-services --global --quiet
gcloud compute ssl-certificates delete cnsrun-certificate --quiet
gcloud compute addresses delete cnsrun-ip --global --quiet
```

NEGも合わせて削除します。

```bash
gcloud beta compute network-endpoint-groups delete cnsrun-app-neg-"$REGION" --region="$REGION" --quiet
```

## **VPCネットワークの削除**

```bash
gcloud compute networks peerings delete servicenetworking-googleapis-com --network=cnsrun-app

PSC_ADDRESSES=$(gcloud compute addresses list --global --filter="purpose=VPC_PEERING" --format=json | jq -r .[].name | grep cnsrun)
gcloud compute addresses delete "$PSC_ADDRESSES" --global --quiet
```

VPCサービスで作成したリソースは他にも、静的IPアドレスやサブネット、VPC本体があります。
しかし、プライベートサービスアクセスを削除後、 静的IPアドレスが削除されるには時間がかかります。

そのため、これらのリソースは数日経過後に削除をお願いします。

## **サービスアカウントの削除**

最後はサービスアカウントです。

```bash
gcloud iam service-accounts delete cnsrun-app-frontend@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --quiet
gcloud iam service-accounts delete cnsrun-app-backend@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --quiet
gcloud iam service-accounts delete cnsrun-app-batch@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --quiet
gcloud iam service-accounts delete cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --quiet
gcloud iam service-accounts delete cnsrun-clouddeploy@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --quiet
```

以上で、5章で作成したリソースの削除が完了しました。
お疲れ様でした。
