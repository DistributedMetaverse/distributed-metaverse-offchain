// 2022-09-11
// https://github.com/DistributedMetaverse/distributed-metaverse-offchain
// 공개SW대회용 오프체인(Offchain) 구현

package main

import (
    "os"
    "context"
    "fmt"
    "time"
    "strconv"
    "net/http"
    "encoding/json"
    "crypto/sha256"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client
var storedTransactions []Transaction

type Block struct {
    PreviousHash string `json:"previousHash"`
    Transactions []Transaction `json:"transactions"`
    Hash string `json:"hash"`
    Datetime string `json:"datetime"`
    LastTransactionId int `json:"lastTransactionId"`
}

type Transaction struct {
    Data string `json:"data"`
    Id int `json:"id"`
    Datetime string `json:"datetime"`
}

// 해시 계산
func (b Block) CalculateHash() string {
    jsonBytes, _ := json.Marshal(b.Transactions)
    blockData := b.PreviousHash + string(jsonBytes) + b.Datetime
    blockHash := sha256.Sum256([]byte(blockData))
    return fmt.Sprintf("%x", blockHash)
}

// 외부 API
func serve() {
    // Echo instance
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Routes
    e.POST("/transaction/publish", publishTransaction)
    e.POST("/block/publish", publishBlock)

    // Start server
    e.Logger.Fatal(e.Start(":1323"))
}

func receiveTransactions() {
    fmt.Println("* receiveTransactions")

    pubsub := rdb.Subscribe(ctx, "transactions_live")
    defer pubsub.Close()

    // 메시지 계속 수신
    for {
        fmt.Println("*")

        // 최근 메시지 수신
        msg, err := pubsub.ReceiveMessage(ctx)
        if err != nil {
            panic(err)
        }

        fmt.Println(msg.Payload)

        transaction := Transaction{}   // 트랜젝션 객체 생성
        json.Unmarshal([]byte(msg.Payload), &transaction)   // JSON을 객체로 변환
        storedTransactions = append(storedTransactions, transaction)    // 트랜젝션 추가

        // 콘솔 메시지 출력
        fmt.Println(msg.Channel, msg.Payload)
    }
}

// 블록 빌드
func buildBlock() (Block, error) {
    var transactions []Transaction   // 트랜젝션 변수 정의

    // 마지막 트랜젝션 ID 구하기
    lastTransactionId, _ := getLastTransactionId()
    
    // 마지막 블록 Hash 구하기
    lastBlockHash, _ := getLastBlockHash()

    // 현재 트랜젝션 복사
    //_storedTransactions := make([]Transaction, len(storedTransactions))
    //copy(storedTransactions, _storedTransactions)
    //storedTransactions = make([]Transaction, 0)

    // 트랜젝션 확인
    for _, transaction := range storedTransactions {
        if transaction.Id <= lastTransactionId {
            transactions = append(transactions, transaction)
        }
    }
    storedTransactions = nil   // 메모리에 남은 모든 트랜젝션 삭제

    // 블록 생성
    block := Block{
        PreviousHash: lastBlockHash,
        Transactions: transactions,
        Datetime: time.Now().String(),
        LastTransactionId: lastTransactionId,
    }

    // 블록 해시 계산
    block.Hash = block.CalculateHash()

    // 블록 반환
    return block, nil
}

// 트랜젝션 빌드
func buildTransaction(data string) (Transaction, error) {
    lastTransactionId, err := getLastTransactionId()

    if err != nil {
        return Transaction{}, err
    }

    transaction := Transaction{
        Data: data,
        Id: lastTransactionId + 1,
        Datetime: time.Now().String(),
    }

    return transaction, nil
}

// 최근 트랜젝션 ID 구하기
func getLastTransactionId() (int, error) {
    lastTransactionId := -1

    result, _ := rdb.Get(ctx, "lastTransactionId").Result()
    transactionId, err := strconv.Atoi(result)

    if err != nil {
        err := rdb.Set(ctx, "lastTransactionId", "0", 0).Err()
        if err != nil {
            lastTransactionId = 0
        }
    } else {
        lastTransactionId = transactionId
    }

    return lastTransactionId, err
}

// 최근 블록 해시 구하기
func getLastBlockHash() (string, error) {
    lastBlockHash := ""

    result, err := rdb.Get(ctx, "lastBlockHash").Result()
    if err != nil {
        //block, err := createBlock()
        //return block.hash, err
        return lastBlockHash, err
    } else {
        lastBlockHash = result
    }

    return lastBlockHash, nil
}

// 시작점(Entrypoint)
func main() {
    // Redis 서버에 연결
    rdb = redis.NewClient(&redis.Options{
        Addr:     "154.12.242.48:60713",
        //Password: "", // no password set
        //DB:       0,  // use default DB
    })
    
    // 최근 트랜젝션 ID 확인
    lastTransactionId, _ := getLastTransactionId()
    fmt.Printf("lastTransactionId: %d\n", lastTransactionId)
    
    // 최근 블록 해시 확인
    lastBlockHash, _ := getLastBlockHash()
    fmt.Printf("lastBlockHash: %s\n", lastBlockHash)

    // 수신 함수 실행
    go receiveTransactions()

    // 웹서버 실행
    serve()
}

// 블록 생성
func createBlock() (Block, error) {
    block, err := buildBlock()

    // 발행할 블록 본문 만들기
    jsonBytes, err := json.Marshal(block)
    if err != nil {
        return block, err
    }

    // 블록 생성
    err2 := rdb.Set(ctx, block.Hash, string(jsonBytes), 0).Err()
    if err2 != nil {
        return block, err2
    }

    // 최근 해시 갱신
    err3 := rdb.Set(ctx, "lastBlockHash", block.Hash, 0).Err()
    if err3 != nil {
        return block, err3
    }

    // 발생된 블록을 로컬에 파일로 저장
    saveBlock(block, jsonBytes)

    return block, err
}

// 디렉토리 존재 여부
func isExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

// 블록을 로컬에 저장
func saveBlock(b Block, jsonBytes []byte) error {
    path := "data"   // 파일 디렉토리 경로 지정

    // 폴더 확인 및 생성
    result, err := isExists(path)
    if result == false || err != nil {
        err := os.Mkdir(path, os.ModePerm)
        if err != nil {
            return err
        }
    }

    // 블록을 저장할 파일 생성
    f, err := os.Create("data/" + b.Hash)
    defer f.Close()    // 작업을 완료하면 파일을 닫음
    if err != nil {
        return err
    }
    
    // 파일에 내용 작성
    n, err := f.Write(jsonBytes)
    if err != nil {
        return err
    }
    fmt.Printf("wrote %d bytes\n", n)

    return nil
}

// 트랜젝션 발행
func publishTransaction(c echo.Context) error {
    // 받은 요청 해석
    jsonMap := make(map[string]interface{})
    err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
    if err != nil {
        return err
    }

    // 변수로 재정렬
    filehash := jsonMap["filehash"].(string)
    filetype := jsonMap["filetype"].(string)
    filename := jsonMap["filename"].(string)

    // 신규 트랜젝션 생성
    transaction, err := buildTransaction(filehash + "," + filetype + "," + filename)
    if err != nil {
        return err
    }

    // 발행할 트랜젝션 본문 만들기
    jsonBytes, err := json.Marshal(transaction)
    if err != nil {
        return err
    }

    // 마지막 트랜젝션 ID 올리기
    err2 := rdb.Set(ctx, "lastTransactionId", strconv.Itoa(transaction.Id), 0).Err()
    if err2 != nil {
        return err2
    }

    // 트랜젝션 발행
    err3 := rdb.Publish(ctx, "transactions_live", string(jsonBytes)).Err()
    if err3 != nil {
        return err3
    }

    // 모든 작업이 완료되었으면 오류 없음으로 반환
    response := map[string]interface{}{
        "success": true,
        "id": transaction.Id,
    }
    return c.JSON(http.StatusOK, response)
}

func publishBlock(c echo.Context) error {
    // 블록 생성
    block, err := createBlock()
    if err != nil {
        return err
    }

    // 모든 작업이 완료되었으면 오류 없음으로 반환
    response := map[string]interface{}{
        "success": true,
        "hash": block.Hash,
    }
    return c.JSON(http.StatusOK, response)
}

// References:
//     https://gist.github.com/LordGhostX/bb92b907731ee8ebe465a28c5c431cb4
//     https://redis.uptrace.dev/guide/go-redis-pubsub.html
//     https://stackoverflow.com/questions/41410655/extract-json-from-golangs-echo-request
//     https://stackoverflow.com/questions/27137521/how-to-convert-interface-to-string
//     https://stackoverflow.com/questions/26327391/json-marshalstruct-returns
//     https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists
