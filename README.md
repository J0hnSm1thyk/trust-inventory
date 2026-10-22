# trust-inventory

端末にインストールされているパッケージマネージャー、言語ランタイム、TLSクライアントを読み取り専用で列挙するCLIです。

現在は、2026年10月に公表された[eMudhra中間CA問題](https://emudhra.we-have-a-problem.com/)について、検出したバージョンを既知の修正版と比較します。

## 対応環境

- Windows x64 / ARM64
- macOS Intel / Apple Silicon
- Linux x64 / ARM64

## 主な検出対象

pip、certifi、uv、Conda、Poetry、Node.js、npm、Yarn、pnpm、Go、Java、Rust、Cargo、Composer、RubyGems、Homebrew、WinGet、Chocolatey、Scoop、curl、Git、OpenSSLなど。

## Goのインストール

[Releases](https://github.com/J0hnSm1thyk/trust-inventory/releases)のビルド済みバイナリを使う場合、Goのインストールは不要です。ソースからビルドする場合のみ、Go 1.27.3以降をインストールしてください。

公式の[Goダウンロードページ](https://go.dev/dl/)から、使用中のOSとCPUに合うファイルを取得します。

- Windows: `.msi`を実行し、画面の案内に従います。完了後にターミナルを開き直します。
- macOS: `.pkg`を実行し、画面の案内に従います。完了後にターミナルを開き直します。
- Linux: 公式の[インストール手順](https://go.dev/doc/install)に従ってアーカイブを`/usr/local/go`へ展開し、`/usr/local/go/bin`を`PATH`へ追加します。既存のGoディレクトリへ上書き展開しないでください。

インストール後、次のコマンドで確認します。

```sh
go version
```

## ビルド

今回のGo証明書修正を含めるため、Go 1.27.3以降を使用してください。

```sh
go test ./...
go build -trimpath .
```

## 使い方

```sh
# 表形式
./trust-inventory

# JSON形式
./trust-inventory --format json

# JSONをファイルへ保存
./trust-inventory --format json --output report.json
```

Windowsでは実行ファイル名が`trust-inventory.exe`になります。

ビルド済みバイナリは[Releases](https://github.com/J0hnSm1thyk/trust-inventory/releases)から取得できます。バイナリは未署名のため、OSの警告が表示される場合があります。同梱の`SHA256SUMS.txt`で整合性を確認してください。

### 接続検査

次のオプションを指定した場合のみ、公開されているeMudhraテストホストへ接続します。

```sh
./trust-inventory --probe-emudhra
```

外部通信は既定で無効です。多数の端末から一斉に接続検査を実行しないでください。

## 判定

- `safe-by-version`: 登録済み修正版以上
- `update-required`: 登録済み修正版より古い
- `safe`: 接続検査で問題の証明書を拒否
- `vulnerable`: 接続検査で問題の証明書を受理
- `inventory-only`: 検出のみ。判定ルールなし
- `unknown`: 判定不能

## プライバシーと制限

- 結果を外部へ自動送信しません。
- ホスト名は`--include-hostname`を指定しない限り収集しません。
- ホームディレクトリは出力時に`~`へ置換します。
- コンテナ、WSL、仮想マシンは各環境内で実行してください。
- アプリ内部に埋め込まれたランタイムや独自CAを検出できない場合があります。
- 本ツールの結果は安全性を保証するものではありません。

## License

[MIT](LICENSE)
