package main

import (
    "context"
    "fmt"
    "time"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "net/http"
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
    data, _ := json.Marshal(b.data)
    blockData := b.previousHash + string(data) + b.timestamp.String() + strconv.Itoa(b.proof)
    blockHash := sha256.Sum256([]byte(blockData))
    return fmt.Sprintf("%x", blockHash)
}


// 블록 추가할 때
func (b *Blockchain) addBlock() string {
    
}

// 블록이 분기되었을 때 머지(병합) 방법 정의
func (b *Blockchain) margeBlock(c1 Block, c2 Block) Block {
    
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
func buildTransaction(data string) Transaction {
	lastTransactionId := getLastTransactionId()

	if lastTransactionId < 0 {
		return nil
	}

	transaction := Transaction{
	    data: data,
		id: lastTransactionId + 1,
		timestamp: time.Now()
	}
	
	return transaction
}

// 최근 트랜젝션 ID 구하기
func getLastTransactionId() int {
	lastTransactionId := -1

    result, err := rdb.Get(ctx, "lastTransactionId").Result()
    if err != nil {
        return lastTransactionId
    }
	lastTransactionId = strconv.Atoi(result)

	return lastTransactionId
}

// 시작점(Entrypoint)
func main() {
    // Connect to Redis server
    rdb = redis.NewClient(&redis.Options{
        Addr:     "154.12.242.48:60713",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
	
	// doProof()
	// doServe()
}

// 트랜젝션 발행
func publish(c echo.Context) error {
	// 받은 요청 해석
	jsonMap := make(map[string]interface{})
	e1 := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if e1 != nil {
		return e1
	}

	// 변수로 재정렬
	hash := jsonMap["hash"]
	filetype := jsonMap["filetype"]
	filename := jsonMap["filename"]

	// 신규 트랜젝션 생성
	newTransaction := buildTransaction(hash + "\t" + filetype + "\t" + filename)

	// 발행할 트랜젝션 본문 만들기
	publishData, _ := json.Marshal(newTransaction)
	
	// 트랜젝션 발행
	e2 := rdb.Publish(ctx, "transaction_live", publishData).Err()
	if e2 != nil {
	    return e2
	}
}



// References:
// https://gist.github.com/LordGhostX/bb92b907731ee8ebe465a28c5c431cb4
// https://redis.uptrace.dev/guide/go-redis-pubsub.html
// https://stackoverflow.com/questions/41410655/extract-json-from-golangs-echo-request