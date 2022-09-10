package main

import (
    "context"
    "fmt"
    "time"
    "strconv"
    "encoding/json"
    "crypto/sha256"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/go-redis/redis/v8"   
)

var ctx = context.Background()
var rdb *redis.Client

type Blockchain struct {}

type Block struct {
    previousHash string
    transactions []Transaction
    hash string
    timestamp time.Time
    proof int
    lastTransactionId int
}

type Transaction struct {
    data string
    id int
    timestamp time.Time
}

func (b Block) calculateHash() string {
    data, _ := json.Marshal(b.transactions)
    blockData := b.previousHash + string(data) + b.timestamp.String() + strconv.Itoa(b.proof)
    blockHash := sha256.Sum256([]byte(blockData))
    return fmt.Sprintf("%x", blockHash)
}


// 블록 추가할 때
func (b *Blockchain) addBlock() {
}


// 증명 절차 정의
func doProof() {
    
}

// 외부 API
func doServe() {
    // Echo instance
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Routes
    e.GET("/publish", publish)

    // Start server
    e.Logger.Fatal(e.Start(":1323"))
}

// 트랜젝션 빌드
func buildTransaction(data string) (Transaction, error) {
    lastTransactionId, err := getLastTransactionId()

    if lastTransactionId < 0 {
        return Transaction{}, err
    }

    transaction := Transaction{
        data: data,
        id: lastTransactionId + 1,
        timestamp: time.Now(),
    }

    return transaction, nil
}

// 최근 트랜젝션 ID 구하기
func getLastTransactionId() (int, error) {
    lastTransactionId := -1

    result, err := rdb.Get(ctx, "lastTransactionId").Result()
    if err != nil {
        return lastTransactionId, err
    }

    id, err := strconv.Atoi(result)
    if err != nil {
        lastTransactionId = id
    } else {
        err := rdb.Set(ctx, "lastTransactionId", "0", 0).Err()
        if err != nil {
            lastTransactionId = 0
        }
    }

    return lastTransactionId, err
}

// 최근 블록 해시 구하기
func getLastBlockHash() (string, error) {
    lastBlockHash := ""

    result, err := rdb.Get(ctx, "lastBlockHash").Result()
    if err != nil {
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
        Password: "", // no password set
        DB:       0,  // use default DB
    })
    
    // 최근 트랜젝션 ID 초기화
    getLastTransactionId()
    
    
    
    // doProof()
    // doServe()
}

// 트랜젝션 발행
func publish(c echo.Context) error {
    // 받은 요청 해석
    jsonMap := make(map[string]interface{})
    err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
    if err != nil {
        return err
    }

    // 변수로 재정렬
    hash := jsonMap["hash"].(string)
    filetype := jsonMap["filetype"].(string)
    filename := jsonMap["filename"].(string)

    // 신규 트랜젝션 생성
    transaction, err := buildTransaction(hash + "\t" + filetype + "\t" + filename)
    if err != nil {
        return err
    }

    // 발행할 트랜젝션 본문 만들기
    publishJsonData, err := json.Marshal(transaction)
    if err != nil {
        return err
    }

    // 트랜젝션 발행
    err2 := rdb.Publish(ctx, "transactions_live", publishJsonData).Err()
    if err2 != nil {
        return err2
    }
    
    // 모든 작업이 완료되었으면 오류 없음으로 반환
    return nil
}

// References:
// https://gist.github.com/LordGhostX/bb92b907731ee8ebe465a28c5c431cb4
// https://redis.uptrace.dev/guide/go-redis-pubsub.html
// https://stackoverflow.com/questions/41410655/extract-json-from-golangs-echo-request
// https://stackoverflow.com/questions/27137521/how-to-convert-interface-to-string
