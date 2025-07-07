# **THE Cloud Run ハンズオン**

今回のハンズオンは、読者の方ができる限り迷わずにハンズオンをするための工夫をたくさんしました。
ハンズオンのファイルはそれぞれ独立したマークダウンで作成しています。

一部のマークダウン拡張の記述がうまく動作しないケースがあるため、できるだけ本ファイルから対象のハンズオンを開始していただけますと幸いです。

## 準備：自身のGitHubアカウントへソースをコピー

本ハンズオンでは、筆者が用意をしたソースコードを更新して、アプリケーションを更新をする箇所があります。
みなさまが自由にソースコードを更新できるようにするため、みなさま自身のGitHubへソースコードをコピー（Fork）していただきます。

1. GitHubのアカウントへログインをします。
2. [ハンズオンのコードURL](https://github.com/uma-arai/cloudrun-handson)を開きます。
3. 画面内にある`[Fork]`ボタンをクリックします。
4. `[Create fork]`ボタンをクリックします。

以上で、みなさまのGitHubアカウントへソースコードがコピーされました。
自身のGitHubアカウントにコピーしたハンズオンコンテンツを取得しておきましょう。

```bash
GITHUB_USERNAME={自身のGitHubユーザ名}
```

```bash
git clone https://github.com/${GITHUB_USERNAME}/cloudrun-handson
```

チュートリアル（teachme)は、Forkしたコンテンツでなくても問題ありませんので、このまま続けていきます。
それでは、ハンズオンを進めていきましょう。

## **ハンズオンの一覧**

### **5章ハンズオン**

```bash
cd ~/cloudrun-handson
teachme doc/handson_chap5.md
```

### **6章ハンズオン**

```bash
cd ~/cloudrun-handson
teachme doc/handson_chap6.md
```
