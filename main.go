package main

import (
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
    previousHash string
    transactions []Transaction
    hash string
    timestamp time.Time
    lastTransactionId int
}

type Transaction struct {
    data string
    id int
    timestamp time.Time
}

func (b Block) calculateHash() {
    data, _ := json.Marshal(b.transactions)
    blockData := b.previousHash + string(data) + b.timestamp.String()
    blockHash := sha256.Sum256([]byte(blockData))
    b.hash = fmt.Sprintf("%x", blockHash)
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

func receiveTransactions() {
    pubsub := rdb.Subscribe(ctx, "transaction_live")

    // 메시지 계속 수신
    for {
        // 최근 메시지 수신
        msg, err := pubsub.ReceiveMessage(ctx)
        if err != nil {
            return
        }

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
    _storedTransactions := make([]Transaction, len(storedTransactions))
    copy(storedTransactions, _storedTransactions)
    storedTransactions = make([]Transaction, 0)

    // 복사된 트랜젝션 확인
    for _, transaction := range _storedTransactions {
        if lastTransactionId <= transaction.id {
            transactions = append(transactions, transaction)
        }
    }

    // 블록 생성
    block := Block{
        previousHash: lastBlockHash,
        transactions: transactions,
        timestamp: time.Now(),
        lastTransactionId: lastTransactionId,
    }

    // 블록 해시 계산
    block.calculateHash()

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

    transactionId, err := strconv.Atoi(result)
    if err != nil {
        lastTransactionId = transactionId
    } else {
        err := rdb.Set(ctx, "lastTransactionId", strconv.Itoa(0), 0).Err()
        if err != nil {
            lastTransactionId = 0
        }
    }

    return lastTransactionId, err
}

// 현재 난이도 구하기
func getLastDifficulty() (int, error) {
    lastDifficulty := -1

    result, err := rdb.Get(ctx, "lastDifficulty").Result()
    if err != nil {
        return lastDifficulty, err
    }

    difficulty, err := strconv.Atoi(result)
    if err != nil {
        lastDifficulty = difficulty
    } else {
        err := rdb.Set(ctx, "lastDifficulty", strconv.Itoa(2), 0).Err()
        if err != nil {
            lastDifficulty = 2
        }
    }

    return lastDifficulty, err
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
    publishedJsonData, err := json.Marshal(transaction)
    if err != nil {
        return err
    }

    // 트랜젝션 발행
    err2 := rdb.Publish(ctx, "transactions_live", publishedJsonData).Err()
    if err2 != nil {
        return err2
    }

    // 모든 작업이 완료되었으면 오류 없음으로 반환
    response := map[string]interface{}{
        "success": true,
    }
    return c.JSON(http.StatusOK, response)
}

// References:
//     https://gist.github.com/LordGhostX/bb92b907731ee8ebe465a28c5c431cb4
//     https://redis.uptrace.dev/guide/go-redis-pubsub.html
//     https://stackoverflow.com/questions/41410655/extract-json-from-golangs-echo-request
//     https://stackoverflow.com/questions/27137521/how-to-convert-interface-to-string
