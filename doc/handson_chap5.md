# **The Cloud Run ハンズオン <br />（基本編）**

## 概要

ハンズオン（基本編）で構築する全体概要図は[こちらのリンク](https://github.com/uma-arai/cloudrun-handson/blob/main/images/05-handson-architecture-overview.png?raw=true)となります。

今回のハンズオンではリージョンは可能な限り`asia-northeast1`を利用します。

また、ハンズオンを実行するユーザは、基本ロールである「`Owner`」権限を持つプロジェクトを利用してください。
「`Editor`」 権限の場合、都度権限が足りないケースが発生します。
たとえば、KMSの操作時にCloud KMS暗号オペレータロールが持つ権限を不足します。
ハンズオンを円滑にすすめるためにできるだけ、**Owner権限で操作**を行ってください。

<walkthrough-footnote>本ハンズオンの事前準備は、Apache-2.0で配布されている [はじめてみよう Cloud Run ハンズオン](https://github.com/google-cloud-japan/gcp-getting-started-cloudrun/blob/main/tutorial.md)の内容を利用しています。</walkthrough-footnote>

## Google Cloud プロジェクトの設定、確認

### **Cloud Shell の起動**

Google Cloudのプロジェクトにアクセスし、画面上部から<walkthrough-spotlight-pointer spotlightId="devshell-activate-button">Cloud Shell</walkthrough-spotlight-pointer>を起動しましょう。

<walkthrough-open-cloud-shell-button></walkthrough-open-cloud-shell-button>

### **プロジェクトの課金が有効化されていることを確認する**

```bash
gcloud beta billing projects describe ${GOOGLE_CLOUD_PROJECT} | grep billingEnabled
```

**Cloud Shell の承認** という確認メッセージが出た場合は **`[承認]`** をクリックします。

出力結果の `billingEnabled` が **`true`** になっていることを確認してください。**`false`** の場合は、こちらのプロジェクトではハンズオンが進められません。別途、課金を有効化したプロジェクトを用意し、本ページの #1 の手順からやり直してください。

## **環境準備**

<walkthrough-tutorial-duration duration=10></walkthrough-tutorial-duration>

最初に、ハンズオンを進めるための環境準備をします。
次の設定を進めていきます。

- gcloud コマンドラインツール設定
- Google Cloud 機能（API）有効化設定

## **gcloud コマンドラインツール**

Google Cloud は、コマンドライン（CLI）、GUI から操作が可能です。ハンズオンでは主に CLI を使い作業をしますが、一部 GUI での操作もあります。

### **1. gcloud コマンドラインツールとは**

gcloud コマンドラインインタフェースは、Google Cloud でメインとなる CLI ツールです。このツールを使用すると、コマンドラインから、またはスクリプトやほかの自動化ツールにより、多くの一般的なタスクを実行できます。

たとえば、gcloud CLI を使用して、次のようなものを作成、管理できます。

- Google Compute Engine 仮想マシン
- Google Kubernetes Engine クラスタ
- Google Cloud SQL インスタンス

**ヒント**: gcloud コマンドラインツールについての詳細は[こちら](https://cloud.google.com/sdk/gcloud?hl=ja)を参照ください。

### **2. gcloud からの Cloud Run のデフォルト設定**

Cloud Run の利用するリージョン、プラットフォームのデフォルト値を設定します。

```bash
gcloud config set run/region asia-northeast1
gcloud config set run/platform managed
```

ここではリージョンを東京、プラットフォームをフルマネージドに設定しました。この設定によりgcloud コマンドから Cloud Run を操作するときに毎回指定する必要がなくなります。
なお、他のサービスにおいてもリージョンを指定する箇所は度々登場します。
今回のハンズオンではリージョンは可能な限り`asia-northeast1`を利用します。

<walkthrough-footnote>CLI（gcloud）で利用するプロジェクトの指定、Cloud Run のデフォルト値の設定が完了しました。次にハンズオンで利用する機能（API）を有効化します。</walkthrough-footnote>

## **参考: Cloud Shell の接続が途切れてしまったときは**

一定時間非アクティブ状態になる、またはブラウザが固まってしまったなどで 「Cloud Shell」 が切れてしまう、またはブラウザのリロードが必要になる場合があります。その場合は次の対応を実施して、チュートリアルを再開してください。

### **1. チュートリアル資材があるディレクトリに移動する**

```bash
cd ~/cloudrun-handson
```

### **2. チュートリアルを開く**

```bash
teachme doc/handson_index.md
```

### **3. gcloud のデフォルト設定**

```bash
gcloud config set run/region asia-northeast1
gcloud config set run/platform managed
REGION=asia-northeast1
```

途中まで進めていたチュートリアルのページまで `[次へ]` ボタンを押し、進めてください。

## **Google Cloud 環境設定**

Google Cloud では利用したい機能（API）ごとに、有効化を行う必要があります。
ここでは、以降のハンズオンで利用する機能を事前に有効化しておきます。

```bash
gcloud services enable \
artifactregistry.googleapis.com \
run.googleapis.com \
cloudbuild.googleapis.com \
container.googleapis.com \
secretmanager.googleapis.com \
cloudscheduler.googleapis.com \
clouddeploy.googleapis.com \
servicenetworking.googleapis.com \
sqladmin.googleapis.com
```

**GUI**: [API ライブラリ](https://console.cloud.google.com/apis/library)

<walkthrough-footnote>必要な機能が使えるようになりました。次に実際にCloud Runアプリケーションをデプロイしていきます。</walkthrough-footnote>

## **サンプルアプリケーション**

今回利用するサンプルアプリケーションは非常にシンプルなものです。

- フロントエンドアプリケーション
  - HTTPリクエストを受け取り、テキストを返却するアプリケーションです。
- バックエンドアプリケーション
  - HTTPリクエストを受け取り、JSON形式の応答を返却するアプリケーションです。
- バッチアプリケーション
  - データベースの通知データを取得し、更新をするアプリケーションです。

### **フォルダ、ファイル構成**

フォルダ構成には、フロントエンド、バックエンド、バッチアプリケーションの3つのアプリケーションが含まれています。

```bash
.
├── app
│ ├── backend
│ ├── batch
│ └── frontend
│     ├── Dockerfile
│     ├── cloudbuild_push.yaml
│     ├── cloudrun.yaml
│     ├── main.go
├── doc：ハンズオンコンテンツ
└── infra
    ├── json：設定ファイルとして利用あり
    └── sampleapp：利用しない
```

<walkthrough-footnote>次に実際にアプリケーションを Cloud Run にデプロイします。</walkthrough-footnote>

## **アプリケーションをCloud Runにデプロイ**

<walkthrough-enable-apis apis="artifactregistry.googleapis.com,run.googleapis.com"></walkthrough-enable-apis>

<walkthrough-tutorial-duration duration=5></walkthrough-tutorial-duration>

Cloud Runでは、さまざまな方法でデプロイができます。今回のハンズオンでは、Dockerfileをベースにコンテナイメージを作成し、コンテナリポジトリにプッシュします。

### **準備**

GUI を操作し Cloud Run の管理画面を開いておきましょう。

<walkthrough-menu-navigation sectionId="SERVERLESS_SECTION"></walkthrough-menu-navigation>

見つからない場合は次のリンクから開くか、画面真ん中上部の検索から遷移をしてください。

<walkthrough-path-nav path="https://console.cloud.google.com/run" >Cloud Run に移動</walkthrough-path-nav>

以降の手順で Cloud Run の管理画面は何度も開くことになるため、ピン留め (Cloud Run メニューにマウスオーバーし、ピンのアイコンをクリック) しておくと便利です。

### **1. アプリケーション用リポジトリを作成（Artifact Registry）**

```bash
gcloud artifacts repositories create cnsrun-app --repository-format=docker --location=asia-northeast1 --description="Docker repository for the-cloud-run app"
```

### **2. docker コマンドの認証設定**

```bash
gcloud auth configure-docker asia-northeast1-docker.pkg.dev --quiet
```

### **3. ローカル（Cloud Shell 上）にコンテナを作成**

```bash
(cd app/frontend && docker build -t asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/frontend:v1 .)
```

**コラム**: カレントディレクトリを変えずに実行するために括弧でくくっています。

### **4. Artifact Registryへプッシュ**

作成したコンテナをコンテナレジストリ（Artifact Registry）へ登録（プッシュ）します。

```bash
docker push asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/frontend:v1
```

### **5. サービスアカウントの作成**

Cloud Runに割り当てるサービスアカウントを作成します。

```bash
gcloud iam service-accounts create cnsrun-app-frontend --display-name "Service Account for cnsrun-frontend"
```

### **6. Cloud Run にデプロイ**

ようやくデプロイまでの準備が整いました。Cloud Run にデプロイをしましょう。

```bash
gcloud run deploy cnsrun-frontend --image=asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/frontend:v1 \
--allow-unauthenticated \
--service-account=cnsrun-app-frontend
```

デプロイしたCloud Runから応答が返却されることを確認しましょう。

```bash
FRONTEND_URL=$(gcloud run services describe cnsrun-frontend --format='value(status.url)')
curl $FRONTEND_URL/frontend
```

`Hello cnsrun handson's user:D`が返却されたことが確認できたら、次に進みましょう。

## **Cloud Buildの設定**

<walkthrough-tutorial-duration duration=10></walkthrough-tutorial-duration>

<walkthrough-enable-apis apis="cloudbuild.googleapis.com"></walkthrough-enable-apis>

Cloud Buildのコンソール画面から設定をしましょう。

<walkthrough-menu-navigation sectionId="CLOUD_BUILD_SECTION"></walkthrough-menu-navigation>

見つからない場合は次のリンクから開くか、画面真ん中上部の検索から遷移をしてください。

<walkthrough-path-nav path="https://console.cloud.google.com/cloud-build" >Cloud Build に移動</walkthrough-path-nav>

### **1. GitHubリポジトリの接続**

<walkthrough-info-message>本手順はGitHubアカウントの状況によって画面が異なるケースがあります。ハンズオン手順と異なる場合、画面の表示内容にあわせて設定ください。</walkthrough-info-message>

今回利用するサンプルアプリケーションは、GitHubリポジトリに格納されています。Cloud Build で GitHubリポジトリを利用するために、GitHubリポジトリとの連携を設定します。

1. <walkthrough-spotlight-pointer cssSelector="a[id='cfctest-section-nav-item-CLOUD_BUILD_REPOSITORIES']" validationPath="/cloud-build/.*">リポジトリ</walkthrough-spotlight-pointer>メニューに遷移します。
2. <walkthrough-spotlight-pointer locator="semantic({tab '第 2 世代'})" validationPath="/cloud-build/repositories/2nd-gen">第2世代</walkthrough-spotlight-pointer>のタブを選択して、<walkthrough-spotlight-pointer locator="semantic({button 'ホスト接続を作成'})" validationPath="/cloud-build/repositories/2nd-gen">ホスト接続を作成</walkthrough-spotlight-pointer>よりGitHubリポジトリとの接続を行います。
3. `[新しいホストに接続]`において、プロバイダ`[GitHub]`を選択します。
   - リージョン：`asia-northeast1`
   - 名前：`cnsrun-app-handson`
4. <walkthrough-spotlight-pointer locator="semantic({button '接続'})" validationPath="/cloud-build/connections/create">接続</walkthrough-spotlight-pointer>ボタンを押します。

GitHubのページに遷移をし、Google Cloud Buildに対するPermissionを求められます。
`[Authorize Google Cloud Build]`を押し、許可をします。

Cloud Buildの画面に戻り、`[既存のGitHubインストールの使用]`モーダルが表示されます。
`[インストール]`ボタンを押し、**組織ではなく個人のGitHubアカウント**を選択して `[確認]`を押します。
ホスト接続が作成できたら、次に`[リポジトリをリンク]`を押下します。

先ほど作成したホスト接続を選択し、リポジトリには今回のサンプルアプリケーションを選択して`[リンク]`を押します。

以上でGitHubリポジトリとの接続が完了です。

### **2. サービスアカウントの作成**

まず、Cloud Buildが利用するサービスアカウントを作成しておきます。

```bash
gcloud iam service-accounts create cnsrun-cloudbuild --display-name "Service Account for Cloud Build in cnsrun"
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/cloudbuild.builds.builder
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/logging.logWriter
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/clouddeploy.releaser
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/iam.serviceAccountUser
```

### **3. Cloud Buildの設定**

次に、Cloud Buildの起動対象となる「ソースコードのプッシュ」対象のリポジトリ名を取得します。

```bash
REPO_NAME=$(gcloud beta builds repositories list --connection=cnsrun-app-handson --region=asia-northeast1 --format=json | jq -r .[].name)
```

最後に、Cloud Buildのトリガを作成します。

```bash
gcloud beta builds triggers create github \
--name=cnsrun-frontend-trigger \
--region=asia-northeast1 \
--repository="$REPO_NAME" \
--branch-pattern=^main$ \
--build-config=app/frontend/cloudbuild_push.yaml \
--included-files=app/frontend/** \
--substitutions=_DEPLOY_ENV=main \
--service-account=projects/${GOOGLE_CLOUD_PROJECT}/serviceAccounts/cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com
```

<walkthrough-footnote>第2世代と第1世代でパラメータが微妙に違うので注意が必要です。`--repo-name`や`--repo-owner`は第1世代向けです。https://cloud.google.com/sdk/gcloud/reference/beta/builds/triggers/create/github </walkthrough-footnote>

<walkthrough-spotlight-pointer locator="semantic({link 'トリガー、5/4'})" validationPath="/cloud-build/.*">トリガー</walkthrough-spotlight-pointer>
に遷移をして、作成されていることを確認し、次に進みましょう。

## **Cloud Deploy の設定**

<walkthrough-tutorial-duration duration=10></walkthrough-tutorial-duration>

<walkthrough-enable-apis apis="clouddeploy.googleapis.com"></walkthrough-enable-apis>

まずはコンソールから Cloud Deploy のページに移動します。

<walkthrough-menu-navigation sectionId="CLOUD_DEPLOY_SECTION"></walkthrough-menu-navigation>

見つからない場合は次のリンクから開くか、画面真ん中上部の検索から遷移をしてください。

<walkthrough-path-nav path="https://console.cloud.google.com/deploy" >Cloud Deploy に移動</walkthrough-path-nav>

### **1. サービスアカウントの作成**

まず、Cloud Deployが利用するサービスアカウントを作成しておきます。

```bash
gcloud iam service-accounts create cnsrun-clouddeploy --display-name "Service Account for Cloud Deploy in cnsrun"
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-clouddeploy@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/logging.logWriter
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-clouddeploy@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/clouddeploy.jobRunner
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-clouddeploy@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/iam.serviceAccountUser
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-clouddeploy@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/run.developer
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-clouddeploy@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/storage.objectUser
```

### **2. デリバリーパイプラインの作成**

Cloud Deployではデリバリーパイプラインを作成し、Cloud Runをデプロイ先ターゲットとする設定を作成します。

```bash
APP_TYPE=frontend
sed -e "s/PROJECT_ID/${GOOGLE_CLOUD_PROJECT}/g" doc/clouddeploy.yml | sed -e "s/REGION/asia-northeast1/g" | sed -e "s/SERVICE_NAME/cnsrun-${APP_TYPE}/g" > /tmp/clouddeploy_${APP_TYPE}.yml
gcloud deploy apply --file=/tmp/clouddeploy_${APP_TYPE}.yml --region asia-northeast1
```
<walkthrough-spotlight-pointer cssSelector="[id=cfctest-section-nav-item-delivery_pipelines]">デリバリーパイプライン</walkthrough-spotlight-pointer>、<walkthrough-spotlight-pointer cssSelector="[id=cfctest-section-nav-item-targets]">デプロイ先ターゲット</walkthrough-spotlight-pointer>の設定が完了したことをコンソールから確認して次に進みましょう。

## **フロントエンドアプリケーションを修正**

<walkthrough-tutorial-duration duration=10></walkthrough-tutorial-duration>

CI/CDがうまく機能をして、アプリケーションへの修正がデプロイされることを確認しましょう。
Cloud Buildに接続したGitHubのリポジトリを開き、`app/frontend/main.go`を開いて`http.HandleFunc("/frontend")`の応答を適当な文字列に変更してみましょう。

```go
-   fmt.Fprintf(w, "Hello cnsrun handson's user:D\n")
+   fmt.Fprintf(w, "Hello first hands-on\n")
```

Cloud RunのYAML設定ファイルの設定も少し変更をします。
次のコマンドを実行してください。

```bash
echo ${GOOGLE_CLOUD_PROJECT}
```

編集ファイルがローカル環境にあり、macOSの場合は次のコマンドで更新もできます。
コマンド実行が難しい場合は、下記のファイルの`PROJECT_ID`を手動で自身のプロジェクトIDに置き換えてください。

- `app/frontend/cloudrun.yaml`
- `app/backend/cloudrun.yaml`
- `app/batch/cloudrun.yaml`

```bash
YOUR_PROJECT_ID=`自身のプロジェクトID`
sed -i -e "s/PROJECT_ID/${YOUR_PROJECT_ID}/g" app/frontend/cloudrun.yaml
sed -i -e "s/PROJECT_ID/${YOUR_PROJECT_ID}/g" app/backend/cloudrun.yaml
sed -i -e "s/PROJECT_ID/${YOUR_PROJECT_ID}/g" app/batch/cloudrun.yaml
```

変更を加えたら、リモートブランチへプッシュをして、Cloud Buildの`[履歴メニュー]`から処理が起動したことを確認します。

```bash
git add app
git commit -m "feat: hands-on step2"
git push origin main
```

ビルド完了まで、おおよそ5分ほどかかります。
ビルドが正常終了したら、再度リクエストを発行して修正が反映されたことを確認しましょう。

```bash
FRONTEND_URL=$(gcloud run services describe cnsrun-frontend --region=asia-northeast1 --format='value(status.url)')
curl -i $FRONTEND_URL/frontend
```

## **Cloud Runの前に外部ALBを設定する**

<walkthrough-tutorial-duration duration=15></walkthrough-tutorial-duration>

### **1. 自己署名証明書の作成**

外部ALBに SSL/TLS証明書を紐付ける必要があります。本ハンズオンでは自己所有のドメインに紐づく証明書ではなく、自己署名証明書を利用します。

```bash
openssl genrsa 2048 > private.key
openssl req -new -x509 -days 3650 -key private.key -sha512 -out cnsrun.crt -subj "/C=JP/ST=Kanagawa/L=Yokohama/O=uma-arai/OU=Container/CN=team.bit.uma.arai@gmail.com"
```

次の2つのファイルが作成されたことを確認します。

```bash
ls -ltr
```

- private.key
  - 秘密鍵ファイル
- cnsrun.crt
  - 証明書ファイル

### **2. 外部ALBの作成**

さきほど作成した証明書などのファイルを使い、ロードバランサを作成します。
Google Cloudのロードバランサでは、複数のコンポーネントが登場します。
コンポーネント間の関連性については書籍を参照ください。
ハンズオン資料ではロードバランサを作成する箇所に注力します。

まずは、外部ALBに紐づけるグローバルIPアドレスを生成します。

```bash
gcloud compute addresses create --global cnsrun-ip
```

次に、ロードバランサの処理をどのサービスに振り分けるかを定義するバックエンドサービスを作成します。

```bash
gcloud compute backend-services create --global cnsrun-backend-services \
--load-balancing-scheme EXTERNAL_MANAGED
```

バックエンドサービスへの振り分け設定を作成します。

```bash
gcloud compute url-maps create cnsrun-urlmaps \
  --default-service=cnsrun-backend-services
```

振り分け設定を紐づけるためのHTTPプロキシを作成します。
HTTPSプロキシにはSSL証明書を紐づける必要があるため、このタイミングでSSL証明書も作成します。

```bash
gcloud compute ssl-certificates create cnsrun-certificate \
  --certificate ./cnsrun.crt --private-key ./private.key --global
gcloud compute target-https-proxies create cnsrun-https-proxies \
  --ssl-certificates=cnsrun-certificate \
  --url-map=cnsrun-urlmaps
```

最後に払い出したグローバルIPアドレスとHTTPSプロキシを紐づけたロードバランサを作成します。


```bash
gcloud compute forwarding-rules create --global cnsrun-lb \
--target-https-proxy=cnsrun-https-proxies \
--address=cnsrun-ip \
--load-balancing-scheme=EXTERNAL_MANAGED \
--ports=443
```

<walkthrough-footnote>`--load-balancing-scheme=EXTERNAL`とするとクラシックバージョンのアプリケーションロードバランサになります。</walkthrough-footnote>

### **3. NEG の作成、バックエンドサービスへの追加**

```bash
gcloud beta compute network-endpoint-groups create cnsrun-app-neg-asia-northeast1 \
    --region=asia-northeast1 \
    --network-endpoint-type=SERVERLESS \
    --cloud-run-service=cnsrun-frontend

gcloud beta compute backend-services add-backend --global cnsrun-backend-services \
    --network-endpoint-group-region=asia-northeast1 \
    --network-endpoint-group=cnsrun-app-neg-asia-northeast1
```

## **フロントエンドアプリケーションの修正**

現状では、Cloud Runが払い出した`"*.run.app"`のURLを利用してアプリケーションにアクセスができます。
しかし、外部ALBをCloud Runに紐づけたため、外部ALBを利用した経路以外からのアクセスができないようにします。
フロントエンドアプリケーションの`app/frontend/cloudrun.yaml`に次の修正をします。

```patch
- run.googleapis.com/ingress: all
+ run.googleapis.com/ingress: internal-and-cloud-load-balancing
```

Cloud Buildを起動して、Cloud Runに定義を反映させます。

```bash
git add app/frontend/cloudrun.yaml
git commit -m "feat: hands-on step3"
git push origin main
```

## **外部ALBの疎通確認**

フロントエンドアプリケーションへアクセスをして、HTTPSロードバランサを利用した経路のみでアクセスできることを確認しましょう。
まずは、Cloud Runから払い出されたURLではアクセスができないことを確認します。

```bash
FRONTEND_URL=$(gcloud run services describe cnsrun-frontend --region=asia-northeast1 --project=${GOOGLE_CLOUD_PROJECT} --format='value(status.url)')
curl -i $FRONTEND_URL/frontend
```

404エラーが返ってくることを確認します。
次に外部ALBに紐づけたグローバルIPアドレスからはアクセスができることを確認します。
なお、自己署名証明書が有効化されるまでは時間がかかる場合があります。

そのため、定期的にAPIリクエストを投げて正常応答が返ってくるかを確認しましょう。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
```

```bash
watch -n 5 curl -sk https://$LB_GLOBAL_IP/frontend
```

正常に応答が返ってくるようになればOKです。

## **VPCネットワークの作成**

<walkthrough-tutorial-duration duration=5></walkthrough-tutorial-duration>

次は、バックエンドアプリケーションに向けた事前準備となります。
バックエンドアプリケーションを内部通信しか受け付けないようにする事前準備としてVPCを作成します。

<walkthrough-menu-navigation sectionId="VIRTUAL_NETWORK_SECTION "></walkthrough-menu-navigation>

見つからない場合は次のリンクから開くか、画面真ん中上部の検索から遷移をしてください。

<walkthrough-path-nav path="https://console.cloud.google.com/networking/networks" >VPC に移動</walkthrough-path-nav>

まずは、次のgcloudコマンドでVPCを作成します。

```bash
gcloud compute networks create cnsrun-app \
--description=Virtual\ Private\ Network\ for\ the-cloud-run\ hands-on \
--subnet-mode=custom --mtu=1460 --bgp-routing-mode=regional
````

次にサブネットの作成です。

```bash
gcloud compute networks subnets create cnsrun-${GOOGLE_CLOUD_PROJECT} \
--range=10.0.0.0/24 \
--stack-type=IPV4_ONLY \
--network=cnsrun-app \
--region=asia-northeast1 \
--enable-private-ip-google-access
```

VPCとサブネットが作成されたことを確認し、次に進みます。

## **アプリケーション（バックエンド）のデプロイ**

### **1. アプリケーションイメージの登録**

フロントエンドアプリケーション同様、Artifact Registryに対してバックエンドアプリケーションのイメージを登録します。

```bash
(cd app/backend && docker build -t asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/backend:v1 .)
```

```bash
docker push asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/backend:v1
```

### **2. Cloud Build の作成**

Cloud Buildのトリガを作成します。

```bash
REPO_NAME=$(gcloud beta builds repositories list --connection=cnsrun-app-handson --region=asia-northeast1 --format=json | jq -r .[].name)
```

```bash
gcloud beta builds triggers create github \
--name=cnsrun-backend-trigger \
--region=asia-northeast1 \
--repository="$REPO_NAME" \
--branch-pattern=^main$ \
--build-config=app/backend/cloudbuild_push.yaml \
--included-files=app/backend/** \
--substitutions=_DEPLOY_ENV=main \
--service-account=projects/${GOOGLE_CLOUD_PROJECT}/serviceAccounts/cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com
```

### **3. Cloud Deploy でのデプロイ**

```bash
APP_TYPE=backend
sed -e "s/PROJECT_ID/${GOOGLE_CLOUD_PROJECT}/g" doc/clouddeploy.yml | sed -e "s/REGION/asia-northeast1/g" | sed -e "s/SERVICE_NAME/cnsrun-${APP_TYPE}/g" > /tmp/clouddeploy_${APP_TYPE}.yml
gcloud deploy apply --file=/tmp/clouddeploy_${APP_TYPE}.yml --region asia-northeast1
```

### **4. フロントエンドアプリケーションからのリクエストを受け付ける**

バックエンドはフロントエンドアプリケーションからのみアクセスを許可します。
これを実現するために、フロントエンドアプリケーションが利用するサービスアカウントに対して」`roles/run.invoker`の権限を付与します。

<walkthrough-footnote>最小権限を意識するならば、Cloud Runに対してのみ権限を付与するべきです。今回は、少し手順をシンプルにするためプロジェクトリソースに対して権限を設定しています。</walkthrough-footnote>

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-app-frontend@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/run.invoker
```

### **5. バックエンド用のサービスアカウントの作成**

Cloud Runに割り当てるサービスアカウントを作成します。

```bash
gcloud iam service-accounts create cnsrun-app-backend --display-name "Service Account for cnsrun-backend"
```

### **6. デプロイ**

フロントエンドアプリケーションは`gcloud run deploy`コマンドでデプロイをしていました。
せっかくCI/CDパイプラインを作成したので、バックエンドアプリケーションはCloud Buildからデプロイをしておきましょう。

```bash
gcloud builds triggers run cnsrun-backend-trigger \
--region=asia-northeast1 \
--branch=main
```

ビルドが完了したことを確認し、次に進みましょう。

<walkthrough-info-message>ここまででバックエンドアプリケーションの修正は完了です。しかし、これだけではフロントエンドからの通信ができないためフロントエンドアプリケーションにも修正を加えていきます。</walkthrough-info-message>

## **フロントエンドアプリケーションの修正**

### **1. バックエンドアプリケーションへの通信設定**

フロントエンドからバックエンドアプリケーションへ向けた通信の設定をします。
バックエンドアプリケーションへのアクセスURLはCloud Runが発行した`"*.run.app"`のURLを利用します。

```bash
gcloud run services describe cnsrun-backend --format='value(status.url)'
```

このURLを利用して、フロントエンドアプリケーションの`cloudrun.yaml`内にあるコンテナ環境変数を修正します。
`{バックエンドアプリケーションCloud RunのURL}`を先ほど取得したURLに置き換えます。

```patch
-    value: "https://cnsrun-backend-noejq743xa-an.a.run.app" # FIXME: Change BACKEND_FQDN value after backend resources is created
+    value: "{バックエンドアプリケーションCloud RunのURL}"
```

### **2. VPCアクセスの設定**

バックエンドアプリケーションはVPCネットワーク内に配置されているため、VPCネットワークにアクセスするための設定を行います。
先ほどと同様、フロントエンドアプリケーションの`cloudrun.yaml`内にあるネットワーク設定を修正します。
`{生成したサブネット名}`を自身のものに置き換えてください。

```patch
- #         TODO: change for your own vpc name
- #        run.googleapis.com/network-interfaces: '[{"network":"cnsrun-app", "subnetwork":"cnsrun-the-cloud-run"}]'
- #        run.googleapis.com/vpc-access-egress: all-traffic
+          run.googleapis.com/network-interfaces: '[{"network":"cnsrun-app", "subnetwork":"{生成したサブネット名}"}]'
+          run.googleapis.com/vpc-access-egress: all-traffic
```

### **3. アプリケーションのデプロイ**

修正を反映するために、フロントエンドアプリケーションをデプロイします。

```bash
git add app/frontend/cloudrun.yaml
git commit -m "feat: hands-on step4"
git push origin main
```

## **バックエンドアプリケーションの疎通確認**

フロントエンドアプリケーションを介してバックエンドアプリケーションにアクセスができることを確認します。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
curl -k https://$LB_GLOBAL_IP/backend
```

バックエンドアプリケーションを介して**JSON形式の応答を受信**できたらOKです。
次のStepに進みましょう。

## **Step5: DB作成**

Step5では、バックエンドアプリケーションに対して、Cloud SQLを利用したデータベース接続をします。

本書でも記載の通り、**Cloud SQLはGoogleが管理するVPC内に配置されます**。
今回のハンズオンでは、「**プライベートサービスアクセス**」を介してCloud SQLへ接続しましょう。
またDBへ接続するためのパスワードは、「**Secret Manager**」に保存します。

Step5では次の手順で構築を進めます。

- Cloud SQLの作成
  - この手続きの中でプライベートサービスアクセスを作成します。
- Secret ManagerへDB接続情報を登録
- サービスアカウントへの権限付与
- バックエンドアプリケーションの修正

それでは、**Step5: DBと接続をする** に進みましょう。

## **Cloud SQLの作成**

<walkthrough-tutorial-duration duration=30></walkthrough-tutorial-duration>

<walkthrough-menu-navigation sectionId="SQL_SECTION"></walkthrough-menu-navigation>

見つからない場合は次のリンクから開くか、画面真ん中上部の検索から遷移をしてください。

<walkthrough-path-nav path="https://console.cloud.google.com/sql/instances" >Cloud SQL に移動</walkthrough-path-nav>

<walkthrough-spotlight-pointer cssSelector="a[aria-label=インスタンスの作成ページに移動するボタン]" validationPath="/sql/instances">インスタンスを作成</walkthrough-spotlight-pointer> をします。

### **1. エンジンの選択**

<walkthrough-spotlight-pointer locator="semantic({button 'PostgreSQL を選択'})" validationPath="sql/choose-instance-engine">PostgreSQL を選択</walkthrough-spotlight-pointer>を押します。

### **2. インスタンスの作成（基本部分）**

インスタンスの情報を入力します。

- `[インスタンスID]`
  - `cnsrun-app-instance`
- `[パスワード]`
  - `Cnsrun-db-pass-1234`（自由に設定してOKです）
- `[パスワードポリシー]`
  - パスワード ポリシーを有効にする
    - 最小の長さを14に設定
    - 複雑さを要求
    - パスワードにユーザー名を許可しない
- `[データベースのバージョン]`
  - `PostgresSQL 15`

Cloud SQL のエディションの選択します。
本番利用ではプリセットは本番を選択しますが、**料金の関係上、`サンドボックス`**を選択します。

- `[エディション]`
  - `Enterprise`
- `[プリセット]`
  - `サンドボックス`

リージョンとゾーンの可用性の選択します。こちらも通常は`複数のゾーン（高可用性）`を選択しますが、料金の関係上`シングルゾーン`にします。

- `[リージョン]`
  - `asia-northeast1`
- `[ゾーンの可用性]`
  - `シングルゾーン`

### **3. インスタンスの作成（詳細部分）**

インスタンスのカスタマイズでは、`マシンの構成`と`接続`の追加設定をします。
こちらについても、**料金の関係上、最小構成**の選択をします。
実際の構成をする際はワークロードに合わせてチューニングをしましょう。

- `[マシンの構成]`
  - `[マシンシェイプ]`
    - `共有コアマシン`
    - `1vCPU、0.614GB`

接続の箇所で、「プライベートサービスアクセス」の作成をします。

- `[接続]`
  - `[パプリックIP]`
    - `**チェックを外す**`
  - `[プライベートIP]`
    - `チェックを付ける`
    - `[ネットワーク]`
      - `cnsrun-app`

「プライベートサービスアクセス接続は必須です」というWarningが出ているはずです。
`[接続を設定]`ボタンを押して、プライベートサービスアクセスの設定に進みます。

- `[IP範囲を割り振る]`
  - `[1つ以上の既存のIP範囲を選択するか、新しいIP範囲を作成する]` → `新しいIP範囲を割り振る`
    - `[名前]`
      - `cnsrn-cnsrun-private-ip-address`
    - `[IPアドレス範囲]`
      - `10.0.200.0/24`

続行を押して次に進み、<walkthrough-spotlight-pointer locator="semantic({button '接続を作成'})" validationPath="/sql/instances/create">接続を作成</walkthrough-spotlight-pointer> を押して、プライベートサービスアクセスを作成します。作成には少し時間がかかります。

![プライベートサービスアクセス完了](https://github.com/uma-arai/cloudrun-handson/blob/main/images/private-service-access-complete.png?raw=true)

設定を続けます。

- `[Google Cloud サービスの承認]`
  - `プライベート パスを有効にする`を有効化

<walkthrough-spotlight-pointer locator="semantic({button 'インスタンスを作成'})" validationPath="/sql/instances/create">インスタンスを作成</walkthrough-spotlight-pointer> ボタンを押しましょう。

インスタンスの作成には時間がかかります。作成が**進行していること**を確認し、次のステップに進みます。

## **Secret ManagerへDB接続情報を登録**

Cloud SQLインスタンスの作成を待っている間に進められる箇所を進めます。
バックエンドアプリケーションがCloud SQLに接続するための情報をSecret Managerに登録しましょう。
`DB_PASSWORD`はパスワードポリシーにあうような値ならば何でもOKです。下記コマンドに設定している値をそのまま利用してもOKです。

```bash
DB_PASSWORD=DB-user-pass-1234
echo -n "$DB_PASSWORD" | gcloud secrets create cnsrun-app-db-password \
--replication-policy=user-managed \
--locations=asia-northeast1 \
--data-file=-
```

<walkthrough-footnote>`--data-file=-`を指定することでバージョン=1のシークレットが登録されます。</walkthrough-footnote>

## **サービスアカウントへの権限付与(Secret Manager)**

Secret Managerに登録したシークレットにアクセスするための権限を、Cloud Runに設定したサービスアカウントに付与します。
データベースに接続するのはバックエンドアプリケーションのため、`cnsrun-app-backend`サービスアカウントに権限を付与します。

```bash
gcloud secrets add-iam-policy-binding cnsrun-app-db-password \
--member=serviceAccount:cnsrun-app-backend@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
--role=roles/secretmanager.secretAccessor
```

<walkthrough-footnote>これで、バックエンドアプリケーションがCloud SQLに接続するための準備が整いました。</walkthrough-footnote>

## **サービスアカウントへの権限付与(Cloud SQL)**

同様に、バックエンドアプリケーションがCloud SQLに接続するための権限をサービスアカウントに付与します。

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
--member=serviceAccount:cnsrun-app-backend@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
--condition=None \
--role=roles/cloudsql.client
```

## **データの作成**

### **1. データベースの作成**

作成されたインスタンスに、データベースを作成します。
インスタンスが作成されたことを確認し、インスタンスの詳細画面に遷移をします。

1. <walkthrough-spotlight-pointer locator="semantic({listitem} {link 'データベース、10/7'})" validationPath="/sql/.*">データベース</walkthrough-spotlight-pointer>メニューを選択します。
2. <walkthrough-spotlight-pointer locator="semantic({button 'データベースの作成'})" validationPath="/sql/instances/.*">データベースの作成</walkthrough-spotlight-pointer>ボタンを押します。
3. <walkthrough-spotlight-pointer cssSelector="[formcontrolname=databaseName]" validationPath="/sql/instances/.*">データベース名</walkthrough-spotlight-pointer>として、`cnsrun`を入力します。
4. <walkthrough-spotlight-pointer cssSelector="button[type=submit]" validationPath="/sql/.*">作成</walkthrough-spotlight-pointer>を押します。

### **2. ユーザーの作成**

続けてデータベースにアクセスするユーザーを作成します。

1. <walkthrough-spotlight-pointer locator="semantic({listitem} {link 'ユーザー、10/6'})" validationPath="/sql/.*">ユーザー</walkthrough-spotlight-pointer>メニューを選択します。
2. <walkthrough-spotlight-pointer locator="semantic({button 'ユーザー アカウントを追加'})" validationPath="/sql/.*">ユーザーアカウントを追加</walkthrough-spotlight-pointer>ボタンを押します。
3. 組み込み認証を選択します。
4. ユーザー名は`app`を入力します。
5. ポリシーに沿って入力をします。以前に`DB_PASSWORD`としてSecret Managerに設定した値（例：DB-app-pass-1234）にします。。
6. <walkthrough-spotlight-pointer cssSelector="button[type=submit]" validationPath="/sql/.*">追加</walkthrough-spotlight-pointer>を押します。

### **3. テーブル作成**

Cloud SQLには、コンソールにある「**Cloud SQL Studio**」から直接データベースにログインができます。

作成したインスタンスの詳細画面に遷移をして、<walkthrough-spotlight-pointer cssSelector="#cfctest-section-nav-item-studio">Cloud SQL Studio</walkthrough-spotlight-pointer>に進みます。

- `[データベース]`
  - `cnsrun-app`
- `[ユーザー]`
  - `app`
- `[パスワード]`
  - `{「2. ユーザーの作成」で設定したパスワード}`

次のSQLを`[エディタ１]`のテキストエリアに記述し、<walkthrough-spotlight-pointer spotlightId="fr-query-run-button">SQLコマンドを実行</walkthrough-spotlight-pointer>してテーブルを作成します。

```postgresql
CREATE TABLE "notification"
(
    "id"        SERIAL         PRIMARY KEY,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "isRead"    BOOLEAN      NOT NULL DEFAULT false,
    "isDeleted" BOOLEAN      NOT NULL DEFAULT false,
    "verification" BOOLEAN   NOT NULL DEFAULT false,
    "email"    varchar(255)         NOT NULL,
    "body"      TEXT
);

CREATE TABLE "users"
(
    "id"        SERIAL         PRIMARY KEY,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "isDeleted" BOOLEAN      NOT NULL DEFAULT false,
    "email"    varchar(255)         NOT NULL
);
CREATE INDEX notification_email_idx ON "notification" (email);
```

<walkthrough-info-message>`user`はPostgreSQLのユーザテーブルと重複するので、`users`としています。</walkthrough-info-message>

### **4. データの登録**

Cloud SQL Studioで続けてデータを登録します。

次のSQLをテキストエリアに記述し、<walkthrough-spotlight-pointer spotlightId="fr-query-run-button" validationPath="/sql/instances/cnsrun-app-.*/studio">SQLコマンドを実行</walkthrough-spotlight-pointer>してテーブルを作成します。

```postgresql
-- Insert 1
INSERT INTO "notification" (email, body)
VALUES ('user1@example.com', 'ご注文の商品が出荷されました！');

-- Insert 2.
INSERT INTO "notification" (email, body)
VALUES ('user1@example.com', 'uma-araiショップから新しいメッセージが届きました');

-- Insert 3
INSERT INTO "notification" (email, body)
VALUES ('user1@example.com', 'THE CLOUD RUNの予約が確定しました');

-- Insert 4
INSERT INTO "notification" (email, body)
VALUES ('user2@example.com', 'あなたのアカウントがuma-araiショップに正常にリンクされました');

-- Insert 5
INSERT INTO "notification" (email, body)
VALUES ('user2@example.com', 'THE CLOUD RUN が再入荷しました！ 売り切れる前にゲットしましょう');

-- User Insert
INSERT INTO "users" ("email") VALUES ('user1@example.com');
```

データが登録されたことを確認しましょう。

```postgresql
SELECT * FROM
  "notification" LIMIT 1000;
```

SELECT文を<walkthrough-spotlight-pointer spotlightId="fr-query-run-button">実行</walkthrough-spotlight-pointer>して、データが登録されていることを確認します。

## **バックエンドアプリケーションの修正**

最後に、作成したDBへ接続するための設定をバックエンドアプリケーションに追加します。


```bash
gcloud sql instances describe cnsrun-app-instance --format='value(ipAddresses[0].ipAddress)'
```

コマンドを実行した結果、Cloud SQLのプライベートIPアドレスが表示されます。
この値を利用して、バックエンドアプリケーションの`app/backend/cloudrun.yaml`を修正します。

- DBへ接続するための各種設定
  - `env`ブロックのコメントアウトを外す
  - `DB\_HOST`を上記のIPアドレスに変更
    - `DBのIPアドレス`には、gcloudコマンドで取得できるCloud SQLのIPアドレスを指定します。

```patch
- #        env:
- #          - name: DB_USER
- #            value: "app"
- #          - name: DB_PASSWORD
- #            valueFrom:
- #              secretKeyRef:
- #                key: latest
- #                name: cnsrun-db-password
- #          - name: DB_HOST
- #            value: "10.0.200.3"  # FIXME: Change DB_HOST value to actual private IP address after Cloud SQL created
- #          - name: DB_PORT
- #            value: "5432"
- #          - name: DB_NAME
- #            value: "cnsrun"
+         env:
+           - name: DB_USER
+             value: "app"
+           - name: DB_PASSWORD
+             valueFrom:
+               secretKeyRef:
+                 key: latest
+                 name: cnsrun-db-password
+           - name: DB_HOST
+             value: `DBのIPアドレス`
+           - name: DB_PORT
+             value: "5432"
+           - name: DB_NAME
+             value: "cnsrun"
```

YAML設定ファイルを修正後、コードをプッシュして、Cloud Build経由でバックエンドアプリケーションをデプロイしましょう。

**注意：** YAMLファイルはインデントが重要です。修正を行う際は、インデントが正しいことを確認してください。筆者も疎通中何度もミスりました。。。


## **DB接続ができることを確認**

バックエンドアプリケーションがCloud SQLに接続できることを確認します。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
curl -k https://$LB_GLOBAL_IP/backend/notification?id=1
```

JSON形式の応答が返ってくればOKです。

## **Cloud Runジョブを利用する**

<walkthrough-tutorial-duration duration=15></walkthrough-tutorial-duration>

最後に、Cloud Runジョブを利用して、定期的にDBに更新をかけるジョブを作成します。
次の流れで構築をします。

- Cloud Runジョブの作成
- 定期実行のためのスケジューラ設定

### **1. アプリケーションイメージの登録**
フロントエンドアプリケーション同様、Artifact Registryに対してジョブのイメージを登録します。

```bash
(cd app/batch && docker build -t asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/batch:v1 .)
```

```bash
docker push asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/batch:v1
```


### ***2. サービスアカウントの作成**

Cloud Runジョブに割り当てるサービスアカウントを作成します。

```bash
gcloud iam service-accounts create cnsrun-app-batch \
 --display-name "Service Account for cnsrun-batch"
```

### **3. Cloud Runジョブのデプロイ**

Cloud Buildから作成してもいいのですが、手動で作成しておく体験をするために、ここでは手動でジョブを作成します。
  
```bash
gcloud run jobs deploy cnsrun-batch \
--image=asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/batch:v1 \
--region=asia-northeast1 \
--service-account=cnsrun-app-batch \
--parallelism=1 \
--execute-now
```

`--execute-now`を指定することで、ジョブは即時実行されます。
環境変数やネットワーク設定をしていないためジョブは失敗します。
ジョブ実行が失敗していることを確認しましょう。

### **4. Cloud Build の作成**

CI/CDの設定のためCloud Buildのトリガを作成します。

```bash
REPO_NAME=$(gcloud beta builds repositories list --connection=cnsrun-app-handson --region=asia-northeast1 --format=json | jq -r .[].name)
```

```bash
gcloud beta builds triggers create github \
--name=cnsrun-batch-trigger \
--region=asia-northeast1 \
--repository="$REPO_NAME" \
--branch-pattern=^main$ \
--build-config=app/batch/cloudbuild_push.yaml \
--included-files=app/batch/** \
--substitutions=_DEPLOY_ENV=main \
--service-account=projects/${GOOGLE_CLOUD_PROJECT}/serviceAccounts/cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com
```

### **5. Cloud Deploy の作成**

こちらも同様に作成します。

```bash
APP_TYPE=batch
sed -e "s/PROJECT_ID/${GOOGLE_CLOUD_PROJECT}/g" doc/clouddeploy.yml | sed -e "s/REGION/asia-northeast1/g" | sed -e "s/SERVICE_NAME/cnsrun-${APP_TYPE}/g" > /tmp/clouddeploy_${APP_TYPE}.yml
gcloud deploy apply --file=/tmp/clouddeploy_${APP_TYPE}.yml --region asia-northeast1
```

足回りができました。次に進みましょう。

## **ジョブのYAML設定ファイル修正**

バックエンドアプリケーション同様、Cloud SQLに接続するための設定をします。
`DBのIPアドレス`には、次のgcloudコマンドで取得できるCloud SQLのIPアドレスを指定します。

```bash
gcloud sql instances describe cnsrun-app-instance --format='value(ipAddresses[0].ipAddress)'
```

- DBへ接続するための各種設定
  - `DB\_HOST`を上記のIPアドレスに変更
  - `DBのIPアドレス`には、gcloudコマンドで取得できるCloud SQLのIPアドレスを指定します。

```patch
             - name: DB_HOST
-             value: "10.0.200.3"  # FIXME: Change DB_HOST value to actual private IP address after Cloud SQL created
+             value: `DBのIPアドレス`
```

YAML設定ファイルを修正後、コードをプッシュして、Cloud Build経由でバックエンドアプリケーションをデプロイしましょう。

## **サービスアカウントへの権限追加**

```bash
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
  --member=serviceAccount:cnsrun-app-batch@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
  --condition=None \
  --role=roles/run.invoker
gcloud projects add-iam-policy-binding ${GOOGLE_CLOUD_PROJECT} \
--member=serviceAccount:cnsrun-app-batch@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
--condition=None \
--role=roles/cloudsql.client
gcloud secrets add-iam-policy-binding cnsrun-app-db-password \
--member=serviceAccount:cnsrun-app-batch@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com \
--role=roles/secretmanager.secretAccessor
```

## **定期実行のためのスケジューラ設定**

最後に、ジョブを定期実行するためのスケジューラを設定します。

```bash
gcloud scheduler jobs create http cnsrun-batch-job-scheduler \
  --location=asia-northeast1 \
  --schedule="* * * * *" \
  --uri="https://asia-northeast1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${GOOGLE_CLOUD_PROJECT}/jobs/cnsrun-batch:run" \
  --http-method POST \
  --time-zone=Asia/Tokyo \
  --attempt-deadline=5m \
  --description="Run the batch job every minute" \
  --oauth-service-account-email=cnsrun-app-batch@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com
```

## **Cloud Runジョブの動作確認**

Cloud Runジョブによって通知テーブルのデータが更新されていることを確認します。
ジョブではUsersテーブルに登録されているメールアドレスの通知に対して、`verification`フラグを立てる処理を行っています。

フロントエンドアプリケーションから取得する通知の内容が変化していることを確認します。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
curl -k https://$LB_GLOBAL_IP/backend/notification?id=1
```

`verification`が `true`となっていることを確認できたらOKです。

## **お疲れ様でした！**

<walkthrough-conclusion-trophy></walkthrough-conclusion-trophy>

ここまでで、つぎの内容を実装してきました。

- 自身で構築したコンテナイメージをCloud Runで動かす
- CDパイプラインを構築する
- Cloud Runの前に外部ALBを設定する
- 複数のCloud Runを連携する
- データベースと接続をする
- Cloud Runジョブを利用する


次のハンズオンでは、ここまで構築した構成をプロダクションレディにしていきます。

- このまま次に進む方
  - リソースはそのまま残して次のハンズオンを起動しましょう。

```bash
teachme doc/handson_chap6.md
```

- いったん一区切りとしてストップする方
  - 後日に次章に取り組む場合、課金を防ぐためにこのままリソースの削除をすることをお勧めします。

```bash
teachme doc/deletion/chap5_deletion.md
```
