# distributed-metaverse-offchain
공개SW대회용 오프체인(Offchain) 구현

## 특징
* 블록은 MQ(Redis 기반 메시지 큐) 및 로컬에 저장됨 (분산 저장)
* 작업증명(PoW)에 따라 블록을 생성

## Deamon test

1. 실행파일이 없는 경우 `go build` 명령으로 빌드 진행하여야 함
2. 윈도우즈는 `distributed-metaverse-offchain.exe`, 리눅스는 `distributed-metaverse-offchain` 실행
3. 자동으로 작업증명(PoW), 트랜젝션 수발신, 블록 수발신이 시작됨

## API 상세

## `/transaction/publish` (POST): 아래 JSON 양식으로 전송함.

### 요청
```json
{"qmhash": "<qmhash>", "mimetype": "<plain/text>", "filename": "<filename>"}
```

### 응답
```json
{
    "success": true,
    "id": 115
}
```

## `/transaction/:id` (GET): 트랜젝션 정보 조회 (파일 다운로드 요청까지 겸함)

### 요청
```
http://<host>:1323/transaction/115
```

### 응답 (파일 존재 시)
```json
{
    "success": true,
    "data": {
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 115,
        "datetime": "2022-09-12 00:54:37.9252438 +0900 KST m=+99.276254501"
    },
    "status": "ok",
    "url": "http://127.0.0.1:1323/downloads/QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA"
}
```

### 응답 (파일 존재하지 않을 시)
```
{
    "success": true,
    "data": {
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 115,
        "datetime": "2022-09-12 00:54:37.9252438 +0900 KST m=+99.276254501"
    },
    "status": "downloading"
}
```

## `/block/:hash` (GET): 블록 정보 조회

### 요청
```
http://<host>:1323/block/000000ae3e856ab81898f840f036b1a9c5d76e08bf3c482c57ff9f6deb303379
```

### 응답
```json
{"data":{"previousHash":"000000b7b5e037df613bf1173f420b5a9228cbaab3581f25005afe791aa1dfe7","transactions":[{"data":"{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}","id":111,"datetime":"2022-09-12 00:54:32.5607562 +0900 KST m=+93.911766901"},{"data":"{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}","id":112,"datetime":"2022-09-12 00:54:34.0608671 +0900 KST m=+95.411877801"},{"data":"{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}","id":113,"datetime":"2022-09-12 00:54:35.4791572 +0900 KST m=+96.830167901"},{"data":"{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}","id":114,"datetime":"2022-09-12 00:54:36.7781704 +0900 KST m=+98.129181101"},{"data":"{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}","id":115,"datetime":"2022-09-12 00:54:37.9252438 +0900 KST m=+99.276254501"}],"hash":"000000ae3e856ab81898f840f036b1a9c5d76e08bf3c482c57ff9f6deb303379","datetime":"2022-09-12 00:54:58.6619114 +0900 KST m=+120.012922101","proof":5546121,"lastTransactionId":115},"success":true}
```

## `/chain/:depth` (GET): 최근 블록 조회

### 요청
```
http://<host>:1323/chain/15
```

### 응답
```json
{"blocks": [<Block>, <Block>, <Block>, ...]}
```

## API test

아래와 같이 `test.json` 파일을 생성 후 cURL 명령을 이용하여 테스트 가능함.

### test.json

```json
{"qmhash": "<qmhash>", "mimetype": "<plain/text>", "filename": "<filename>"}
```

### cURL command

#### 요청
```bash
curl -X POST -d @test.json http://127.0.0.1:1323/transaction/publish
```

##### 응답
```json
{
    "success": true,
    "id": 115
}
```

## 블록 예시

```json
{
    "previousHash": "000000b7b5e037df613bf1173f420b5a9228cbaab3581f25005afe791aa1dfe7",
    "transactions": [{
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 111,
        "datetime": "2022-09-12 00:54:32.5607562 +0900 KST m=+93.911766901"
    }, {
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 112,
        "datetime": "2022-09-12 00:54:34.0608671 +0900 KST m=+95.411877801"
    }, {
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 113,
        "datetime": "2022-09-12 00:54:35.4791572 +0900 KST m=+96.830167901"
    }, {
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 114,
        "datetime": "2022-09-12 00:54:36.7781704 +0900 KST m=+98.129181101"
    }, {
        "data": "{\"qmhash\":\"QmVEL2JqkH1Gd58EuhfkdYJMLEdHLzqK6EJbFDz8RyNLvA\",\"mimetype\":\"video/x-msvideo\",\"filename\":\"big_buck_bunny_1080p_stereo.avi\"}",
        "id": 115,
        "datetime": "2022-09-12 00:54:37.9252438 +0900 KST m=+99.276254501"
    }],
    "hash": "000000ae3e856ab81898f840f036b1a9c5d76e08bf3c482c57ff9f6deb303379",
    "datetime": "2022-09-12 00:54:58.6619114 +0900 KST m=+120.012922101",
    "proof": 5546121,
    "lastTransactionId": 115
}
```

## 문의
문의사항이 있으면 알려주세요

* Go Namhyeon <gnh1201@gmail.com>
