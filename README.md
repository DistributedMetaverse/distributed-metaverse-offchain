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
    "previousHash": "0000007e2d969207638009a10f6e92c08b081a9ec9ad5a8bccc9471fc0e96fa3",
    "transactions": [{
        "data": "123123123,QM1234,image/jpg",
        "id": 92,
        "datetime": "2022-09-11 17:51:24.8870431 +0900 KST m=+176.558629101"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 93,
        "datetime": "2022-09-11 17:51:38.4658174 +0900 KST m=+190.137403401"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 94,
        "datetime": "2022-09-11 17:51:40.0039828 +0900 KST m=+191.675568801"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 95,
        "datetime": "2022-09-11 17:51:41.8941423 +0900 KST m=+193.565728301"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 96,
        "datetime": "2022-09-11 17:51:43.1744591 +0900 KST m=+194.846045101"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 97,
        "datetime": "2022-09-11 17:51:44.3177225 +0900 KST m=+195.989308501"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 98,
        "datetime": "2022-09-11 17:51:45.3428911 +0900 KST m=+197.014477101"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 99,
        "datetime": "2022-09-11 17:51:46.4758142 +0900 KST m=+198.147400201"
    }, {
        "data": "123123123,QM1234,image/jpg",
        "id": 100,
        "datetime": "2022-09-11 17:52:18.5046022 +0900 KST m=+230.176188201"
    }],
    "hash": "000000b5c013ea9b28cea69dc1e08a2bb554ffc26ac0660e31e932f707ba0b63",
    "datetime": "2022-09-11 17:52:41.3425134 +0900 KST m=+253.014099401",
    "proof": 13595465,
    "lastTransactionId": 100
}
```

## 문의
문의사항이 있으면 알려주세요

* Go Namhyeon <gnh1201@gmail.com>
