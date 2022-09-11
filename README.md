# distributed-metaverse-offchain
공개SW대회용 오프체인(Offchain) 구현

## 특징
* 블록은 MQ(Redis 기반 메시지 큐) 및 로컬에 저장됨 (분산 저장)
* 작업증명(PoW)에 따라 블록을 생성

## Deamon test

1. `go build` 명령 후 윈도우즈는 `distributed-metaverse-offchain.exe`, 리눅스는 `distributed-metaverse-offchain` 실행
2. 자동으로 작업증명(PoW), 트랜젝션 수발신, 블록 수발신이 시작됨

## API test

아래와 같이 `test.json` 파일을 생성 후 CURL 명령을 이용하여 테스트 가능함.

### test.json

```
{"checksum": "<checksum>", "qmhash": "<qmhash>", "mimetype": "<plain/text>"}
```

### cURL command

Request (URI: http://127.0.0.1:1323/transaction/publish)
```
curl -X POST -d @test.json http://127.0.0.1:1323/transaction/publish
```

Response
```
{
    "success": true,
    "id": 32
}
```

## 문의
문의사항이 있으면 알려주세요

* Go Namhyeon <gnh1201@gmail.com>
