# いつもヘルプとは？
簡易作業管理サービスです

# 開発環境の構築
開発環境はdocker-composeで構築しています
事前に3つの環境変数を設定する必要があります

開発環境用文字列は別途問い合わせてください

```
$ export ITODOENV=local
$ export GOOGLE_OAUTH2_CLIENT_ID=開発環境用の文字列
$ export GOOGLE_OAUTH2_CLIENT_SECRET=開発環境用の秘密文
$ git clone https://github.com/itsumohelp/itsumo.git
$ cd itsumo
$ docker-compose up
```

起動後URL
http://localhost:5556/

# API reference
## get /todos
```
# response
{
 "A",
 "B",
 "C"
}
```

## get /todos/{todo_id}
```
# response
[
  {
    "value":"A",
    "done":0
  },
  {
    "value":"A",
    "done":0
  }
]
```

## post /elements/{todo_id}
```
# request
[
  {
    "value":"A",
    "done":0
  },
  {
    "value":"A",
    "done":0
  }
]
```

