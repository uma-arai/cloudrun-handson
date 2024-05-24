# **The Cloud Run ハンズオン <br />（実践編）**

## 概要

ハンズオン（実践編）で構築する全体概要図は[こちらのリンク](https://github.com/uma-arai/cloudrun-handson/blob/main/images/06-handson-architecture-overview.png?raw=true)となります。

本ハンズオンは事前に5章までのハンズオンを完了していることを前提としています。
今回のハンズオンではリージョンは可能な限り`asia-northeast1`を利用します。
また、ハンズオンを実行するユーザは、基本ロールである**Owner権限を持つプロジェクト**を利用してください。

では、ハンズオンをはじめましょう。

## **準備：Google Cloud 環境設定**

Google Cloud では利用したい機能（API）ごとに、有効化を行う必要があります。
ここでは、6章のハンズオンで利用する機能を事前に有効化しておきます。

```bash
gcloud services enable \
domains.googleapis.com \
dns.googleapis.com \
containeranalysis.googleapis.com \
cloudkms.googleapis.com \
binaryauthorization.googleapis.com
```

それではハンズオンを進めていきましょう。

## **Cloud RunへのSSL/TLS接続をTLS1.2以上にする**

<walkthrough-tutorial-duration duration=5></walkthrough-tutorial-duration>

まずは、フロントエンドアプリケーションへのSSL/TLS接続の強化です。
現状のロードバランサへのアクセスでTLS 1.0のプロトコルを利用して疎通をするか確認します。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
curl -k -v --tls-max 1.0 https://${LB_GLOBAL_IP}/frontend
```

`TLSv1.0 (IN), TLS handshake`と表示されてTLS 1.0で通信が行われているにもかかわらず、HTTP200の応答が得られていることを確認します。

**ヒント：**2024年5月23日時点の筆者のCloud shellから実行した場合、TLS 1.0を指定した`curl`コマンドがエラーになりました。ローカル環境からも試してみてください。

では、これを防ぐ設定を入れていきましょう。

### **SSLポリシーの作成**

まずは、SSLポリシーを作成します。

```bash
gcloud compute ssl-policies create cnsrun-ssl-policy --profile=RESTRICTED --min-tls-version=1.2 --global
```

### **ロードバランサにSSLポリシーを適用**

作成したSSLポリシーをロードバランサに適用します。

```bash
gcloud compute target-https-proxies update cnsrun-https-proxies \
    --ssl-policy cnsrun-ssl-policy \
    --global
```

ポリシーがロードバランサに反映されるまで時間がかかります。
十分に時間をおいた後、次のテストでTLS1.0での通信ができないか確認しましょう。

### **テスト**

TLS1.0での通信可否を確認します。

```bash
curl -k -v --tls-max 1.0 https://${LB_GLOBAL_IP}/frontend
```

`tlsv1 alert protocol version`と表示されエラーとなることが確認できたらOKです。
続けて、TLS1.2まで利用OKの状態でリクエストを送信してみましょう。

```bash
curl -k -v --tls-max 1.2 https://${LB_GLOBAL_IP}/frontend
```

TLS1.2でTLS Handshakeが実行され、SSLコネクションを確立できたことを確認できます。
HTTP200が返却されていれば設定完了です。

## **コンテナセキュリティの強化（Artifact Registryの脆弱性スキャン）**

<walkthrough-tutorial-duration duration=5></walkthrough-tutorial-duration>

次に、コンテナイメージのセキュリティを強化します。
Artifact Registryには、コンテナイメージの脆弱性スキャン機能があります。
これらを有効化する方法は非常にシンプルです。

Google Cloud ConsoleからArtifact Registryのページに移動して機能をONにするだけでプロジェクト全体での脆弱性スキャンが有効化されます。

<walkthrough-watcher-block link-url="https://console.cloud.google.com/artifacts"> Artifact Registry に移動</walkthrough-watcher-block>

<walkthrough-spotlight-pointer cssSelector="[id=cfctest-section-nav-item-settings]" validationPath="/artifacts/settings">設定</walkthrough-spotlight-pointer> に移動します。

`[スキャン:オフ]`となっていることを確認し、<walkthrough-spotlight-pointer cssSelector="[cfciamcheck='servicemanagement.services.bind']" validationPath="/artifacts/settings" > 有効にする </walkthrough-spotlight-pointer> をクリックします。

`[スキャン:オン]`となれば完了です。非常にシンプルですね。

イメージスキャンは、コンテナイメージをArtifact Registryにプッシュする際に自動的に実行されます。
フロントエンドアプリケーションの新しいイメージをプッシュして、スキャンが実行されることを確認しましょう。


```bash
(cd app/frontend && touch dummy && docker build -t asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/frontend:v2 .)
```

```bash
docker push asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/frontend:v2
```

1. <walkthrough-spotlight-pointer cssSelector="[id=cfctest-section-nav-item-repositories]" validationPath="/artifacts">リポジトリ</walkthrough-spotlight-pointer> に移動します。
2. `[cnsrun-app]` -> `[frontend]`の順に遷移をします。
3. `[脆弱性]`列を見ます。古いイメージは`スキャンされません`となっていますが、新しいイメージはスキャンされていることが確認できます。

以上で、Artifact Registryによるコンテナイメージの脆弱性スキャンが完了しました。
やはり、非常にシンプルですね。

## **Cloud RunとWAFを組み合わせる**

<walkthrough-tutorial-duration duration=10></walkthrough-tutorial-duration>

一般ユーザーからのアクセスを受け付ける場合、通常のユーザー以外に悪意のあるユーザーからのアクセスは必ず発生します。
Google Cloudでは、それらに対するサービスとして[Cloud Armor](https://cloud.google.com/security/products/armor)があります。

前回のハンズオンで作成した外部ALBに対し、Cloud Armorを設定してDDoS攻撃などから保護しましょう。

<walkthrough-path-nav path="https://console.cloud.google.com/net-security/securitypolicies/list"> Cloud Armorページに移動</walkthrough-path-nav>

## **セキュリティポリシーの作成**

Cloud Armorを利用するためには、セキュリティポリシーを作成する必要があります。
セキュリティポリシーに対して、どのルールを適用するか、どのターゲットにポリシーをアタッチするかを設定します。

さっそく作成に移りましょう。


```bash
SECURITY_POLICY_NAME=cnsrun-waf-policy
gcloud compute security-policies create $SECURITY_POLICY_NAME 
```
## **ルールの作成**

ルールがなければ、セキュリティポリシーは動作しません。
次に、セキュリティポリシーに適用するルールを作成します。

<walkthrough-info-message>正確にはすべてを通すルールがアタッチされてはいます</walkthrough-info-message>

作成するルールの内容は次の通りです。

- SQLインジェクションを拒否
- クロスサイトスクリプティングを拒否
- Local file inclusion（LFI）を拒否
- Remote file inclusion（RFI）を拒否
- リモートコード実行を拒否
- メソッド絞り込み
- スキャン検知
- プロトコルアタックを拒否
- セッション固定攻撃を拒否
- Log4jの脆弱性攻撃を拒否（cve canary）
- JSONベースのSQLインジェクションバイパスの脆弱性攻撃を拒否

順番ずつ作成していきましょう。

```bash
gcloud compute security-policies rules create 1001 \
--security-policy $SECURITY_POLICY_NAME  \
--description "SQL injection" \
--expression "evaluatePreconfiguredExpr('sqli-v33-stable')" \
--action=deny-403
```

```bash
gcloud compute security-policies rules create 1002 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Cross-site scripting" \
--expression "evaluatePreconfiguredExpr('xss-v33-stable')" \
--action=deny-403
```

Cloud Armorのコンソール画面を確認し、ルールが追加されていることも是非確認してください。
とはいえ、一つずつは手間になってきましたね。一気に作成しちゃいましょう。

```bash
gcloud compute security-policies rules create 1003 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Local file inclusion" \
--expression "evaluatePreconfiguredExpr('lfi-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1004 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Remote file inclusion" \
--expression "evaluatePreconfiguredExpr('rfi-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1005 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Remote code execution" \
--expression "evaluatePreconfiguredExpr('rce-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1006 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Method enforcement" \
--expression "evaluatePreconfiguredExpr('methodenforcement-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1007 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Scanner detection" \
--expression "evaluatePreconfiguredExpr('scannerdetection-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1008 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Protocol attack" \
--expression "evaluatePreconfiguredExpr('protocolattack-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1009 \
--security-policy $SECURITY_POLICY_NAME  \
--description "Session fixation attack" \
--expression "evaluatePreconfiguredExpr('sessionfixation-v33-stable')" \
--action=deny-403

gcloud compute security-policies rules create 1101 \
--security-policy $SECURITY_POLICY_NAME  \
--description "cve-canary" \
--expression "evaluatePreconfiguredExpr('cve-canary')" \
--action=deny-403

gcloud compute security-policies rules create 1102 \
--security-policy $SECURITY_POLICY_NAME  \
--description "json-sqli-canary" \
--expression "evaluatePreconfiguredExpr('json-sqli-canary')" \
--action=deny-403
```

**注意：** 追加可能なセキュリティポリシールールの数には制限があり、なかなか厳しい制限となっています（デフォルト：20）。上限に達した場合、次のエラーが表示され、クオータの引き上げ申請が必要となります。

```
 - Quota 'SECURITY_POLICY_CEVAL_RULES' exceeded.  Limit: 20.0 globally.
ERROR: (gcloud.compute.security-policies.rules.create) Could not fetch resource:
        metric name = compute.googleapis.com/security_policy_ceval_rules
        limit name = SECURITY-POLICY-CEVAL-RULES-per-project
        limit = 20.0
        dimensions = global: global
Try your request in another zone, or view documentation on how to increase quotas: https://cloud.google.com/compute/quotas.
```

<walkthrough-footnote>ルール一覧：https://cloud.google.com/armor/docs/waf-rules</walkthrough-footnote>


## **ターゲットの設定**

セキュリティポリシーを適用するターゲットを設定します。

<walkthrough-footnote>参考：https://cloud.google.com/armor/docs/configure-security-policies?hl=ja#attach-policies</walkthrough-footnote>

```bash
BACKEND_SERVICE_NAME=$(gcloud compute backend-services list --format=json | jq -r .[].name | grep cnsrun)
gcloud compute backend-services update $BACKEND_SERVICE_NAME \
    --security-policy $SECURITY_POLICY_NAME \
    --global
```

## **Cloud Armorの設定チェック**

セキュリティポリシーが正しく設定されているか、テストしましょう。

まず正常系が通るか確認します。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
```

```bash
curl -i -k https://$LB_GLOBAL_IP/backend?id=none
# HTTP200 OK
```

HTTP200が正しく返却されます。

次に異常系です。
今回はクエリストリングに対して、XSSを発生させるようなリクエストを送信します。

<walkthrough-info-message>設定がバックエンドサービスに反映されるまで、少し時間がかかります。HTTP200が応答される場合、しばらく時間をおきましょう。</walkthrough-info-message>


```bash
curl -i -k https://$LB_GLOBAL_IP/backend?id="<script>alert('XSS')</script>"
# HTTP403 Forbidden
```

想定通り、HTTP403 Forbiddenが返却されました。

同様にSQL Injectionも試してみましょう。

```bash
curl -i -k https://$LB_GLOBAL_IP/backend?id="foo\%27\%20OR\%20bar\%27\%3D\%27bar"
# HTTP403 Forbidden
```

こちらも想定通りにHTTP403 Forbiddenが返却されました。

---

このように、Cloud Armorを利用することで一般的なセキュリティ攻撃に対する対策を簡易にできます。
ただし、前述したようにCloud Runに直接Cloud Armorを設定はできないため、外部ALBを利用する必要がある点には注意をしておきましょう。

## **レジストリ内のイメージ保持設定**

<walkthrough-tutorial-duration duration=5></walkthrough-tutorial-duration>

まず、Artifact Registryのクリーンアップポリシーを確認します。

```bash
gcloud artifacts repositories describe cnsrun-app --location=asia-northeast1 --format=json | jq .cleanupPolicies
```

最終行が`null`であれば、クリーンアップポリシーが設定されていないことを示しています。

次に、クリーンアップポリシーを設定します。
条件付き削除、条件付き保持、最新イメージ保持のポリシーを1つのJSONに記述をして、そのJSONファイルをインプットとして指定します。

```bash
gcloud artifacts repositories set-cleanup-policies cnsrun-app \
--location=asia-northeast1 \
--policy=./infra/json/cleanup-policy.json
```

再度、Artifact Registryのクリーンアップポリシーを確認します。

```bash
gcloud artifacts repositories describe cnsrun-app --location=asia-northeast1 --format=json | jq .cleanupPolicies
```

クリーンアップポリシーが表示されていることを確認できればOKです。
残念ながら、2024年5月現在ではクリーンアップポリシーを手動で実行することができません。
Google Cloud側で自動実行され、頻度はドキュメントに未記載であり1日1回程度となります。

ハンズオン環境を保持する場合、無事削除されたことを是非確認をしてみてください。

## **デプロイにおける承認プロセスの考慮**

デプロイの際に、承認プロセスを導入することでデプロイの安全性を高めることができます。
Cloud Deployの機能を利用することで、デプロイの際に承認プロセスを導入することができます。

今回は、フロントエンドアプリケーションのデプロイに承認プロセスを導入してみましょう。

### **Cloud Deployの定義ファイルの編集**

Cloud Deployでは、既存のターゲット定義をコンソールから更新できません。
5章で利用した定義ファイルに対して、承認プロセスを追加してgcloudコマンドで更新します。

```bash
APP_TYPE=frontend
sed -e "s/PROJECT_ID/${GOOGLE_CLOUD_PROJECT}/g" doc/clouddeploy.yml | sed -e "s/REGION/asia-northeast1/g" | sed -e "s/requireApproval: false/requireApproval: true/g" | sed -e "s/SERVICE_NAME/cnsrun-${APP_TYPE}/g" > /tmp/clouddeploy_${APP_TYPE}.yml
```

再度、gcloudコマンドでターゲット定義を更新します。

```bash
gcloud deploy apply --file=/tmp/clouddeploy_${APP_TYPE}.yml --region asia-northeast1
```

<walkthrough-path-nav path="https://console.cloud.google.com/deploy" >Cloud Deploy に移動</walkthrough-path-nav>

1. <walkthrough-spotlight-pointer cssSelector="[id=cfctest-section-nav-item-delivery_pipelines]" validationPath="/deploy"> デリバリーパイプライン </walkthrough-spotlight-pointer>を選択します。
2. `[cnsrun-frontend]`を選択します。 
3. <walkthrough-spotlight-pointer locator="semantic({tab 'ターゲット'})" validationPath="/deploy/delivery-pipelines/asia-northeast1/cnsrun-frontend">ターゲット</walkthrough-spotlight-pointer>タブを選択します。
4. `[名前]`の列にあるリンクを選択し、ターゲットの詳細を確認します。

`[承認が必要です]`の箇所が`はい`になっていればOKです。

では実際にどのように承認プロセスが挟まるか確認してみましょう。

1. フロントエンドアプリケーションに対して変更を加えてコードプッシュをするか、Cloud Buildのトリガを手動実行してみましょう。
2. しばらくするとCode Deployが起動します。
<walkthrough-path-nav path="https://console.cloud.google.com/deploy" >Cloud Deploy に移動</walkthrough-path-nav>
<walkthrough-spotlight-pointer cssSelector="[id=cfctest-section-nav-item-delivery_pipelines]" validationPath="/deploy/delivery-pipelines/asia-northeast1/cnsrun-frontend"> デリバリーパイプライン </walkthrough-spotlight-pointer> → `[cnsrun-frontend]`を選択し、
デプロイが保留されていることが確認できます。
3. 保留中の絵の下にある<walkthrough-spotlight-pointer locator="semantic({button '確認'})" validationPath="/deploy/delivery-pipelines/.*">確認</walkthrough-spotlight-pointer>ボタンを押します。
4. 行の一番右にある<walkthrough-spotlight-pointer locator="semantic({link 'Review'})" validationPath="/deploy/delivery-pipelines/.*">REVIEW</walkthrough-spotlight-pointer>ボタンを押します。
5. 承認画面から、どういった変更が発生したかを確認できます。内容を確認したら、<walkthrough-spotlight-pointer locator="semantic({button '承認'})" validationPath="/deploy/delivery-pipelines/asia-northeast1/cnsrun-frontend/releases/.*">承認</walkthrough-spotlight-pointer>ボタンを押してください。

以上で、承認プロセスを導入したアプリケーションのデプロイが完了しました。

## **SLO指標の設定**

今回は5章で作成したフロントエンドアプリケーションに対して、SLOを設定します。
では早速、Cloud Runのコンソール画面からSLOを設定していきましょう。

<walkthrough-path-nav path="https://console.cloud.google.com/run" >Cloud Run に移動</walkthrough-path-nav>

1. フロントエンドアプリケーションを選択します。
2. <walkthrough-spotlight-pointer cssSelector="[cfcrouterlink=slos]">SLO</walkthrough-spotlight-pointer> タブを選択します。
3. SLOがまだ作られていないことを確認し、<walkthrough-spotlight-pointer locator="semantic({button '+ SLO を作成'})">SLOを作成</walkthrough-spotlight-pointer>ボタンを押します。

### **SLIの設定**

まずSLIを決定します。`[指標の選択]`では`[可用性]`を選択し、`[リクエストベース]`のSLIとして<walkthrough-spotlight-pointer locator="semantic({button '続行'})" validationPath="/run/detail/asia-northeast1/cnsrun.*">続行</walkthrough-spotlight-pointer>を押します。

SLI の詳細では、そのまま<walkthrough-spotlight-pointer locator="semantic({button '続行'})" validationPath="/run/detail/asia-northeast1/cnsrun.*">続行</walkthrough-spotlight-pointer>を押します。

### **サービスレベル目標（SLO）の作成** 

次はSLOの設定です。次のように設定をします。

- コンプライアンス期間
  - カレンダー
  - 1暦日
- パフォーマンス目標
  - 99%

<walkthrough-spotlight-pointer locator="semantic({button '続行'})" validationPath="/run/detail/asia-northeast1/cnsrun.*">続行</walkthrough-spotlight-pointer>を押します。

最後の確認画面でSLOの設定内容を確認し、<walkthrough-spotlight-pointer cssselector="button[type='submit']" validationPath="/run/detail/asia-northeast1/cnsrun.*">SLOを作成</walkthrough-spotlight-pointer>を押下して完了です。

通常はこのあと、アラートの設定をして、SLOの監視を行いますが、今回はSLOの設定のみで終了となります。

<walkthrough-info-message>
フロントエンドアプリケーションにランダムでエラーを応答するパスがあります。abコマンドがローカル環境にある場合、次のコマンドでCloud Runの可用性を下げて、SLOの表示がどう変化するか見てみるのもよいですね。

```bash
ab -n 50000 -c 2 https://$LB_GLOBAL_IP/random
```
</walkthrough-info-message>

## **未承認のコンテナイメージのデプロイを防ぐ**

コンテナイメージのセキュリティを強化するために、Binary Authorizationを利用して未承認のコンテナイメージのデプロイを防ぎます。

## **1. KMSを利用した鍵ペアの作成**

<walkthrough-enable-apis apis="cloudkms.googleapis.com"></walkthrough-enable-apis>

まずは、Binary Authorizationで利用するための鍵ペアを作成します。

KMSのキーリングを作成します。
    
```bash
KEYRING_NAME=cnsrun-keyring
gcloud kms keyrings create $KEYRING_NAME --location=asia-northeast1
```

次にキーリングに紐づける形で鍵ペアを作成します。

```bash
KEY_NAME=cnsrun-attestor-key
gcloud kms keys create $KEY_NAME --location=asia-northeast1 --keyring=$KEYRING_NAME --purpose=asymmetric-signing --default-algorithm=ec-sign-p256-sha256
```

## **2. 認証者(Attestor)の作成**

次に、Binary Authorizationで利用する認証者(Attestor)を作成します。

<walkthrough-path-nav path="https://console.cloud.google.com/security/binary-authorization/attestors" > Binary Authorization ページに移動</walkthrough-path-nav>

1. <walkthrough-spotlight-pointer cssSelector="[cfcrouterlink=attestors]" validationPath="/security/binary-authorization/.*">認証者</walkthrough-spotlight-pointer>タブを選択します。
2. <walkthrough-spotlight-pointer locator="semantic({link '認証者を作成'})" validationPath="/security/binary-authorization/attestors">認証者を作成</walkthrough-spotlight-pointer>ボタンを押します。
3. 認証者の名前として、`cnsrun-attestor`と入力します。
4. 公開鍵として<walkthrough-spotlight-pointer locator="semantic({button 'pkix 公開鍵を追加'})" validationPath="/security/binary-authorization/attestors">PKIX 公開鍵</walkthrough-spotlight-pointer>を選択します。
5. 次のコマンドで表示されるキーバージョンのリソースIDをメモします。次の手順で利用します。
```bash
KEY_VERSION=$(gcloud kms keys versions list --location=asia-northeast1 --keyring=$KEYRING_NAME --key=$KEY_NAME --format=json | jq -r .[].name)
echo $KEY_VERSION
```
6. <walkthrough-spotlight-pointer locator="semantic({button 'Cloud KMS から公開鍵マテリアルをインポート'})" validationPath="/security/binary-authorization/attestors">CLOUD KMSからインポート</walkthrough-spotlight-pointer>を選択します。
7. メモしておいたリソースIDを入力します。
7. 詳細設定の<walkthrough-spotlight-pointer locator="semantic({button '詳細設定を切り替えます'})" validationPath="/security/binary-authorization/attestors/create">アコーディオン</walkthrough-spotlight-pointer>を開いて、<walkthrough-spotlight-pointer cssSelector="[type=checkbox]" validationPath="/security/binary-authorization/attestors/create">Container Analysis メモを自動生成する</walkthrough-spotlight-pointer>にチェックを付けます。
8. <walkthrough-spotlight-pointer locator="semantic({button '作成'})" validationPath="/security/binary-authorization/attestors/create">作成</walkthrough-spotlight-pointer>ボタンを押します。

## **3. 証明書の作成**

次に、Binary Authorizationで利用するための証明書を作成します。

<walkthrough-info-message>通常はCI/CDプロセスの中で証明書を作成しますが、シンプルに試すために手動でも作成します。</walkthrough-info-message>

まず、対象とするArtifact RegistryのURIを取得します。
次のコマンドを実行して、最新のイメージダイジェスト（sha:256~~となっている文字列）を取得します。

```bash
IMAGE_PATH="asia-northeast1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/cnsrun-app/frontend"
gcloud artifacts docker images list $IMAGE_PATH --sort-by="~UPDATE_TIME" --format='value(version)' --limit=1
```

次に、Artifact RegistryのURIとイメージダイジェストを組み合わせて、証明書を作成するための変数を設定します。
`{ダイジェストの値}`には先ほど出力された値を設定してください。

```
IMAGE_TO_ATTEST="${IMAGE_PATH}@{ダイジェストの値}"
```

各種変数が設定されていることを確認し、証明書を作成します。

```bash
echo KEYRING_NAME=$KEYRING_NAME
echo KEY_NAME=$KEY_NAME
echo KEY_VERSION=$KEY_VERSION
```

```bash
gcloud beta container binauthz attestations sign-and-create \
    --artifact-url="${IMAGE_TO_ATTEST}" \
    --attestor=cnsrun-attestor \
    --attestor-project="${GOOGLE_CLOUD_PROJECT}" \
    --keyversion-location=asia-northeast1 \
    --keyversion-keyring="${KEYRING_NAME}" \
    --keyversion-key="${KEY_NAME}" \
    --keyversion="${KEY_VERSION}"
```

## **4. Binary Authorizationのポリシー設定**

最後に、Binary Authorizationのポリシーを設定します。
今回作成した認証者によって承認されたコンテナイメージのみデプロイを許可するように設定します。

1. <walkthrough-spotlight-pointer cssSelector="[cfcrouterlink=policy]" validationPath="/security/binary-authorization/.*">ポリシー</walkthrough-spotlight-pointer>タブを選択します。
2. <walkthrough-spotlight-pointer locator="semantic({link 'ポリシーを編集'})" validationPath="/security/binary-authorization/policy">ポリシーを編集</walkthrough-spotlight-pointer>を編集を押します。
3. デフォルトのルールとして、<walkthrough-spotlight-pointer cssSelector="input[value=REQUIRE_ATTESTATION]" validationPath="/security/binary-authorization/policy/edit">証明書を要求</walkthrough-spotlight-pointer>にチェックを入れます。
4. 要求した証明書を認証する認証者（attestor)を指定するために、 <walkthrough-spotlight-pointer locator="semantic({button '認証者の追加'})" validationPath="/security/binary-authorization/policy/edit">認証者の追加</walkthrough-spotlight-pointer>を押します。
5. `プロジェクトと認証者の名前により追加`にチェックを入れたままにし、アテスターの`[認証者の名前]`として、`cnsrun-attestor`を入力して、`[認証者の追加]`ボタンを押します。
6. <walkthrough-spotlight-pointer locator="semantic({button 'ポリシーを保存'})" validationPath="/security/binary-authorization/policy/edit">ポリシーを保存</walkthrough-spotlight-pointer>を押して編集を完了します。

## **5. Cloud Runのセキュリティ設定変更**

最後に、Cloud Runのセキュリティ設定を変更して、Binary Authorizationによる証明書の検証を有効にします。

<walkthrough-path-nav path="https://console.cloud.google.com/run">Cloud Runページに移動</walkthrough-path-nav>

本来はCloud Runのコードから修正をしますが手早く結果を見たいので、一旦はコンソールから編集をします。

1. フロントエンドアプリケーション `cnsrun-frontend` を選択します。
2. <walkthrough-spotlight-pointer cssSelector="[cfcrouterlink=security]" validationPath="/run/detail/.*">セキュリティ</walkthrough-spotlight-pointer> タブを選択します。
3. <walkthrough-spotlight-pointer cssSelector="[formcontrolname=binauthzNamedPolicySelect]" validationPath="/run/detail/.*">ポリシーを選択</walkthrough-spotlight-pointer> から、`default`を選択します。
4. <walkthrough-spotlight-pointer locator="semantic({button 'サービスの Binary Authorization の変更を適用する'})" validationPath="/run/detail/asia-northeast1/cnsrun-frontend/security">適用</walkthrough-spotlight-pointer>を押します。

以上で、Binary Authorizationによる証明書の検証を有効にする設定が完了しました。

## **6. テスト**

では、実際に証明書が付与されていないイメージダイジェストを持つリビジョンをCloud Runコンソールから手動でデプロイしてみましょう。

1. フロントエンドアプリケーション `cnsrun-frontend` を選択します。
2. <walkthrough-spotlight-pointer cssSelector="a[cfciamcheck='run.services.update']" validationPath="/run/detail/asia-northeast1/cnsrun-frontend/.*">新しいリビジョンの編集とデプロイ</walkthrough-spotlight-pointer>を押します。
3. 証明書がないコンテナイメージに切り替えます。`[コンテナイメージのURL]`から <walkthrough-spotlight-pointer locator="semantic({button '選択'})" validationPath="/run/deploy/.*">選択</walkthrough-spotlight-pointer> を押して、古いハッシュのイメージを選択します。
4. <walkthrough-spotlight-pointer cssSelector="[formcontrolname=serveImmediately]" validationPath="/run/deploy/.*">このリビジョンをすぐに利用する</walkthrough-spotlight-pointer>にチェックを入れます。
5. <walkthrough-spotlight-pointer cssSelector="[type=submit]" validationPath="/run/deploy/.*">デプロイ</walkthrough-spot-pointer> を押します。

リビジョンのデプロイがはじまると、Binary Authorizationによって証明書がないためデプロイが拒否されることが確認できます。
`No attestations found that were valid and signed by a key trusted by the attestor`

このように証明書が内部で検証され、証明書がない場合はデプロイが拒否されることが確認できました。


最後に、証明書を付与したイメージではデプロイができることを確認しておきましょう。
次のコマンドで表示される値に合致するコンテナイメージで、先ほどと同様にデプロイを実施してください。

```bash
echo $IMAGE_TO_ATTEST | cut -d: -f2 | cut -c 1-10
```

問題なくデプロイができることを確認できましたか？
これにより、証明書を付与したイメージであればデプロイが許可されていることを確認しました。

### **オプション: CI/CDプロセスの中で証明書を作成する**

証明者が検証する証明書は、通常CI/CDプロセスの中で証明書を作成します。
Cloud Buildのステップに変更を加えて、新しいイメージがデプロイされる際に証明を作成するようにしましょう。

#### **Cloud Buildのサービスアカウントの権限追加**

証明書を作成するための権限をCloud Buildに与えます。

- Artifact Analyticsのメモへの権限
- KMSの鍵の公開鍵の取得権限
- KMSの鍵の署名権限
- Binary Authorizationの証明者を取得する権限

```bash
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --role roles/containeranalysis.notes.editor --condition=None
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --role roles/containeranalysis.notes.occurrences.viewer --condition=None
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --role roles/containeranalysis.occurrences.editor --condition=None
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --role roles/cloudkms.publicKeyViewer --condition=None
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --role roles/cloudkms.signer --condition=None
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member serviceAccount:cnsrun-cloudbuild@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com --role roles/binaryauthorization.attestorsViewer --condition=None
```

#### **Cloud Buildのステップの追加**

証明書作成ステップを追加します。
フロントエンドアプリケーションのCloud BuildのYAML設定ファイル（`app/frontend/cloudbuild_push.yaml`）を編集し、証明書作成ステップを追加します。
具体的には、`「Chap6ハンズオンのみ使用。Binary Authorizationの証明書を作成するステップ」`と記載されたステップのコメントアウトを解除します。

#### **Cloud Runの設定ファイルにBinary Authorizationの設定を追加**

さきほどはコンソールから設定を変更していました。
このままCI/CDを回すと、Binary Authorizationの設定が無効化されるため、設定ファイルにBinary Authorizationの設定を追加します。

`frontend/cloudrun.yaml`を開き、設定を追加しましょう。

```patch
- #    run.googleapis.com/binary-authorization: default
+     run.googleapis.com/binary-authorization: default
```

ファイルを編集後、コードをプッシュしてCloud Buildを自動起動しましょう。

#### **テスト**

CI/CDプロセスが完了し、デプロイが成功することを確認しましょう。
祈りましょう。

Binary Authorizationが有効化されており、新しいアプリケーションのデプロイが成功していればOKです！

## **Cloud Runをカスタムドメインでホスティング**

Cloud Runはデフォルトで`*.run.app`のドメインで提供されますが、カスタムドメインを利用することで、自社のドメインで提供することができます。

外部ALBを使用する場合、自己署名証明書を利用することもできますが、カスタムドメインを利用することで信頼できるCAによる証明書を利用することができます。
本番利用では、APIアクセスの際にIP指定ではなくドメインでアクセスすることが多いです。

そのため、本ハンズオンでは独自ドメインを取得します。
そして、取得した独自ドメインを利用してCloud Runにアクセスする方法を試していきましょう。

<walkthrough-info-message><strong>注意：Cloud Domainでドメインを取得する場合、ドメインの取得費用がかかります。すでに独自ドメインを取得している場合や無料ドメインを利用する場合はCloud Domainの手順はスキップをしてください。</strong></walkthrough-info-message>

手順としては次の流れですすめます。

- Cloud Domainでドメインを取得
- ドメインに対する証明書を取得
- ロードバランサに対して証明書を設定

また、本ハンズオンでは、取得するドメイン名を`uma-arai.com`として記載をしています。
本箇所は各自の取得したいドメイン名に置き換えてください。

### **Cloud Domainでドメインを取得**

<walkthrough-info-message>再掲：Cloud Domainでドメインを取得する場合、ドメインの取得費用がかかります。すでに独自ドメインを取得している場合や無料ドメインを利用する場合はCloud Domainの手順はスキップをしてください。</walkthrough-info-message>

<walkthrough-enable-apis apis="domains.googleapis.com, dns.googleapis.com"></walkthrough-enable-apis>

1. コンソールから Cloud Domainsのページに移動します。
<walkthrough-watcher-block link-url="https://console.cloud.google.com/net-services/domains"> Cloud Domains に移動</walkthrough-watcher-block>

2. `[ドメインを登録]` をクリックします。

3. 購入する利用可能なドメイン名を検索し、`[選択]` をクリックしてカートに追加します。利用可能なドメインごとに料金が表示されています。

**注意：** こちらはドメイン取得時点で課金されるため、ご注意ください。

### **ドメインに対するDNSを設定**

`[DNS構成]` セクションで、`Cloud DNS を使用する（推奨）` がデフォルトで選択されています。
ゾーン名の変更は不要であるため、`[続行]` をクリックしましょう。

### **ドメインのプライバシー設定**

デフォルトでは、`プライバシー保護を有効にする`が選択されています。
2024年5月時点ではこちらのオプションを選択できますが、`「プライバシー保護を有効にする」のサポートは 2024 年初頭に終了します。このオプションを使用するすべての登録は、「一般公開される情報を制限する」に更新されます。`との注意書きがあります。
こちらは極力プライバシー保護をつけた状態にしておきましょう。

### **連絡先情報の入力**

指定するドメインの連絡先情報を入力します。
デフォルトでは、入力した連絡先情報が、登録者、管理者、技術担当者の連絡先に適用されます。
各項目を入力していきましょう。**この際、メールアドレスは自身が受信をできるアドレスを指定してください。**

ドメインを登録するには、`[登録]` をクリックします。

登録の処理には数分かかることがあります。
ドメインを登録したら、Cloud Domains から受信した確認メールに返信する必要があります。

### メールアドレスの認証

1. 新しいブラウザを開いて、ドメインの登録に使用したメールアカウントにログインします。 
2. Google Domains(domains-noreply@google.com）から届いた「Action required: Please verify your email address for `取得したドメイン名`」　という件名のメールを開きます。
3. `[Verify email now]` をクリックします。 
4. 遷移先の画面で、メールアドレスの確認が完了したことを示す確認メッセージが表示されます。

以上で、ドメインの取得が完了です。

--- 

こちらのドメインを利用して次のステップに進んでいきましょう。

## **ドメインに対する証明書を取得**

```bash
gcloud compute ssl-certificates create cnsrun-frontend     --description=DESCRIPTION     --domains=cnsrunapp.uma-arai.com     --global
```

### **ロードバランサに対して証明書を設定**

```bash
gcloud compute target-https-proxies update cnsrun-https-proxies \
    --ssl-certificates cnsrun-frontend \
    --global-ssl-certificates \
    --global
```

ロードバランサに証明書が正しく紐づいているか確認しましょう。

```bash
gcloud compute target-https-proxies describe cnsrun-https-proxies --global --format="get(sslCertificates)"
```

末尾に、`cnsrun-frontend`と表示されていればOKです。

## **DNSレコードの設定**

次に、いままでロードバランサが使用していたグローバルIPアドレスに対して、Aレコードを設定します。

```bash
LB_GLOBAL_IP=$(gcloud compute addresses describe cnsrun-ip --global --format='value(address)')
```

取得したIPアドレスをDNSレコードに設定します。
今回はCloud DNSをDNSサーバとして利用した場合の設定例を示します。
ほかのDNSサービスを利用している場合、そのサービスの設定方法に従って設定してください。

次のコマンドで表示される値に合致するゾーン名をメモしてください。

```bash
gcloud dns managed-zones list --format=json | jq -r .[].name
```

先ほどメモしたゾーン名を変数に入れておきます。

```bash
CNSRUN_ZONE={メモしたゾーン名}
```

```bash
gcloud dns record-sets transaction start --zone=${CNSRUN_ZONE}
gcloud dns record-sets transaction add $LB_GLOBAL_IP --name=cnsrunapp.uma-arai.com. --ttl=300 --type=A --zone=${CNSRUN_ZONE}
gcloud dns record-sets transaction execute --zone=${CNSRUN_ZONE}
```

ここまでの設定をすることで、GoogleマネージドSSL証明書が有効になるための準備ができました。
設定したDNSレコードが伝搬されて証明書が有効化されるまでは時間がかかる場合があります。
気長に待ちましょう（筆者の環境では5分で終わるときもあれば、30分かかる時もありました）。

<walkthrough-footnote>DNSレコードの伝搬に時間がかかる：https://cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-certs?hl=ja#dns_record_propagation_time</walkthrough-footnote>

十分に時間をおいた後、証明書が正しく設定されているか確認しましょう。
ステータスが`ACTIVE`であれば正常に設定されています。

```bash
gcloud compute ssl-certificates describe cnsrun-frontend --global --format="get(name,managed.status, managed.domainStatus)"
```

## **カスタムドメインのテスト**

最後はフロントエンドアプリケーションにアクセスして、ドメインでアクセスができるか確認しましょう。
ただし、こちらについてもロードバランサがSSL証明書を利用するまでに時間がかかることがあります。

<walkthrough-footnote>https://cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-certs?hl=ja#step_5_test_with_openssl</walkthrough-footnote>

次のコマンドでロードバランサがクライアントに対して提示する証明書が想定通りであることを確認します。

```bash
echo | openssl s_client -showcerts -servername cnsrunapp.uma-arai.com -connect ${LB_GLOBAL_IP}:443 -verify 99 -verify_return_error
```

出力の最終行に、`Verify return code: 0 (ok)`と表示されていれば正常に証明書が設定されています。
続けてフロントエンドアプリケーションにアクセスしてみましょう。

```bash
curl -i https://cnsrunapp.uma-arai.com/backend?id=none
# HTTP200 OK
```

HTTP200 OKが返却されれば正常に設定されています。

このようにロードバランサやDNS周辺の設定が少しありましたがシンプルにできたかと思います。
Google Cloudのマネージドサービスを利用することで、簡単に証明書を取得してカスタムドメインを利用することができます。

本設定を利用することで、さらにプロダクション環境に近い状態でCloud Runを利用したサービスの提供が可能となります。

## **お疲れ様でした！**

<walkthrough-conclusion-trophy></walkthrough-conclusion-trophy>

#### リソースの削除

最後に課金を防ぐため、リソースの削除に進みましょう。
6章で作成したリソースを先に削除し、次に5章で作成をしたリソースを削除します。
各章の削除手順を記載したマークダウンを順次確認してください。
    
```bash
teachme doc/deletion/chap6_deletion.md
```
