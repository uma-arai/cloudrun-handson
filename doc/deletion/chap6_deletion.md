# **The Cloud Run ハンズオン <br />　6章削除手順**

この章では、6章で作成したリソースを削除する手順を説明します。
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

## **Cloud Armorの削除**

```bash
BACKEND_SERVICE_NAME=$(gcloud compute backend-services list --format=json | jq -r .[].name | grep cnsrun)
gcloud compute backend-services update $BACKEND_SERVICE_NAME --global --security-policy=
gcloud compute security-policies delete cnsrun-waf-policy --quiet
```

## **GoogleマネージドSSL証明書の削除**

まずはロードバランサのHTTPSターゲットプロキシに設定された証明書を取り外します。
5章で作成した証明書のみ紐づける設定に戻しましょう。

```bash
CNSRUN_CERT=$(gcloud compute ssl-certificates list --format=json | jq -r '.[] | select(.type == "SELF_MANAGED") | .name')
gcloud compute target-https-proxies update cnsrun-https-proxies --ssl-certificates=${CNSRUN_CERT}
```

続いて、GoogleマネージドSSL証明書の削除をします。

```bash
gcloud compute ssl-certificates delete cnsrun-frontend --quiet
```

## **DNSレコードの削除**

本ハンズオン用にCloud Domainsでドメインを取得している場合のみ、本手順を実施します。

<walkthrough-info-message>自身で管理していたドメインを利用している場合は、この手順は不要です。ハンズオンで追加したDNSレコードは削除しておきましょう</walkthrough-info-message>

ロードバランサのグローバルIPに紐づけたAレコードの削除をしましょう。

<walkthrough-path-nav path="https://console.cloud.google.com/net-services/dns" >Cloud DNS に移動</walkthrough-path-nav>

1. 作成したDNSゾーンを選択します。
2. SOAレコード、NSレコードを除くすべてのレコードを削除します。
3. `[ゾーンを削除]`から対象のゾーンを削除します。

## **ドメインの削除**
本ハンズオン用にCloud Domainsでドメインを取得している場合のみ、本手順を実施します。

<walkthrough-info-message>自身で管理していたドメインを利用している場合は、この手順は不要です</walkthrough-info-message>

<walkthrough-watcher-block link-url="https://console.cloud.google.com/net-services/domains"> Cloud Domains に移動</walkthrough-watcher-block>

1. `[登録]`画面で削除するドメイン名を選択します。
2. 削除するドメイン名の横にある`[その他の操作]`ボタンを押します。
3. ドメインを削除するには、`[削除]`を押します。

## **SSLポリシーの削除**

```bash
gcloud compute target-https-proxies update cnsrun-https-proxies --clear-ssl-policy
gcloud compute ssl-policies delete cnsrun-ssl-policy --global --quiet
```

## **Artifact Registryの脆弱性スキャン**

Artifact Registryのコンソール画面から`スキャン：オフ`に戻しましょう。

<walkthrough-watcher-block link-url="https://console.cloud.google.com/artifacts"> Artifact Registry に移動</walkthrough-watcher-block>

## **Binary Authorizationの削除**

### **1. Binary Authorizationのポリシーを戻す**

<walkthrough-path-nav path="https://console.cloud.google.com/security/binary-authorization/attestors" > Binary Authorization ページに移動</walkthrough-path-nav>

1. `[ポリシーを編集]`を押します。
2. `[すべてのイメージを許可]`に設定を戻し、保存します。

### **2. 認証者(Attesotr)の削除**

```bash
gcloud beta container binauthz attestors delete cnsrun-attestor
```

### **3. KMSの削除について**

Cloud KMSについては、削除はすぐには実施できません。
詳細は、[Cloud KMSのドキュメント](https://cloud.google.com/kms/docs/destroy-restore?hl=ja#timeline)を参照してください。

Cloud KMSのコンソール画面に進み、作成したキーリングを選択し、キーの無効化だけをしておきましょう。

## **その他**

下記については、もともとあったリソースへの設定です。無理に削除する必要はないため、そのままにしておきます。
5章のリソース削除でまとめて削除をします。

- クリーンアップポリシー
- 承認プロセス
- SLOの設定

以上で、6章で作成したリソースの削除が完了しました。
お疲れ様でした。

## **5章で作成したリソースの削除**

続けて5章で作成したリソースも削除していきましょう。
なお、**「プロジェクトごと削除」を選択した場合は、この手順は不要です。**

```bash
teachme doc/deletion/chap5_deletion.md
```

