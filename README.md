# distributed-metaverse-offchain
공개SW대회용 오프체인(Offchain) 구현

## 특징
* 블록은 MQ(Redis 기반 메시지 큐) 및 로컬에 저장됨 (분산 저장)
* 작업증명(PoW)에 따라 블록을 생성

## Deamon test

1. 실행파일이 없는 경우 `go build` 명령으로 빌드 진행하여야 함
2. 윈도우즈는 `distributed-metaverse-offchain.exe`, 리눅스는 `distributed-metaverse-offchain` 실행
3. 자동으로 작업증명(PoW), 트랜젝션 수발신, 블록 수발신이 시작됨

## API test

아래와 같이 `test.json` 파일을 생성 후 CURL 명령을 이용하여 테스트 가능함.

### test.json

```json
{"checksum": "<checksum>", "qmhash": "<qmhash>", "mimetype": "<plain/text>"}
```

### cURL command

Request (URI: http://127.0.0.1:1323/transaction/publish)
```bash
curl -X POST -d @test.json http://127.0.0.1:1323/transaction/publish
```

Response
```json
{
    "success": true,
    "id": 32
}
```

## 블록 예시

```json
{
    "previousHash": "000000d7665350257713f53041a89f1e60dd0aa633899e65440ff1c0479b56fe",
    "transactions": [{
        "data": "ea20b1c2578f1bd88603a6ca217eb768,QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 101,
        "datetime": "2022-09-11 22:04:27.6124375 +0900 KST m=+1126.530436101"
    }, {
        "data": "ea20b1c2578f1bd88603a6ca217eb768,QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 102,
        "datetime": "2022-09-11 22:04:29.3946141 +0900 KST m=+1128.312612701"
    }, {
        "data": "ea20b1c2578f1bd88603a6ca217eb768,QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 103,
        "datetime": "2022-09-11 22:04:30.4866927 +0900 KST m=+1129.404691301"
    }, {
        "data": "ea20b1c2578f1bd88603a6ca217eb768,QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 104,
        "datetime": "2022-09-11 22:04:31.8887047 +0900 KST m=+1130.806703301"
    }, {
        "data": "ea20b1c2578f1bd88603a6ca217eb768,QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 105,
        "datetime": "2022-09-11 22:04:33.5135995 +0900 KST m=+1132.431598101"
    }],
    "hash": "0000006a1967b72e171ed11ab011669e7b46b7619fe109f03dffe847662ba10c",
    "datetime": "2022-09-11 22:05:07.6625966 +0900 KST m=+1166.580595201",
    "proof": 8104994,
    "lastTransactionId": 105
}
```

## 문의
문의사항이 있으면 알려주세요

* Go Namhyeon <gnh1201@gmail.com>
