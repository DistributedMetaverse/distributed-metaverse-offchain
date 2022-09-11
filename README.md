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
{"qmhash": "<qmhash>", "mimetype": "<plain/text>"}
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
    "id": 110
}
```

## 블록 예시

```json
{
    "previousHash": "00000036aa9d9b0549584728c15fa4b3972e689377b39b098e192a753f6e7f8f",
    "transactions": [{
        "data": "QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 107,
        "datetime": "2022-09-11 22:12:28.4575376 +0900 KST m=+5.182499101"
    }, {
        "data": "QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 108,
        "datetime": "2022-09-11 22:12:30.4761618 +0900 KST m=+7.201123301"
    }, {
        "data": "QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 109,
        "datetime": "2022-09-11 22:12:31.4964231 +0900 KST m=+8.221384601"
    }, {
        "data": "QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA,video/x-msvideo",
        "id": 110,
        "datetime": "2022-09-11 22:12:32.8842194 +0900 KST m=+9.609180901"
    }],
    "hash": "000000e063ea9a97469ca4351bafad1d626d5f3051f7bb9e56f59ac0aa56d17a",
    "datetime": "2022-09-11 22:13:01.2140948 +0900 KST m=+37.939056301",
    "proof": 1275593,
    "lastTransactionId": 110
}
```

## 문의
문의사항이 있으면 알려주세요

* Go Namhyeon <gnh1201@gmail.com>
